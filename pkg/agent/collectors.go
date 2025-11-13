package agent

import (
	"sort"
	"time"

	models "github.com/RakeshSubramani/process-monitoring/pkg/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessInfo minimal process info

func CollectSystem(prevSent, prevRecv uint64, prevTime time.Time) (models.Metrics, uint64, uint64, time.Time, error) {
	now := time.Now()
	perCore, _ := cpu.Percent(0, true)
	total, _ := cpu.Percent(0, false)
	memStats, _ := mem.VirtualMemory()
	diskStats, _ := disk.Usage("/")
	l, _ := load.Avg()
	netStats, _ := gnet.IOCounters(false)

	var sent, recv uint64
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

	m := models.Metrics{
		CPUPercent:       0.0,
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

func CollectTopProcesses(limit int) ([]models.ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	out := make([]models.ProcessInfo, 0, len(procs))
	for _, p := range procs {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}
		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()
		out = append(out, models.ProcessInfo{Pid: p.Pid, Name: name, CPUPercent: cpuPct, MemPercent: memPct})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CPUPercent > out[j].CPUPercent })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
