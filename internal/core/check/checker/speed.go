package checker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bestruirui/bestsub/internal/core/mihomo"
	"github.com/bestruirui/bestsub/internal/core/node"
	"github.com/bestruirui/bestsub/internal/core/system"
	"github.com/bestruirui/bestsub/internal/core/task"
	"github.com/bestruirui/bestsub/internal/database/op"
	checkModel "github.com/bestruirui/bestsub/internal/models/check"
	nodeModel "github.com/bestruirui/bestsub/internal/models/node"
	"github.com/bestruirui/bestsub/internal/modules/register"
	"github.com/bestruirui/bestsub/internal/utils/log"
)

const mbToBytes = 1024 * 1024

type Speed struct {
	Thread  int `json:"thread" name:"线程数" value:"5"`
	Timeout int `json:"timeout" name:"超时时间" value:"60" desc:"单个节点检测的超时时间(s)"`

	Download      bool   `json:"download" name:"下载测试" value:"true"`
	DownloadSkip  bool   `json:"download_skip" name:"是否跳过已经有下载速度的节点" value:"false"`
	DownloadUrl   string `json:"download_url" name:"测试链接" value:"https://speed.cloudflare.com/__down?bytes=104857600" desc:"最好自定义一个测试链接,部分节点可能屏蔽此默认链接"`
	DownloadSize  int64  `json:"download_size" name:"下载大小" value:"100" desc:"到达指定大小后停止测速(MB)"`
	DownloadSpeed int64  `json:"download_speed" name:"下载速度" value:"1" desc:"下载速度达到指定值并且达到指定个数后停止测速(KB/s)"`
	DownloadCount int    `json:"download_count" name:"节点个数" value:"5" desc:"符合下载速度的节点个数,满足后停止测试"`

	Upload      bool   `json:"upload" name:"上传测试" value:"false"`
	UploadSkip  bool   `json:"upload_skip" name:"是否跳过已经有上传速度的节点" value:"false"`
	UploadUrl   string `json:"upload_url" name:"上传链接" value:"https://speed.cloudflare.com/__up" desc:"最好自定义一个测试链接,部分节点可能屏蔽此默认链接"`
	UploadSize  int64  `json:"upload_size" name:"上传大小" value:"100" desc:"到达指定大小后停止测速(MB)"`
	UploadSpeed int64  `json:"upload_speed" name:"上传速度" value:"1" desc:"上传速度达到指定值并且达到指定个数后停止测速(KB/s)"`
	UploadCount int    `json:"upload_count" name:"节点个数" value:"5" desc:"符合上传速度的节点个数,满足后停止测试"`
}

func (e *Speed) Init() error {
	return nil
}

func (e *Speed) Run(ctx context.Context, log *log.Logger, subID []uint16) checkModel.Result {
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
		log.Warnf("speed check task failed, no nodes")
		return checkModel.Result{
			Msg:      "no nodes",
			LastRun:  time.Now(),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}
	sem := make(chan struct{}, threads)
	defer close(sem)
	var downloadCount, uploadCount int

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
			defer client.Release()
			client.Timeout = time.Duration(e.Timeout) * time.Second
			if e.Download && downloadCount < e.DownloadCount && (!e.DownloadSkip || n.Info.SpeedDown.Average() == 0) {
				speed := e.download(ctx, client.Client)
				if speed > 0 {
					n.Info.SpeedDown.Update(uint32(speed))
					node.UpdateRegistrySpeed(n.Base.SubId, n.Base.UniqueKey, 0, uint32(speed), "speed_task")
					if err := op.CreateNodeLog(ctx, &nodeModel.NodeLog{
						SubID:     n.Base.SubId,
						NodeKey:   n.Base.UniqueKey,
						NodeName:  nodeName,
						Level:     "info",
						Source:    nodeModel.LogSourceCheck,
						CheckID:   checkID,
						Message:   fmt.Sprintf("download: %d KB/s", speed),
						CreatedAt: time.Now(),
					}); err != nil {
						log.Warnf("failed to create node log: %v", err)
					}
				}
				if speed > e.DownloadSpeed {
					downloadCount++
				}
			}
			client.Timeout = time.Duration(e.Timeout) * time.Second
			if e.Upload && uploadCount < e.UploadCount && (!e.UploadSkip || n.Info.SpeedUp.Average() == 0) {
				speed := e.upload(ctx, client.Client)
				if speed > 0 {
					n.Info.SpeedUp.Update(uint32(speed))
					node.UpdateRegistrySpeed(n.Base.SubId, n.Base.UniqueKey, uint32(speed), 0, "speed_task")
					if err := op.CreateNodeLog(ctx, &nodeModel.NodeLog{
						SubID:     n.Base.SubId,
						NodeKey:   n.Base.UniqueKey,
						NodeName:  nodeName,
						Level:     "info",
						Source:    nodeModel.LogSourceCheck,
						CheckID:   checkID,
						Message:   fmt.Sprintf("upload: %d KB/s", speed),
						CreatedAt: time.Now(),
					}); err != nil {
						log.Warnf("failed to create node log: %v", err)
					}
				}
				if speed > e.UploadSpeed {
					uploadCount++
				}
			}
		}); err != nil {
			<-sem
			wg.Done()
			log.Warnf("speed check task submit failed: %v", err)
		}
	}
	wg.Wait()
	return checkModel.Result{
		Msg:      fmt.Sprintf("success, download count: %d, upload count: %d", downloadCount, uploadCount),
		LastRun:  time.Now(),
		Duration: time.Since(startTime).Milliseconds(),
	}
}
func (e *Speed) download(ctx context.Context, client *http.Client) int64 {
	request, err := http.NewRequestWithContext(ctx, "GET", e.DownloadUrl, nil)
	if err != nil {
		return 0
	}
	response, err := client.Do(request)
	if err != nil {
		return 0
	}
	defer response.Body.Close()
	startTime := time.Now()

	limitReader := io.LimitReader(response.Body, e.DownloadSize*mbToBytes)
	bytes, _ := io.Copy(io.Discard, limitReader)
	duration := time.Since(startTime).Milliseconds()
	if duration <= 0 || bytes <= 0 {
		return 0
	}
	system.AddDownloadBytes(uint64(bytes))
	return bytes / duration
}

func (e *Speed) upload(ctx context.Context, client *http.Client) int64 {
	uploadBytes := e.UploadSize * mbToBytes
	reader := &trackingZeroReader{remaining: uploadBytes}
	request, err := http.NewRequestWithContext(ctx, "POST", e.UploadUrl, reader)
	if err != nil {
		return 0
	}
	request.ContentLength = uploadBytes
	startTime := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return 0
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return 0
	}
	io.Copy(io.Discard, response.Body)
	duration := time.Since(startTime).Milliseconds()
	if duration <= 0 || reader.bytesRead <= 0 {
		return 0
	}
	system.AddUploadBytes(uint64(reader.bytesRead))
	return reader.bytesRead / duration
}

type trackingZeroReader struct {
	remaining int64
	bytesRead int64
}

func (r *trackingZeroReader) Read(p []byte) (n int, err error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}

	if int64(len(p)) > r.remaining {
		n = int(r.remaining)
	} else {
		n = len(p)
	}

	clear(p[:n])

	r.remaining -= int64(n)
	r.bytesRead += int64(n)
	return n, nil
}

func init() {
	register.Check(&Speed{})
}
