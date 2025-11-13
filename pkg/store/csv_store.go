package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

type CSVStore struct {
	filePath string
}

// Snapshot holds one complete system monitoring snapshot
type Snapshot struct {
	Timestamp   time.Time     `json:"timestamp"`
	System      Metrics       `json:"system"`
	Processes   []ProcessInfo `json:"processes"`
	Connections []ConnInfo    `json:"connections,omitempty"`
}

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

// ConnInfo represents a network connection
type ConnInfo struct {
	Pid    int32  `json:"pid"`
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Domain string `json:"domain,omitempty"`
	Status string `json:"status"`
}

// NewCSVStore creates a new CSV file and writes header if needed
func NewCSVStore(path string) (*CSVStore, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		header := []string{
			"Timestamp", "CPU (%)", "Memory (%)", "Disk (%)", "Net Sent (MB)", "Net Recv (MB)",
		}
		writer.Write(header)
	}
	return &CSVStore{filePath: path}, nil
}

// AppendSnapshotCSV writes a single snapshot row to CSV
func (c *CSVStore) AppendSnapshotCSV(snap Snapshot) error {
	file, err := os.OpenFile(c.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	row := []string{
		snap.Timestamp.Format(time.RFC3339),
		fmt.Sprintf("%.2f", snap.System.CPUPercent),
		fmt.Sprintf("%.2f", snap.System.MemoryPercent),
		// fmt.Sprintf("%.2f", snap.System.DiskPercent),
		// strconv.FormatFloat(snap.System.NetSentMB, 'f', 2, 64),
		// strconv.FormatFloat(snap.System.NetRecvMB, 'f', 2, 64),
	}

	return writer.Write(row)
}
