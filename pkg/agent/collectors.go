package agent

import (
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type Metrics struct {
	CPUPercent      float64
	MemoryUsedMB    float64
	DiskUsedMB      float64
	NetBytesSent    uint64
	NetBytesRecv    uint64
	UploadSpeedMB   float64
	DownloadSpeedMB float64
}

// ProcessInfo for API/UI/Prometheus
type ProcessInfo struct {
	PID    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu_percent"`
	Memory float32 `json:"memory_percent"`
}

// CollectSystem collects cpu/mem/disk/net + bandwidth speeds
func CollectSystem(prevSent, prevRecv uint64, prevTime time.Time) (Metrics, uint64, uint64, time.Time, error) {
	cpuPercent, _ := cpu.Percent(0, false)
	memStats, _ := mem.VirtualMemory()
	diskStats, _ := disk.Usage("/")
	netStats, _ := net.IOCounters(false)

	now := time.Now()
	var upSpeed, downSpeed float64
	if !prevTime.IsZero() {
		secs := now.Sub(prevTime).Seconds()
		if secs > 0 {
			upSpeed = float64(netStats[0].BytesSent-prevSent) / secs / 1024 / 1024
			downSpeed = float64(netStats[0].BytesRecv-prevRecv) / secs / 1024 / 1024
		}
	}

	m := Metrics{
		CPUPercent:      cpuPercent[0],
		MemoryUsedMB:    float64(memStats.Used) / 1024 / 1024,
		DiskUsedMB:      float64(diskStats.Used) / 1024 / 1024,
		NetBytesSent:    netStats[0].BytesSent,
		NetBytesRecv:    netStats[0].BytesRecv,
		UploadSpeedMB:   upSpeed,
		DownloadSpeedMB: downSpeed,
	}
	return m, netStats[0].BytesSent, netStats[0].BytesRecv, now, nil
}

// CollectTopProcesses returns top N processes sorted by CPU
func CollectTopProcesses(limit int) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	// First sampling (warm-up)
	for _, p := range procs {
		_, _ = p.CPUPercent()
	}

	time.Sleep(300 * time.Millisecond) // allow time to measure CPU usage

	out := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()

		if cpuPct == 0 && memPct == 0 {
			continue // ignore idle processes
		}

		out = append(out, ProcessInfo{
			PID:    p.Pid,
			Name:   name,
			CPU:    cpuPct,
			Memory: memPct,
		})
	}

	// sort by CPU descending
	sort.Slice(out, func(i, j int) bool { return out[i].CPU > out[j].CPU })

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// CollectAllProcesses returns all processes sorted by CPU
func CollectAllProcesses() ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	out := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
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

	sort.Slice(out, func(i, j int) bool { return out[i].CPU > out[j].CPU })
	return out, nil
}
