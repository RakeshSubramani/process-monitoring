package models

import (
	"time"
)

// Metrics holds system metrics
type Metrics struct {
	CPUPercent       float64   `json:"cpu_percent"`
	PerCore          []float64 `json:"per_core"`
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

type ProcessInfo struct {
	Pid        int32
	Name       string
	CPUPercent float64
	MemPercent float32
}

type Snapshot struct {
	Timestamp   time.Time     `json:"timestamp"`
	System      Metrics       `json:"system"`
	Processes   []ProcessInfo `json:"processes"`
	Connections []ConnInfo    `json:"connections,omitempty"`
	Ready       bool          `json:"ready"`
}

type ConnInfo struct {
	Pid    int32  `json:"pid"`
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Domain string `json:"domain,omitempty"`
	Status string `json:"status"`
}
