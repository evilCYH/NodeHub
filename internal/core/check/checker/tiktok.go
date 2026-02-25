package checker

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/evilCYH/NodeHub/internal/core/mihomo"
	"github.com/evilCYH/NodeHub/internal/core/node"
	"github.com/evilCYH/NodeHub/internal/core/task"
	"github.com/evilCYH/NodeHub/internal/database/op"
	"github.com/evilCYH/NodeHub/internal/models/check"
	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	"github.com/evilCYH/NodeHub/internal/modules/register"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/evilCYH/NodeHub/internal/utils/ua"
)

type TikTok struct {
	Thread  int `json:"thread" name:"线程数" value:"200"`
	Timeout int `json:"timeout" name:"超时时间" value:"10" desc:"单个节点检测的超时时间(s)"`
}

func (e *TikTok) Init() error {
	return nil
}

func (e *TikTok) Run(ctx context.Context, log *log.Logger, subID []uint16) check.Result {
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
	if threads == 0 || len(nodes) == 0 {
		log.Warnf("tiktok check task failed, no nodes")
		return check.Result{
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
			result := e.detectTikTok(ctx, raw)
			switch result {
			case 1:
				n.Info.SetAliveStatus(nodeModel.TikTok, true)
				n.Info.SetAliveStatus(nodeModel.TikTokIDC, false)
			case 2:
				n.Info.SetAliveStatus(nodeModel.TikTok, false)
				n.Info.SetAliveStatus(nodeModel.TikTokIDC, true)
			default:
				n.Info.SetAliveStatus(nodeModel.TikTok, false)
				n.Info.SetAliveStatus(nodeModel.TikTokIDC, false)
			}
			node.UpdateNodeTikTokInPool(
				n.Base.UniqueKey,
				n.Info.AliveStatus&nodeModel.TikTok != 0,
				n.Info.AliveStatus&nodeModel.TikTokIDC != 0,
			)
			node.UpdateRegistryTikTok(n.Base.SubId, n.Base.UniqueKey, n.Info.AliveStatus, "tiktok_task")
			message := "tiktok: unavailable"
			if result == 1 {
				message = "tiktok: available"
			} else if result == 2 {
				message = "tiktok: idc"
			}
			if err := op.CreateNodeLog(ctx, &nodeModel.NodeLog{
				SubID:     n.Base.SubId,
				NodeKey:   n.Base.UniqueKey,
				NodeName:  nodeName,
				Level:     "info",
				Source:    nodeModel.LogSourceCheck,
				CheckID:   checkID,
				Message:   message,
				CreatedAt: time.Now(),
			}); err != nil {
				log.Warnf("failed to create node log: %v", err)
			}
		}); err != nil {
			<-sem
			wg.Done()
			log.Warnf("tiktok check task submit failed: %v", err)
		}
	}
	wg.Wait()

	log.Debugf("tiktok check task end")
	return check.Result{
		Msg:      "success",
		LastRun:  time.Now(),
		Duration: time.Since(startTime).Milliseconds(),
	}
}

func (e *TikTok) detectTikTok(ctx context.Context, raw map[string]any) uint8 {
	client := mihomo.Proxy(raw)
	if client == nil {
		return 0
	}
	client.Timeout = time.Duration(e.Timeout) * time.Second
	defer client.Release()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.tiktok.com/", nil)
	if err != nil {
		return 0
	}

	ua.SetHeader(req)
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0
	}
	if extractRegion(body) {
		return 1
	}

	req, err = http.NewRequestWithContext(ctx, "GET", "https://www.tiktok.com/api/passport/web/region/get/", nil)
	if err != nil {
		return 0
	}
	ua.SetHeader(req)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "en")

	resp, err = client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0
	}
	if extractRegion(body) {
		return 2
	}

	return 0
}

func extractRegion(html []byte) bool {
	return bytes.Contains(html, []byte(`"region":`))
}

func init() {
	register.Check(&TikTok{})
}
