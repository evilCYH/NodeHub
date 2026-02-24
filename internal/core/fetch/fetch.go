package fetch

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/evilCYH/NodeHub/internal/core/mihomo"
	"github.com/evilCYH/NodeHub/internal/core/node"
	"github.com/evilCYH/NodeHub/internal/core/subconv"
	"github.com/evilCYH/NodeHub/internal/database/op"
	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	subModel "github.com/evilCYH/NodeHub/internal/models/sub"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"gopkg.in/yaml.v3"
)

var testingDone = make(map[uint16]<-chan struct{})
var testingDoneMu sync.RWMutex

func GetTestingDone(subID uint16) (<-chan struct{}, bool) {
	testingDoneMu.RLock()
	defer testingDoneMu.RUnlock()
	done, ok := testingDone[subID]
	return done, ok
}

func DeleteTestingDone(subID uint16) {
	testingDoneMu.Lock()
	defer testingDoneMu.Unlock()
	delete(testingDone, subID)
}

func Do(ctx context.Context, subID uint16, config string) subModel.Result {
	startTime := time.Now()
	retry := 0
	var lastErr error
	if ctx == nil {
		ctx = context.Background()
	}

	runLog := subModel.RunLog{
		SubID:     subID,
		Status:    "running",
		CreatedAt: time.Now(),
	}
	runLogCtx, runLogCancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := op.CreateSubRun(runLogCtx, &runLog); err != nil {
		log.Warnf("failed to create sub run: %v", err)
		runLog.ID = 0
	}
	runLogCancel()
	addRunEvent := func(step, level, message string) {
		if runLog.ID == 0 {
			return
		}
		eventCtx, eventCancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := op.CreateSubRunEvent(eventCtx, &subModel.RunEvent{
			RunID:   runLog.ID,
			SubID:   subID,
			Step:    step,
			Level:   level,
			Message: message,
		})
		eventCancel()
		if err != nil {
			log.Warnf("failed to create sub run event: %v", err)
		}
	}
	defer func() {
		if runLog.ID == 0 {
			return
		}
		if runLog.Status == "" || runLog.Status == "running" {
			runLog.Status = "error"
			runLog.Message = "unexpected termination"
		}

		runLog.DurationMs = uint32(time.Since(startTime).Milliseconds())
		updateCtx, updateCancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := op.UpdateSubRun(updateCtx, &runLog); err != nil {
			log.Warnf("failed to update sub run: %v", err)
		}
		updateCancel()
	}()

	var subConfig subModel.Config
	if err := json.Unmarshal([]byte(config), &subConfig); err != nil {
		addRunEvent("config", "error", fmt.Sprintf("invalid config: %v", err))
		log.Warnf("fetch task %d failed: %v", subID, err)
		runLog.Status = "error"
		runLog.Message = err.Error()
		return createFailureResult(err.Error(), startTime)
	}

	log.Debugf("fetch task %d started", subID)

	addRunEvent("fetch", "info", "start fetch subscription")
	client := mihomo.Default(subConfig.Proxy)
	if client == nil {
		addRunEvent("fetch", "error", "proxy config error")
		log.Warnf("fetch task %d failed: proxy config error", subID)
		runLog.Status = "error"
		runLog.Message = "proxy config error"
		return createFailureResult("proxy config error", startTime)
	}
	defer client.Release()
	for retry < 3 {
		time.Sleep(time.Duration(retry) * time.Second)
		retry++
		timeoutSec := subConfig.Timeout
		if timeoutSec <= 0 {
			timeoutSec = 10
		}
		client.Timeout = time.Duration(timeoutSec) * time.Second

		attemptCtx, attemptCancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
		req, err := http.NewRequestWithContext(attemptCtx, "GET", subConfig.Url, nil)
		if err != nil {
			attemptCancel()
			addRunEvent("fetch", "error", fmt.Sprintf("create request failed: %v", err))
			log.Warnf("fetch task %d failed: %v", subID, err)
			lastErr = err
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			attemptCancel()
			addRunEvent("fetch", "error", fmt.Sprintf("request failed: %v", err))
			log.Warnf("fetch task %d failed: %v", subID, err)
			lastErr = err
			continue
		}

		content, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		attemptCancel()
		if err != nil {
			addRunEvent("fetch", "error", fmt.Sprintf("read response failed: %v", err))
			log.Warnf("fetch task %d failed: %v", subID, err)
			lastErr = err
			continue
		}
		addRunEvent("convert", "info", "start subconv convert")
		subconvCtx, subconvCancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
		contentStr, err := subconv.ConvertData(subconvCtx, string(content), "mihomo")
		subconvCancel()
		if err != nil {
			addRunEvent("convert", "error", fmt.Sprintf("convert failed: %v", err))
			log.Errorf("fetch task %d failed: %v", subID, err)
			runLog.Status = "error"
			runLog.Message = err.Error()
			return createFailureResult(err.Error(), startTime)
		}
		content = []byte(contentStr)

		globalProtocolFilterEnable := op.GetSettingBool(setting.NODE_PROTOCOL_FILTER_ENABLE)
		globalProtocolFilterMode := op.GetSettingBool(setting.NODE_PROTOCOL_FILTER_MODE)
		globalProtocolFilter := strings.Split(op.GetSettingStr(setting.NODE_PROTOCOL_FILTER), ",")

		var nodes []nodeModel.Base
		var unique nodeModel.UniqueKey
		lines := bytes.Split(content, []byte("\n"))
		lines = lines[1:]
		rawCount := uint32(0)
		addRunEvent("parse", "info", "start parse nodes")
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			trimmed := bytes.TrimLeft(line, " \t")
			if len(trimmed) == 0 || trimmed[0] != '-' {
				continue
			}
			trimmed = bytes.TrimSpace(trimmed[1:])
			if len(trimmed) == 0 {
				continue
			}
			line = trimmed
			if err := yaml.Unmarshal(line, &unique); err != nil {
				continue
			}
			rawCount++
			if subConfig.ProtocolFilterEnable {
				if subConfig.ProtocolFilterMode {
					if !slices.Contains(subConfig.ProtocolFilter, unique.Type) {
						continue
					}
				} else {
					if slices.Contains(subConfig.ProtocolFilter, unique.Type) {
						continue
					}
				}
			} else {
				if globalProtocolFilterEnable {
					if globalProtocolFilterMode {
						if !slices.Contains(globalProtocolFilter, unique.Type) {
							log.Debugf("全局协议过滤启用,协议包含模式 丢弃协议: %v", unique.Type)
							continue
						}
					} else {
						if slices.Contains(globalProtocolFilter, unique.Type) {
							log.Debugf("全局协议过滤启用,协议排除模式 丢弃协议: %v", unique.Type)
							continue
						}
					}
				}
			}
			// 修复订阅转换服务可能返回的错误格式
			fixedLine := fixProxyConfigIfNeeded(line)
			nodes = append(nodes, nodeModel.Base{
				Raw:       fixedLine,
				SubId:     subID,
				UniqueKey: unique.Gen(),
			})
		}

		count := len(nodes)

		done, processed := node.Add(subID, nodes, runLog.ID)
		if done != nil && processed > 0 {
			testingDoneMu.Lock()
			testingDone[subID] = done
			testingDoneMu.Unlock()
		}
		addRunEvent("node_add", "info", fmt.Sprintf("raw=%d accepted=%d", rawCount, count))

		log.Infof("fetch task %d completed, raw node count: %d, accepted: %d, duration: %dms",
			subID, rawCount, count, time.Since(startTime).Milliseconds())

		runLog.Status = "success"
		runLog.Message = "sub updated successfully"
		runLog.RawCount = rawCount
		runLog.Accepted = uint32(count)
		return createSuccessResult(rawCount, startTime, count == 0)
	}
	if lastErr != nil {
		log.Errorf("fetch task %d failed after %d retries: %v", subID, retry, lastErr)
	} else {
		log.Errorf("fetch task %d failed after %d retries: all tries failed", subID, retry)
	}
	addRunEvent("fetch", "error", "fetch task failed after retries")
	runLog.Status = "error"
	runLog.Message = "fetch task failed"
	return createFailureResult("fetch task failed", startTime)
}

