package node

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"os"
	"path"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bestruirui/bestsub/internal/config"
	"github.com/bestruirui/bestsub/internal/core/mihomo"
	"github.com/bestruirui/bestsub/internal/core/task"
	"github.com/bestruirui/bestsub/internal/database/op"
	nodeModel "github.com/bestruirui/bestsub/internal/models/node"
	"github.com/bestruirui/bestsub/internal/models/setting"
	"github.com/bestruirui/bestsub/internal/utils/generic"
	"github.com/bestruirui/bestsub/internal/utils/log"
)

func InitNodePool(size int) {
	pool = make([]nodeModel.Data, 0, size)
	nodeExist = NewExist(size)
	nodeProcess = NewExist(size)
	sessionFile := config.Base().Session.NodePath
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return
	}

	file, err := os.Open(sessionFile)
	if err != nil {
		return
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&pool); err != nil {
		log.Warnf("restore node pool failed: %v", err)
		os.Remove(sessionFile)
		return
	}
	for i := range pool {
		nodeExist.Add(pool[i].Base.UniqueKey)
		registry.Upsert(nodeModel.Record{
			Base:       pool[i].Base,
			Info:       pool[i].Info,
			InitStatus: nodeModel.InitPassed,
		})
		// 确保 Queue 正确初始化
		if pool[i].Info != nil {
			if pool[i].Info.Delay.Data == nil {
				pool[i].Info.Delay = *generic.NewQueue[uint16](5)
			}
			if pool[i].Info.SpeedUp.Data == nil {
				pool[i].Info.SpeedUp = *generic.NewQueue[uint32](5)
			}
			if pool[i].Info.SpeedDown.Data == nil {
				pool[i].Info.SpeedDown = *generic.NewQueue[uint32](5)
			}
		}
	}
}

func CloseNodePool() error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(pool); err != nil {
		log.Warnf("save node pool failed: %v", err)
		return err
	}
	filePath := config.Base().Session.NodePath

	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	if os.WriteFile(filePath, buf.Bytes(), 0600) != nil {
		log.Warnf("node pool save failed")
	}
	log.Debugf("node pool saved")

	return nil
}

type nameNode struct {
	Name string
}

func getNodeType(raw map[string]any) string {
	if nodeType, ok := raw["type"].(string); ok && nodeType != "" {
		return nodeType
	}
	return "unknown"
}

// getNodeName 从 raw 配置中获取节点名称
func getNodeName(raw map[string]any) string {
	if name, ok := raw["name"].(string); ok && name != "" {
		return name
	}
	if server, ok := raw["server"].(string); ok && server != "" {
		return server
	}
	return "unknown"
}

// isTimeoutError 判断是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "i/o timeout")
}

// isNetworkError 判断是否为网络错误（非超时）
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "connection reset")
}

func classifyTestError(err error) string {
	if err == nil {
		return "unknown"
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "certificate has expired") || strings.Contains(msg, "not yet valid") {
		return "tls_cert_expired"
	}
	if strings.Contains(msg, "failed to verify certificate") || strings.Contains(msg, "x509") {
		return "tls_verify_failed"
	}
	if strings.Contains(msg, "unexpected eof") {
		return "unexpected_eof"
	}
	if msg == "eof" || strings.HasSuffix(msg, ": eof") || strings.Contains(msg, " eof") {
		return "eof"
	}
	if isTimeoutError(err) {
		return "timeout"
	}
	if strings.Contains(msg, "no such host") {
		return "dns_error"
	}
	if isNetworkError(err) {
		return "network_error"
	}
	return "request_error"
}

type addStats struct {
	subID      uint16
	runID      uint64
	start      time.Time
	rawCount   uint16
	candidate  uint32
	duplicate  uint32
	invalid    uint32
	testFailed uint32
	accepted   uint32
	validNodes []nodeModel.Data
	validMu    sync.Mutex
	detailMu   sync.Mutex
	details    []string
	// 节点级详细日志
	nodeLogs []nodeModel.NodeLog
	logMu    sync.Mutex
}

