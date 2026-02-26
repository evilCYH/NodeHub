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

const (
	subRequestTimeout = 4 * time.Second
)

var fetchRoundUserAgents = [][]string{
	{"Clash", "clash-verge/v2.4.0", "v2rayNG/1.8.12"},
	{"Clash"},
	{"Clash"},
}

type SubInfo struct {
	Upload   int64
	Download int64
	Total    int64
	Expire   int64
	hasUpload bool
	hasDown   bool
	hasTotal  bool
	hasExpire bool
}

func (s *SubInfo) HasUpload() bool {
	return s != nil && s.hasUpload
}

func (s *SubInfo) HasDownload() bool {
	return s != nil && s.hasDown
}

func (s *SubInfo) HasTotal() bool {
	return s != nil && s.hasTotal
}

func (s *SubInfo) HasExpire() bool {
	return s != nil && s.hasExpire
}

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

	content, subInfo, err := fetchSubscriptionWithFallback(ctx, client, subConfig.Url, addRunEvent)
	if err != nil {
		addRunEvent("fetch", "error", fmt.Sprintf("fetch failed: %v", err))
		log.Errorf("fetch task %d failed: %v", subID, err)
		runLog.Status = "error"
		runLog.Message = err.Error()
		return createFailureResult("fetch task failed", startTime)
	}

	addRunEvent("convert", "info", "start subconv convert")
	convertTimeoutSec := subConfig.Timeout
	if convertTimeoutSec <= 0 {
		convertTimeoutSec = 10
	}
	subconvCtx, subconvCancel := context.WithTimeout(ctx, time.Duration(convertTimeoutSec)*time.Second)
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
	for idx, line := range lines {
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
			log.Debugf("node parse unmarshal failed line=%d err=%v raw=%s", idx+2, err, string(line))
			continue
		}
		log.Debugf(
			"node parse unmarshal line=%d raw=%s unique={server:%q servername:%q port:%q type:%q uuid:%q username:%q password:%q}",
			idx+2,
			string(line),
			unique.Server,
			unique.Servername,
			unique.Port,
			unique.Type,
			unique.Uuid,
			unique.Username,
			unique.Password,
		)
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

	if subInfo != nil {
		updateInfoCtx, updateInfoCancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := updateSubInfo(updateInfoCtx, subID, subInfo); err != nil {
			addRunEvent("sub_info", "warn", fmt.Sprintf("update sub info failed: %v", err))
			log.Warnf("failed to update sub info for sub %d: %v", subID, err)
		} else {
			addRunEvent("sub_info", "info", "subscription userinfo updated")
		}
		updateInfoCancel()
	}

	log.Infof("fetch task %d completed, raw node count: %d, accepted: %d, duration: %dms",
		subID, rawCount, count, time.Since(startTime).Milliseconds())

	runLog.Status = "success"
	runLog.Message = "sub updated successfully"
	runLog.RawCount = rawCount
	runLog.Accepted = uint32(count)
	return createSuccessResult(rawCount, startTime, count == 0)
}

func fetchSubscriptionWithFallback(
	ctx context.Context,
	client *mihomo.HC,
	url string,
	addRunEvent func(step, level, message string),
) ([]byte, *SubInfo, error) {
	var content []byte
	var info *SubInfo
	var lastErr error

	for roundIdx, userAgents := range fetchRoundUserAgents {
		if roundIdx > 0 {
			time.Sleep(time.Duration(roundIdx) * time.Second)
		}
		for _, userAgent := range userAgents {
			addRunEvent("fetch", "info", fmt.Sprintf("try fetch with ua=%s round=%d", userAgent, roundIdx+1))
			currentContent, currentInfo, err := fetchOnce(ctx, client, url, userAgent)
			if err != nil {
				addRunEvent("fetch", "warn", fmt.Sprintf("fetch failed ua=%s round=%d: %v", userAgent, roundIdx+1, err))
				lastErr = err
				continue
			}

			if len(content) == 0 {
				content = currentContent
			}
			if info == nil && currentInfo != nil {
				info = currentInfo
				return content, info, nil
			}
		}

		if len(content) > 0 {
			return content, info, nil
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("all attempts failed")
	}
	return nil, nil, lastErr
}

func fetchOnce(ctx context.Context, client *mihomo.HC, url, userAgent string) ([]byte, *SubInfo, error) {
	client.Timeout = subRequestTimeout
	reqCtx, reqCancel := context.WithTimeout(ctx, subRequestTimeout)
	defer reqCancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response failed: %w", err)
	}

	info := parseSubscriptionUserInfo(resp.Header.Get("subscription-userinfo"))
	return content, info, nil
}

func parseSubscriptionUserInfo(header string) *SubInfo {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil
	}

	info := &SubInfo{}
	parsedCount := 0
	parts := strings.Split(header, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		value := strings.TrimSpace(kv[1])
		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}
		switch key {
		case "upload":
			info.Upload = num
			info.hasUpload = true
			parsedCount++
		case "download":
			info.Download = num
			info.hasDown = true
			parsedCount++
		case "total":
			info.Total = num
			info.hasTotal = true
			parsedCount++
		case "expire":
			info.Expire = num
			info.hasExpire = true
			parsedCount++
		}
	}
	if parsedCount == 0 {
		return nil
	}
	return info
}

func updateSubInfo(ctx context.Context, subID uint16, info *SubInfo) error {
	oldSub, err := op.GetSubByID(ctx, subID)
	if err != nil {
		return err
	}

	upload := oldSub.Upload
	download := oldSub.Download
	total := oldSub.Total
	expire := oldSub.Expire
	if info.HasUpload() {
		upload = info.Upload
	}
	if info.HasDownload() {
		download = info.Download
	}
	if info.HasTotal() {
		total = info.Total
	}
	if info.HasExpire() {
		expire = info.Expire
	}
	now := time.Now()
	return op.UpdateSubInfo(ctx, subID, upload, download, total, expire, &now)
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
