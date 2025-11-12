package agent

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// System-level gauges
	GCPU = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage_percent",
		Help: "Current CPU usage percent",
	})
	GMem = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_used_mb",
		Help: "Memory used in MB",
	})
	GDisk = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_used_mb",
		Help: "Disk used in MB",
	})
	GNetSent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_bytes_sent_total",
		Help: "Total bytes sent",
	})
	GNetRecv = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_bytes_recv_total",
		Help: "Total bytes recv",
	})

	// Per-process gauges (labelled by pid and process name)
	GProcCPU = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "process_cpu_percent",
		Help: "CPU percent per process",
	}, []string{"pid", "name"})

	GProcMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "process_memory_percent",
		Help: "Memory percent per process",
	}, []string{"pid", "name"})
)

func RegisterPromMetrics() {
	prometheus.MustRegister(GCPU, GMem, GDisk, GNetSent, GNetRecv, GProcCPU, GProcMem)
}

func UpdatePromMetrics(sys Metrics, procs []ProcessInfo) {
	GCPU.Set(sys.CPUPercent)
	GMem.Set(sys.MemoryUsedMB)
	GDisk.Set(sys.DiskUsedMB)
	GNetSent.Set(float64(sys.NetBytesSent))
	GNetRecv.Set(float64(sys.NetBytesRecv))

	// Reset per-process vectors before setting new values to avoid stale labels
	GProcCPU.Reset()
	GProcMem.Reset()

	for _, p := range procs {
		labels := prometheus.Labels{"pid": fmt.Sprintf("%d", p.PID), "name": p.Name}
		GProcCPU.With(labels).Set(p.CPU)
		GProcMem.With(labels).Set(float64(p.Memory))
	}
}