func newAddStats(subID, rawCount uint16, runID uint64) *addStats {
	return &addStats{
		subID:    subID,
		runID:    runID,
		start:    time.Now(),
		rawCount: rawCount,
	}
}

func (s *addStats) ResetFailedNodes(subID uint16) {
	failedNodeMu.Lock()
	delete(failedNodeStore, subID)
	failedNodeMu.Unlock()
}

func (s *addStats) AddFailedNode(node nodeModel.FailedNode) {
	if node.Name == "" {
		node.Name = "unknown"
	}
	if node.Type == "" {
		node.Type = "unknown"
	}
	failedNodeMu.Lock()
	defer failedNodeMu.Unlock()

	list := failedNodeStore[node.SubID]
	if len(list) >= 500 {
		return
	}
	for _, existing := range list {
		if existing.UniqueKey == node.UniqueKey && existing.Reason == node.Reason {
			return
		}
	}
	failedNodeStore[node.SubID] = append(list, node)
}

func (s *addStats) IncCandidate() { atomic.AddUint32(&s.candidate, 1) }
func (s *addStats) IncDuplicate() { atomic.AddUint32(&s.duplicate, 1) }
func (s *addStats) IncInvalid()   { atomic.AddUint32(&s.invalid, 1) }
func (s *addStats) IncFailed()    { atomic.AddUint32(&s.testFailed, 1) }

func (s *addStats) AddValid(node nodeModel.Data) {
	s.validMu.Lock()
	s.validNodes = append(s.validNodes, node)
	s.validMu.Unlock()
	atomic.AddUint32(&s.accepted, 1)
}

func (s *addStats) AddDetail(reason string) {
	s.detailMu.Lock()
	if len(s.details) < 50 {
		s.details = append(s.details, reason)
	}
	s.detailMu.Unlock()
}

// AddNodeLog 添加节点级详细日志
func (s *addStats) AddNodeLog(level string, nodeKey uint64, nodeName, message string) {
	s.logMu.Lock()
	defer s.logMu.Unlock()

	// 限制存储数量，避免内存爆炸
	if len(s.nodeLogs) >= 500 {
		return
	}

	s.nodeLogs = append(s.nodeLogs, nodeModel.NodeLog{
		SubID:     s.subID,
		NodeKey:   nodeKey,
		NodeName:  nodeName,
		Level:     level,
		Source:    nodeModel.LogSourceInit,
		RunID:     s.runID,
		Message:   message,
		CreatedAt: time.Now(),
	})
}

func (s *addStats) Finalize() {
	merged := 0
	if len(s.validNodes) > 0 {
		merged = mergeNodesToPool(s.validNodes)
		RebuildPoolFromRegistry(op.GetSettingInt(setting.NODE_POOL_SIZE))
		RefreshInfo()
	}
	log.Infof("Receipt successful, %d new nodes added", merged)

	candidate := clampToUint16(atomic.LoadUint32(&s.candidate))
	duplicate := clampToUint16(atomic.LoadUint32(&s.duplicate))
	invalid := clampToUint16(atomic.LoadUint32(&s.invalid))
	testFailed := clampToUint16(atomic.LoadUint32(&s.testFailed))
	accepted := clampToUint16(atomic.LoadUint32(&s.accepted))
	mergedU16 := clampToUint16(uint32(merged))
	dropped := uint16(0)
	if accepted > mergedU16 {
		dropped = accepted - mergedU16
	}

	updateLog := nodeModel.UpdateLog{
		SubID:      s.subID,
		CreatedAt:  time.Now(),
		DurationMs: uint16(time.Since(s.start).Milliseconds()),
		RawCount:   s.rawCount,
		Candidate:  candidate,
		Duplicate:  duplicate,
		Invalid:    invalid,
		TestFailed: testFailed,
		Accepted:   accepted,
		Merged:     mergedU16,
		Dropped:    dropped,
		Details:    append([]string(nil), s.details...),
	}
	if err := op.CreateNodeUpdateLog(context.Background(), &updateLog); err != nil {
		log.Warnf("failed to save node update log: %v", err)
	}

	// 保存节点级详细日志
	if len(s.nodeLogs) > 0 {
		saveNodeTestLogs(s.subID, s.nodeLogs)
	}
}

