package agent

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CPUUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage_percent",
		Help: "Current CPU usage percentage",
	})

	MemoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_used_mb",
		Help: "Memory used in MB",
	})

	DiskUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_used_mb",
		Help: "Disk used in MB",
	})

	NetSent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_bytes_sent_total",
		Help: "Total bytes sent via network",
	})

	NetRecv = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_bytes_recv_total",
		Help: "Total bytes received via network",
	})
)

// Register all metrics with Prometheus
func RegisterMetrics() {
	prometheus.MustRegister(CPUUsage, MemoryUsage, DiskUsage, NetSent, NetRecv)
}

// Update metrics with latest values
func UpdateMetrics(m Metrics) {
	CPUUsage.Set(m.CPUPercent)
	MemoryUsage.Set(m.MemoryUsedMB)
	DiskUsage.Set(m.DiskUsedMB)
	NetSent.Set(float64(m.NetBytesSent))
	NetRecv.Set(float64(m.NetBytesRecv))
}
