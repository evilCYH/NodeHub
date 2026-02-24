package system

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/evilCYH/NodeHub/internal/models/system"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/shirou/gopsutil/v4/process"
)

var (
	startTime     string
	uploadBytes   uint64
	downloadBytes uint64
)

func init() {
	startTime = time.Now().Format(time.RFC3339)
}

func AddUploadBytes(bytes uint64) {
	atomic.AddUint64(&uploadBytes, bytes)
}

func AddDownloadBytes(bytes uint64) {
	atomic.AddUint64(&downloadBytes, bytes)
}

func GetSystemInfo() *system.Info {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Debugf("Failed to create process instance: %v", err)
		return nil
	}

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		log.Debugf("Failed to get process memory info: %v", err)
		return nil
	}

	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		log.Debugf("Failed to get process CPU percent: %v", err)
		return nil
	}

	return &system.Info{
		MemoryUsed:    memInfo.RSS,
		CPUPercent:    cpuPercent,
		StartTime:     startTime,
		UploadBytes:   atomic.LoadUint64(&uploadBytes),
		DownloadBytes: atomic.LoadUint64(&downloadBytes),
	}
}

func Reset() {
	atomic.StoreUint64(&uploadBytes, 0)
	atomic.StoreUint64(&downloadBytes, 0)
}