func clampToUint16(value uint32) uint16 {
	if value > uint32(^uint16(0)) {
		return ^uint16(0)
	}
	return uint16(value)
}

func Add(subID uint16, nodes []nodeModel.Base, runID uint64) (<-chan struct{}, int) {
	var nodesToProcess []nodeModel.Base
	if len(nodes) == 0 {
		return nil, 0
	}
	if subID == 0 {
		subID = nodes[0].SubId
	}
	if subID == 0 {
		return nil, 0
	}
	stats := newAddStats(subID, uint16(len(nodes)), runID)
	stats.ResetFailedNodes(subID)

	for _, n := range nodes {
		// 注册节点到主表（Registry）
		rawCopy := append([]byte(nil), n.Raw...)
		n.Raw = rawCopy
		registry.Upsert(nodeModel.Record{
			Base:       n,
			Info:       nil,
			InitStatus: nodeModel.InitUnknown,
		})
		var nameNode nameNode
		if err := yaml.Unmarshal(n.Raw, &nameNode); err != nil {
			log.Warnf("yaml.Unmarshal failed: %v", err)
			stats.IncInvalid()
			stats.AddDetail("yaml_unmarshal_failed")
			UpdateRegistryInitStatus(subID, n.UniqueKey, nodeModel.InitFailed, "yaml_unmarshal_failed")
			stats.AddFailedNode(nodeModel.FailedNode{
				SubID:     subID,
				UniqueKey: n.UniqueKey,
				Name:      "unknown",
				Type:      "unknown",
				Reason:    "yaml_unmarshal_failed",
			})
			continue
		}

		// 检查节点是否正在处理中（避免并发重复处理）
		if nodeProcess.Exist(n.UniqueKey) {
			log.Debugf("node already in process: %s", nameNode.Name)
			continue
		}
		// 检查节点是否已存在于池中（入库初测只处理新节点）
		if nodeExist.Exist(n.UniqueKey) {
			log.Debugf("node already exist: %s", nameNode.Name)
			stats.IncDuplicate()
			continue
		}
		log.Debugf("add process node: %s", nameNode.Name)
		nodeProcess.Add(n.UniqueKey)
		stats.IncCandidate()
		nodesToProcess = append(nodesToProcess, n)
	}

	log.Debugf("add %d nodes to process", len(nodesToProcess))
	if len(nodesToProcess) == 0 {
		stats.Finalize()
		return 0
	}

	var wg sync.WaitGroup
	for _, node := range nodesToProcess {
		n := node // capture loop variable
		wg.Add(1)
		task.Submit(func() {
			defer wg.Done()
			defer nodeProcess.Remove(n.UniqueKey)
			var raw map[string]any
			if err := yaml.Unmarshal(n.Raw, &raw); err != nil {
				log.Warnf("yaml.Unmarshal failed: %v", err)
				stats.IncInvalid()
				stats.AddDetail("yaml_unmarshal_failed")
				UpdateRegistryInitStatus(subID, n.UniqueKey, nodeModel.InitFailed, "yaml_unmarshal_failed")
				stats.AddFailedNode(nodeModel.FailedNode{
					SubID:     subID,
					UniqueKey: n.UniqueKey,
					Name:      "unknown",
					Type:      "unknown",
					Reason:    "yaml_unmarshal_failed",
				})
				return
			}

			// 获取节点名称
			nodeName := getNodeName(raw)

			// 开始测试 - info 级别
			stats.AddNodeLog("info", n.UniqueKey, nodeName, "开始节点初测")

			client := mihomo.Proxy(raw)
			if client == nil {
				stats.AddNodeLog("error", n.UniqueKey, nodeName, "代理解析失败：配置无效或协议不支持")
				stats.IncInvalid()
				stats.AddDetail("proxy_parse_failed")
				UpdateRegistryInitStatus(subID, n.UniqueKey, nodeModel.InitFailed, "proxy_parse_failed")
				stats.AddFailedNode(nodeModel.FailedNode{
					SubID:     subID,
					UniqueKey: n.UniqueKey,
					Name:      nodeName,
					Type:      getNodeType(raw),
					Reason:    "proxy_parse_failed",
				})
				return
			}
			defer client.Release()

			testURL := op.GetSettingStr(setting.NODE_TEST_URL)
			stats.AddNodeLog("info", n.UniqueKey, nodeName, "发送测试请求至 "+testURL)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(op.GetSettingInt(setting.NODE_TEST_TIMEOUT))*time.Second)
			defer cancel()

			// 使用 httptrace 测量首字节时间
			var firstByteTime time.Time
			startTime := time.Now()

			trace := &httptrace.ClientTrace{
				GotFirstResponseByte: func() {
					firstByteTime = time.Now()
				},
			}

			reqCtx := httptrace.WithClientTrace(ctx, trace)
			request, err := http.NewRequestWithContext(reqCtx, "GET", testURL, nil)
			if err != nil {
				stats.AddNodeLog("error", n.UniqueKey, nodeName, "创建请求失败: "+err.Error())
				stats.IncInvalid()
				stats.AddDetail("request_create_failed")
				UpdateRegistryInitStatus(subID, n.UniqueKey, nodeModel.InitFailed, "request_create_failed")
				stats.AddFailedNode(nodeModel.FailedNode{
					SubID:     subID,
					UniqueKey: n.UniqueKey,
					Name:      nodeName,
					Type:      getNodeType(raw),
					Reason:    "request_create_failed",
				})
				return
			}

			response, err := client.Do(request)
			if err != nil {
				// 根据错误类型区分级别
				level := "error"
				errMsg := err.Error()
				if isTimeoutError(err) {
					level = "warn"
					errMsg = "连接超时: " + errMsg
				} else if isNetworkError(err) {
					level = "warn"
					errMsg = "网络错误: " + errMsg
				} else {
					errMsg = "请求失败: " + errMsg
				}
				stats.AddNodeLog(level, n.UniqueKey, nodeName, errMsg)
				stats.IncFailed()
				stats.AddDetail("test_request_failed: " + err.Error())
				UpdateRegistryInitStatus(subID, n.UniqueKey, nodeModel.InitFailed, classifyTestError(err))
				stats.AddFailedNode(nodeModel.FailedNode{
					SubID:     subID,
					UniqueKey: n.UniqueKey,
					Name:      nodeName,
					Type:      getNodeType(raw),
					Reason:    "test_request_failed",
				})
				return
			}
			defer response.Body.Close()

			if response.StatusCode != 204 {
				msg := fmt.Sprintf("状态码不符: 期望 204, 实际 %d", response.StatusCode)
				stats.AddNodeLog("warn", n.UniqueKey, nodeName, msg)
				stats.IncFailed()
				stats.AddDetail("unexpected_status: " + response.Status)
				UpdateRegistryInitStatus(subID, n.UniqueKey, nodeModel.InitFailed, "unexpected_status_"+response.Status)
				stats.AddFailedNode(nodeModel.FailedNode{
					SubID:     subID,
					UniqueKey: n.UniqueKey,
					Name:      nodeName,
					Type:      getNodeType(raw),
					Reason:    "unexpected_status",
				})
				return
			}

			// 确保获取到首字节时间
			if firstByteTime.IsZero() {
				firstByteTime = time.Now()
			}

			delay := firstByteTime.Sub(startTime).Milliseconds()
			stats.AddNodeLog("info", n.UniqueKey, nodeName, fmt.Sprintf("初测通过，延迟: %dms", delay))

			var info nodeModel.Info
			// 正确初始化 Queue，设置容量为 5
			info.Delay = *generic.NewQueue[uint16](5)
			info.SpeedUp = *generic.NewQueue[uint32](5)
			info.SpeedDown = *generic.NewQueue[uint32](5)
			info.Delay.Update(uint16(delay))
			info.SetAliveStatus(nodeModel.Alive, true)
			rawCopy := append([]byte(nil), n.Raw...)
			n.Raw = rawCopy
			registry.Upsert(nodeModel.Record{
				Base:            n,
				Info:            &info,
				InitStatus:      nodeModel.InitPassed,
				LastCheckAt:     time.Now(),
				LastCheckSource: "initial",
			})

			// 新节点，添加到候选列表
			stats.AddValid(nodeModel.Data{
				Base: n,
				Info: &info,
			})
			if rawName, ok := raw["name"].(string); ok {
				log.Debugf("node: %s test end, Delay: %d", rawName, info.Delay.Average())
			}
		})
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		stats.Finalize()
		close(done)
	}()
	return done, len(nodesToProcess)
}

