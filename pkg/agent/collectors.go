package agent

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// Metrics holds system data
type Metrics struct {
	CPUPercent   float64
	MemoryUsedMB float64
	DiskUsedMB   float64
	NetBytesSent uint64
	NetBytesRecv uint64
}

// CollectMetrics gathers current system stats
func CollectMetrics() (Metrics, error) {
	cpuPercent, _ := cpu.Percent(0, false)
	memStats, _ := mem.VirtualMemory()
	diskStats, _ := disk.Usage("/")
	netStats, _ := net.IOCounters(false)

	return Metrics{
		CPUPercent:   cpuPercent[0],
		MemoryUsedMB: float64(memStats.Used) / 1024 / 1024,
		DiskUsedMB:   float64(diskStats.Used) / 1024 / 1024,
		NetBytesSent: netStats[0].BytesSent,
		NetBytesRecv: netStats[0].BytesRecv,
	}, nil
}

// AutoCollect runs the collector every interval seconds
func AutoCollect(interval time.Duration, update func(Metrics)) {
	for {
		m, _ := CollectMetrics()
		update(m)
		time.Sleep(interval)
	}
}