// fixProxyConfigIfNeeded 修复订阅转换服务返回的错误代理配置格式
// 某些订阅转换服务会将 http 代理的完整地址 (username:password@host:port) 进行 base64 编码后放入 server 字段
// 这会导致 Mihomo 无法正确解析，本函数检测并修复这种格式
func fixProxyConfigIfNeeded(raw []byte) []byte {
	var proxy map[string]any
	if err := yaml.Unmarshal(raw, &proxy); err != nil {
		return raw
	}

	// 检查 server 字段是否为 base64 编码的完整代理地址
	server, ok := proxy["server"].(string)
	if !ok || server == "" {
		return raw
	}

	// 尝试 base64 解码
	decoded, err := base64.StdEncoding.DecodeString(server)
	if err != nil {
		// 不是 base64，无需修复
		return raw
	}

	decodedStr := string(decoded)

	// 检查解码后是否包含 @ 符号（username:password@host:port 格式）
	if !strings.Contains(decodedStr, "@") {
		return raw
	}

	// 解析 username:password@host:port 格式
	// 首先分割 @ 符号
	atIndex := strings.LastIndex(decodedStr, "@")
	if atIndex == -1 {
		return raw
	}

	credsPart := decodedStr[:atIndex]
	hostPortPart := decodedStr[atIndex+1:]

	// 解析 username:password
	colonIndex := strings.Index(credsPart, ":")
	if colonIndex == -1 {
		return raw
	}

	username := credsPart[:colonIndex]
	password := credsPart[colonIndex+1:]

	// 解析 host:port
	// IPv6 地址处理（包含 [ ]）
	var host, portStr string
	if strings.HasPrefix(hostPortPart, "[") {
		// IPv6 地址
		bracketEnd := strings.LastIndex(hostPortPart, "]")
		if bracketEnd == -1 {
			return raw
		}
		host = hostPortPart[1:bracketEnd]
		if len(hostPortPart) > bracketEnd+1 && hostPortPart[bracketEnd+1] == ':' {
			portStr = hostPortPart[bracketEnd+2:]
		}
	} else {
		// IPv4 地址或域名
		lastColon := strings.LastIndex(hostPortPart, ":")
		if lastColon == -1 {
			host = hostPortPart
			portStr = ""
		} else {
			host = hostPortPart[:lastColon]
			portStr = hostPortPart[lastColon+1:]
		}
	}

	// 验证并解析端口
	var port int
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			// 端口无效，使用默认端口
			port = 443
		} else {
			port = p
		}
	} else {
		port = 443
	}

	// 获取原始代理类型，如果是 https 则改为 http
	proxyType, _ := proxy["type"].(string)
	if proxyType == "https" {
		proxy["type"] = "http"
	}

	// 重建配置
	proxy["server"] = host
	proxy["port"] = port
	proxy["username"] = username
	proxy["password"] = password
	proxy["tls"] = true

	// 删除错误的 sni 字段（如果存在且也是 base64 编码）
	if sni, ok := proxy["sni"].(string); ok {
		if _, err := base64.StdEncoding.DecodeString(sni); err == nil {
			delete(proxy, "sni")
		}
	}

	// 序列化回 YAML
	fixed, err := yaml.Marshal(proxy)
	if err != nil {
		log.Warnf("fix proxy config failed: %v", err)
		return raw
	}

	log.Debugf("fixed proxy config: %s -> %s", proxy["name"], host)
	return fixed
}

func createFailureResult(msg string, startTime time.Time) subModel.Result {
	return subModel.Result{
		Success:    0,
		Fail:       1,
		Msg:        msg,
		LastStatus: "error",
		LastRun:    time.Now(),
		Duration:   uint32(time.Since(startTime).Milliseconds()),
	}
}

func createSuccessResult(count uint32, startTime time.Time, nodeNull bool) subModel.Result {
	nodeNullCount := uint32(0)
	if nodeNull {
		nodeNullCount = 1
	}
	return subModel.Result{
		Success:       1,
		Fail:          0,
		NodeNullCount: nodeNullCount,
		Msg:           "sub updated successfully",
		LastStatus:    "success",
		RawCount:      count,
		LastRun:       time.Now(),
		Duration:      uint32(time.Since(startTime).Milliseconds()),
	}
}