func GetUpdateLog(subID uint16, limit int) nodeModel.UpdateLogResponse {
	logs, err := op.ListNodeUpdateLogs(context.Background(), subID, limit)
	if err != nil {
		log.Warnf("failed to list node update logs: %v", err)
		return nodeModel.UpdateLogResponse{}
	}
	if len(logs) == 0 {
		return nodeModel.UpdateLogResponse{}
	}
	return nodeModel.UpdateLogResponse{
		Latest:  &logs[0],
		History: logs,
	}
}

func ForEach(fn func(node []byte)) {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	for _, node := range pool {
		fn(node.Raw)
	}
}

func GetAll() []nodeModel.Data {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	result := make([]nodeModel.Data, len(pool))
	for i, n := range pool {
		result[i] = copyNodeData(n)
	}
	return result
}

func GetRegistryAll() []nodeModel.Record {
	return registry.GetAll()
}

func GetBySubIdExclude(subId []uint16) []uint16 {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	var result []uint16
	for _, node := range pool {
		if !slices.Contains(subId, node.Base.SubId) {
			result = append(result, node.Base.SubId)
		}
	}
	return result
}

func GetSubIDsFromPool() []uint16 {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	seen := make(map[uint16]struct{})
	result := make([]uint16, 0)
	for _, node := range pool {
		if _, ok := seen[node.Base.SubId]; ok {
			continue
		}
		seen[node.Base.SubId] = struct{}{}
		result = append(result, node.Base.SubId)
	}
	return result
}

