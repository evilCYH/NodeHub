package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/evilCYH/NodeHub/internal/core/mihomo"
	"github.com/evilCYH/NodeHub/internal/core/node"
	"github.com/evilCYH/NodeHub/internal/core/task"
	"github.com/evilCYH/NodeHub/internal/database/op"
	checkModel "github.com/evilCYH/NodeHub/internal/models/check"
	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	"github.com/evilCYH/NodeHub/internal/modules/register"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

type Alive struct {
	URL         string `json:"url" name:"测试链接" value:"https://www.gstatic.com/generate_204"`
	ExptectCode int    `json:"exptect_code" name:"期望状态码" value:"204"`
	Thread      int    `json:"thread" name:"线程数" value:"30"`
	Timeout     int    `json:"timeout" name:"超时时间" value:"10" desc:"单个节点检测的超时时间(s)"`
}
type Result struct {
	AliveCount uint16 `json:"alive_count" desc:"存活节点数量"`
	DeadCount  uint16 `json:"dead_count" desc:"死亡节点数量"`
	Delay      uint16 `json:"delay" desc:"平均延迟"`
}

func (e *Alive) Init() error {
	return nil
}

func (e *Alive) Run(ctx context.Context, log *log.Logger, subID []uint16) checkModel.Result {
	startTime := time.Now()
	checkID := getCheckID(ctx)
	var nodes []nodeModel.Data
	var aliveCount, deadCount, totalDelay int64
	if len(subID) == 0 {
		nodes = node.GetAll()
	} else {
		nodes = *node.GetBySubId(subID)
	}
	threads := e.Thread
	if threads <= 0 || threads > len(nodes) {
		threads = len(nodes)
	}
	if threads > task.MaxThread() {
		threads = task.MaxThread()
	}
	if threads == 0 || len(nodes) == 0 {
		log.Warnf("alive check task failed, no nodes")
		return checkModel.Result{
			Msg:      "no nodes",
			LastRun:  time.Now(),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}
	sem := make(chan struct{}, threads)
	defer close(sem)

	var wg sync.WaitGroup
	for _, nd := range nodes {
		sem <- struct{}{}
		wg.Add(1)
		n := nd
		if err := task.Submit(func() {
			defer func() {
				<-sem
				wg.Done()
			}()
			if n.Info == nil {
				return
			}
			var raw map[string]any
			if err := yaml.Unmarshal(n.Raw, &raw); err != nil {
				log.Warnf("yaml.Unmarshal failed: %v", err)
				if err := op.CreateNodeLog(ctx, &nodeModel.NodeLog{
					SubID:     n.Base.SubId,
					NodeKey:   n.Base.UniqueKey,
					NodeName:  "unknown",
					Level:     "error",
					Source:    nodeModel.LogSourceCheck,
					CheckID:   checkID,
					Message:   "yaml unmarshal failed",
					CreatedAt: time.Now(),
				}); err != nil {
					log.Warnf("failed to create node log: %v", err)
				}
				return
			}
			nodeName := getNodeName(raw)
			delay, alive := e.detectWithDelay(ctx, raw)
			if alive {
				atomic.AddInt64(&aliveCount, 1)
				n.Info.SetAliveStatus(nodeModel.Alive, true)
				n.Info.Delay.Update(delay)
				atomic.AddInt64(&totalDelay, int64(n.Info.Delay.Average()))
				node.UpdateNodeAliveInPool(n.Base.UniqueKey, true, delay)
				node.UpdateRegistryAlive(n.Base.SubId, n.Base.UniqueKey, true, delay, "alive_task")
				if err := op.CreateNodeLog(ctx, &nodeModel.NodeLog{
					SubID:     n.Base.SubId,
					NodeKey:   n.Base.UniqueKey,
					NodeName:  nodeName,
					Level:     "info",
					Source:    nodeModel.LogSourceCheck,
					CheckID:   checkID,
					Message:   fmt.Sprintf("alive: %dms", delay),
					CreatedAt: time.Now(),
				}); err != nil {
					log.Warnf("failed to create node log: %v", err)
				}
			} else {
				atomic.AddInt64(&deadCount, 1)
				n.Info.SetAliveStatus(nodeModel.Alive, false)
				node.UpdateNodeAliveInPool(n.Base.UniqueKey, false, 0)
				node.UpdateRegistryAlive(n.Base.SubId, n.Base.UniqueKey, false, 0, "alive_task")
				if err := op.CreateNodeLog(ctx, &nodeModel.NodeLog{
					SubID:     n.Base.SubId,
					NodeKey:   n.Base.UniqueKey,
					NodeName:  nodeName,
					Level:     "warn",
					Source:    nodeModel.LogSourceCheck,
					CheckID:   checkID,
					Message:   "dead",
					CreatedAt: time.Now(),
				}); err != nil {
					log.Warnf("failed to create node log: %v", err)
				}
			}
		}); err != nil {
			<-sem
			wg.Done()
			log.Warnf("alive check task submit failed: %v", err)
		}
	}
	wg.Wait()
	avgDelay := int64(0)
	if aliveCount > 0 {
		avgDelay = totalDelay / aliveCount
	}
	log.Debugf("alive check task end, alive: %d, dead: %d, average delay: %dms", aliveCount, deadCount, avgDelay)
	return checkModel.Result{
		Msg:      fmt.Sprintf("success, alive: %d, dead: %d, average delay: %dms", aliveCount, deadCount, avgDelay),
		LastRun:  time.Now(),
		Duration: time.Since(startTime).Milliseconds(),
		Extra: map[string]any{
			"alive": aliveCount,
			"dead":  deadCount,
			"delay": avgDelay,
		},
	}
}

func (e *Alive) detect(ctx context.Context, raw map[string]any) bool {
	client := mihomo.Proxy(raw)
	if client == nil {
		return false
	}
	client.Timeout = time.Duration(e.Timeout) * time.Second
	defer client.Release()
	request, err := http.NewRequestWithContext(ctx, "GET", e.URL, nil)
	if err != nil {
		return false
	}
	response, err := client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == e.ExptectCode
}

// detectWithDelay 使用 httptrace 测量首字节时间(TTFB)，返回延迟和是否存活
func (e *Alive) detectWithDelay(ctx context.Context, raw map[string]any) (uint16, bool) {
	client := mihomo.Proxy(raw)
	if client == nil {
		return 0, false
	}
	client.Timeout = time.Duration(e.Timeout) * time.Second
	defer client.Release()

	var firstByteTime time.Time
	startTime := time.Now()

	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			firstByteTime = time.Now()
		},
	}

	reqCtx := httptrace.WithClientTrace(ctx, trace)
	request, err := http.NewRequestWithContext(reqCtx, "GET", e.URL, nil)
	if err != nil {
		return 0, false
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, false
	}
	defer response.Body.Close()

	// 确保我们获取到了首字节时间
	if firstByteTime.IsZero() {
		firstByteTime = time.Now()
	}

	delay := uint16(firstByteTime.Sub(startTime).Milliseconds())
	if response.StatusCode == e.ExptectCode {
		return delay, true
	}
	return delay, false
}

func init() {
	register.Check(&Alive{})
}
