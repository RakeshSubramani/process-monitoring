package agent

import (
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// Metrics holds system metrics
type Metrics struct {
	CPUPercent       float64   `json:"cpu_percent"`
	PerCore          []float64 `json:"per_core,omitempty"`
	MemoryUsedMB     float64   `json:"memory_used_mb"`
	MemoryTotalMB    float64   `json:"memory_total_mb"`
	MemoryPercent    float64   `json:"memory_percent"`
	DiskUsedMB       float64   `json:"disk_used_mb"`
	DiskTotalMB      float64   `json:"disk_total_mb"`
	Load1            float64   `json:"load1"`
	Load5            float64   `json:"load5"`
	Load15           float64   `json:"load15"`
	NetBytesSent     uint64    `json:"net_bytes_sent"`
	NetBytesRecv     uint64    `json:"net_bytes_recv"`
	UploadSpeedMBs   float64   `json:"upload_mbps"`
	DownloadSpeedMBs float64   `json:"download_mbps"`
	Timestamp        time.Time `json:"timestamp"`
}

// ProcessInfo minimal process info
type ProcessInfo struct {
	PID    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu_percent"`
	Memory float32 `json:"memory_percent"`
}

// CollectSystem collects system metrics, returns metrics and current totals for network
func CollectSystem(prevSent, prevRecv uint64, prevTime time.Time) (Metrics, uint64, uint64, time.Time, error) {
	now := time.Now()

	perCore, _ := cpu.Percent(0, true)
	total, _ := cpu.Percent(0, false)

	memStats, _ := mem.VirtualMemory()
	diskStats, _ := disk.Usage("/")
	l, _ := load.Avg()

	netStats, _ := gnet.IOCounters(false)

	var sent uint64
	var recv uint64
	if len(netStats) > 0 {
		sent = netStats[0].BytesSent
		recv = netStats[0].BytesRecv
	}

	var upSpeed, downSpeed float64
	if !prevTime.IsZero() {
		secs := now.Sub(prevTime).Seconds()
		if secs > 0 {
			upSpeed = float64(sent-prevSent) / secs / 1024 / 1024
			downSpeed = float64(recv-prevRecv) / secs / 1024 / 1024
		}
	}

	m := Metrics{
		CPUPercent:       0,
		PerCore:          perCore,
		MemoryUsedMB:     float64(memStats.Used) / 1024 / 1024,
		MemoryTotalMB:    float64(memStats.Total) / 1024 / 1024,
		MemoryPercent:    memStats.UsedPercent,
		DiskUsedMB:       float64(diskStats.Used) / 1024 / 1024,
		DiskTotalMB:      float64(diskStats.Total) / 1024 / 1024,
		Load1:            l.Load1,
		Load5:            l.Load5,
		Load15:           l.Load15,
		NetBytesSent:     sent,
		NetBytesRecv:     recv,
		UploadSpeedMBs:   upSpeed,
		DownloadSpeedMBs: downSpeed,
		Timestamp:        now,
	}

	if len(total) > 0 {
		m.CPUPercent = total[0]
	}

	return m, sent, recv, now, nil
}

// CollectTopProcesses returns top N processes sorted by CPU (if limit==0 returns all)
func CollectTopProcesses(limit int) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	out := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}
		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()

		out = append(out, ProcessInfo{
			PID:    p.Pid,
			Name:   name,
			CPU:    cpuPct,
			Memory: memPct,
		})
	}

	// sort desc by CPU
	// simple bubble? use sort
	// import sort at top
	// but to avoid extra import here, we'll sort
	// actually add import
	// (see top imports)
	// we'll rely on caller to import sort in this file; add it now

	// we'll perform sort below
	return sortProcesses(out, limit), nil
}

// helper to sort and slice
func sortProcesses(list []ProcessInfo, limit int) []ProcessInfo {
	sort.Slice(list, func(i, j int) bool { return list[i].CPU > list[j].CPU })
	if limit > 0 && len(list) > limit {
		return list[:limit]
	}
	return list
}