func GetBySubId(subId []uint16) *[]nodeModel.Data {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	var result []nodeModel.Data
	for _, node := range pool {
		if slices.Contains(subId, node.Base.SubId) {
			result = append(result, copyNodeData(node))
		}
	}
	return &result
}

func GetRegistryBySubId(subId []uint16) []nodeModel.Record {
	return registry.GetBySubIDs(subId)
}

func GetByFilter(filter nodeModel.Filter) *[]nodeModel.Data {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	var result []nodeModel.Data
	for _, node := range pool {
		if len(filter.SubId) > 0 {
			if filter.SubIdExclude && slices.Contains(filter.SubId, node.Base.SubId) {
				continue
			}
			if !filter.SubIdExclude && !slices.Contains(filter.SubId, node.Base.SubId) {
				continue
			}
		}
		if filter.AliveStatus != 0 && node.Info.AliveStatus&filter.AliveStatus != filter.AliveStatus {
			continue
		}
		if len(filter.Country) > 0 {
			if filter.CountryExclude && slices.Contains(filter.Country, node.Info.Country) {
				continue
			}
			if !filter.CountryExclude && !slices.Contains(filter.Country, node.Info.Country) {
				continue
			}
		}
		if filter.SpeedUpMore != 0 && node.Info.SpeedUp.Average() < filter.SpeedUpMore {
			continue
		}
		if filter.SpeedDownMore != 0 && node.Info.SpeedDown.Average() < filter.SpeedDownMore {
			continue
		}
		if filter.DelayLessThan != 0 && node.Info.Delay.Average() > filter.DelayLessThan {
			continue
		}
		if filter.RiskLessThan != 0 && node.Info.Risk > filter.RiskLessThan {
			continue
		}
		result = append(result, copyNodeData(node))
	}
	return &result
}

