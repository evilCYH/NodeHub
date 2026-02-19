package checker

import (
	"context"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bestruirui/bestsub/internal/core/mihomo"
	"github.com/bestruirui/bestsub/internal/core/node"
	"github.com/bestruirui/bestsub/internal/core/task"
	"github.com/bestruirui/bestsub/internal/database/op"
	checkModel "github.com/bestruirui/bestsub/internal/models/check"
	nodeModel "github.com/bestruirui/bestsub/internal/models/node"
	"github.com/bestruirui/bestsub/internal/modules/country"
	"github.com/bestruirui/bestsub/internal/modules/register"
	"github.com/bestruirui/bestsub/internal/utils/log"
)

type Country struct {
	Thread  int `json:"thread" name:"线程数" value:"100"`
	Timeout int `json:"timeout" name:"超时时间" value:"10" desc:"单个节点检测的超时时间(s)"`
}

func (e *Country) Init() error {
	return nil
}

func (e *Country) Run(ctx context.Context, log *log.Logger, subID []uint16) checkModel.Result {
	startTime := time.Now()
	checkID := getCheckID(ctx)
	var nodes []nodeModel.Data
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
	if threads == 0 {
		log.Warnf("country check task failed, no nodes")
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
		if nd.Info.AliveStatus&nodeModel.Country != 0 {
			continue
		}
		sem <- struct{}{}
		wg.Add(1)
		n := nd
		task.Submit(func() {
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
			client := mihomo.Proxy(raw)
			if client == nil {
				if err := op.CreateNodeLog(ctx, &nodeModel.NodeLog{
					SubID:     n.Base.SubId,
					NodeKey:   n.Base.UniqueKey,
					NodeName:  nodeName,
					Level:     "error",
					Source:    nodeModel.LogSourceCheck,
					CheckID:   checkID,
					Message:   "proxy parse failed",
					CreatedAt: time.Now(),
				}); err != nil {
					log.Warnf("failed to create node log: %v", err)
				}
				return
			}
			client.Timeout = time.Duration(e.Timeout) * time.Second
			defer client.Release()
			countryCode := country.GetCode(ctx, client.Client)
			if countryCode != "" {
				n.Info.Country = countryCode
				n.Info.SetAliveStatus(nodeModel.Country, true)
			} else {
				n.Info.SetAliveStatus(nodeModel.Country, false)
			}
			node.UpdateRegistryCountry(n.Base.SubId, n.Base.UniqueKey, n.Info.Country, countryCode != "", "country_task")
			if err := op.CreateNodeLog(ctx, &nodeModel.NodeLog{
				SubID:     n.Base.SubId,
				NodeKey:   n.Base.UniqueKey,
				NodeName:  nodeName,
				Level:     "info",
				Source:    nodeModel.LogSourceCheck,
				CheckID:   checkID,
				Message:   "country: " + n.Info.Country,
				CreatedAt: time.Now(),
			}); err != nil {
				log.Warnf("failed to create node log: %v", err)
			}
		})
	}
	wg.Wait()
	return checkModel.Result{
		Msg:      "success",
		LastRun:  time.Now(),
		Duration: time.Since(startTime).Milliseconds(),
	}
}

func init() {
	register.Check(&Country{})
}