func GetRegistryByFilter(filter nodeModel.Filter) []nodeModel.Record {
	items := registry.GetAll()
	result := make([]nodeModel.Record, 0, len(items))
	for _, item := range items {
		if len(filter.SubId) > 0 {
			if filter.SubIdExclude && slices.Contains(filter.SubId, item.Base.SubId) {
				continue
			}
			if !filter.SubIdExclude && !slices.Contains(filter.SubId, item.Base.SubId) {
				continue
			}
		}
		if filter.AliveStatus != 0 {
			if item.Info == nil || item.Info.AliveStatus&filter.AliveStatus != filter.AliveStatus {
				continue
			}
		}
		if len(filter.Country) > 0 {
			if item.Info == nil {
				continue
			}
			if filter.CountryExclude && slices.Contains(filter.Country, item.Info.Country) {
				continue
			}
			if !filter.CountryExclude && !slices.Contains(filter.Country, item.Info.Country) {
				continue
			}
		}
		if filter.SpeedUpMore != 0 {
			if item.Info == nil || item.Info.SpeedUp.Average() < filter.SpeedUpMore {
				continue
			}
		}
		if filter.SpeedDownMore != 0 {
			if item.Info == nil || item.Info.SpeedDown.Average() < filter.SpeedDownMore {
				continue
			}
		}
		if filter.DelayLessThan != 0 {
			if item.Info == nil || item.Info.Delay.Average() > filter.DelayLessThan {
				continue
			}
		}
		if filter.RiskLessThan != 0 {
			if item.Info == nil || item.Info.Risk > filter.RiskLessThan {
				continue
			}
		}
		result = append(result, item)
	}
	return result
}

func GetFailedBySubId(subId []uint16) []nodeModel.FailedNode {
	result := make([]nodeModel.FailedNode, 0)
	failedNodeMu.RLock()
	defer failedNodeMu.RUnlock()
	for _, id := range subId {
		if list, ok := failedNodeStore[id]; ok {
			result = append(result, list...)
		}
	}
	return result
}

func copyNodeData(node nodeModel.Data) nodeModel.Data {
	if node.Info == nil {
		return node
	}
	info := *node.Info
	info.SpeedUp = cloneQueue(info.SpeedUp)
	info.SpeedDown = cloneQueue(info.SpeedDown)
	info.Delay = cloneQueue(info.Delay)
	return nodeModel.Data{
		Base: node.Base,
		Info: &info,
	}
}

func cloneQueue[T generic.Integer](q generic.Queue[T]) generic.Queue[T] {
	// 如果原队列为空（未初始化），创建一个新的队列
	if q.Data == nil {
		return *generic.NewQueue[T](5)
	}
	data := append([]T(nil), q.Data...)
	return generic.Queue[T]{
		Data: data,
		Ptr:  q.Ptr,
		Full: q.Full,
	}
}

// UpdateNodeInPool 更新池中节点的状态
// 如果节点存在，更新其 Info；如果不存在，返回 false
// 如果测试失败（info 为 nil），从池中移除节点
func UpdateNodeInPool(uniqueKey uint64, info *nodeModel.Info) bool {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	for i := range pool {
		if pool[i].Base.UniqueKey == uniqueKey {
			if info == nil {
				// 测试失败，从池中移除
				nodeExist.Remove(uniqueKey)
				pool = append(pool[:i], pool[i+1:]...)
				return true
			}
			// 更新节点信息
			pool[i].Info = info
			return true
		}
	}
	return false
}

func mergeNodesToPool(newNodes []nodeModel.Data) int {
	sort.Slice(newNodes, func(i, j int) bool {
		return newNodes[i].Info.Delay.Average() < newNodes[j].Info.Delay.Average()
	})

	poolMutex.Lock()
	defer poolMutex.Unlock()

	poolLen := len(pool)
	poolCap := cap(pool)

	if poolLen < poolCap {
		remainingCap := poolCap - poolLen
		if len(newNodes) < remainingCap {
			pool = append(pool, newNodes...)
			for _, node := range newNodes {
				nodeExist.Add(node.Base.UniqueKey)
			}
			return len(newNodes)
		} else {
			pool = append(pool, newNodes[:remainingCap]...)
			for _, node := range newNodes[:remainingCap] {
				nodeExist.Add(node.Base.UniqueKey)
			}
			newNodes = newNodes[remainingCap:]
		}
	}

	sort.Slice(pool, func(i, j int) bool {
		return pool[i].Info.Delay.Average() < pool[j].Info.Delay.Average()
	})

	newNodeIndex := 0
	for i := len(pool) - 1; i >= 0 && newNodeIndex < len(newNodes); i-- {
		if newNodes[newNodeIndex].Info.Delay.Average() < pool[i].Info.Delay.Average() {
			log.Debugf("new node delay %dms < old delay %dms,merge", newNodes[newNodeIndex].Info.Delay.Average(), pool[i].Info.Delay.Average())
			nodeExist.Remove(pool[i].Base.UniqueKey)
			pool[i] = newNodes[newNodeIndex]
			nodeExist.Add(newNodes[newNodeIndex].Base.UniqueKey)
			newNodeIndex++
		} else {
			log.Debugf("new node delay %dms > old delay %dms,not merge", newNodes[newNodeIndex].Info.Delay.Average(), pool[i].Info.Delay.Average())
			return newNodeIndex
		}
	}
	return 0
}

func RebuildPoolFromRegistry(maxSize int) int {
	items := registry.GetAll()
	candidates := make([]nodeModel.Data, 0, len(items))
	for _, item := range items {
		if item.InitStatus != nodeModel.InitPassed || item.Info == nil {
			continue
		}
		candidates = append(candidates, nodeModel.Data{Base: item.Base, Info: item.Info})
	}
	if len(candidates) == 0 {
		poolMutex.Lock()
		pool = nil
		poolMutex.Unlock()
		return 0
	}
	// 同一 unique_key 只保留延迟最小的记录
	bestByKey := make(map[uint64]nodeModel.Data)
	for _, node := range candidates {
		existing, ok := bestByKey[node.Base.UniqueKey]
		if !ok || node.Info.Delay.Average() < existing.Info.Delay.Average() {
			bestByKey[node.Base.UniqueKey] = node
		}
	}
	unique := make([]nodeModel.Data, 0, len(bestByKey))
	for _, node := range bestByKey {
		unique = append(unique, copyNodeData(node))
	}
	sort.Slice(unique, func(i, j int) bool {
		return unique[i].Info.Delay.Average() < unique[j].Info.Delay.Average()
	})
	if maxSize > 0 && len(unique) > maxSize {
		unique = unique[:maxSize]
	}
	poolMutex.Lock()
	pool = unique
	poolMutex.Unlock()
	keys := make([]uint64, 0, len(unique))
	for _, node := range unique {
		keys = append(keys, node.Base.UniqueKey)
	}
	nodeExist.Reset(keys)
	return len(unique)
}

func GetSubInfo(subID uint16) nodeModel.SimpleInfo {
	refreshMutex.Lock()
	defer refreshMutex.Unlock()
	return subInfoMap[subID]
}

func GetCountryInfo(country string) nodeModel.SimpleInfo {
	refreshMutex.Lock()
	defer refreshMutex.Unlock()
	return countryInfoMap[country]
}
func DeleteBySubId(subID uint16) {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	end := len(pool) - 1
	for i := 0; i <= end; {
		if pool[i].Base.SubId == subID {
			nodeExist.Remove(pool[i].Base.UniqueKey)
			pool[i] = pool[end]
			end--
		} else {
			i++
		}
	}

	pool = pool[:end+1]
	registry.DeleteBySubID(subID)
}

// saveNodeTestLogs 保存节点测试日志
func saveNodeTestLogs(subID uint16, logs []nodeModel.NodeLog) {
	for i := range logs {
		entry := logs[i]
		entry.SubID = subID
		if entry.Source == "" {
			entry.Source = nodeModel.LogSourceInit
		}
		if err := op.CreateNodeLog(context.Background(), &entry); err != nil {
			log.Warnf("failed to save node log: %v", err)
		}
	}
}
