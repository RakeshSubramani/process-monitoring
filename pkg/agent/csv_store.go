package agent

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

type CSVStore struct {
	f *os.File
	w *csv.Writer
}

func NewCSVStore(path string) (*CSVStore, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)
	// if empty file, write header
	info, _ := f.Stat()
	if info.Size() == 0 {
		w.Write([]string{"ts", "cpu_pct", "mem_mb", "disk_mb", "net_sent_bytes", "net_recv_bytes", "up_mb_s", "down_mb_s"})
		w.Flush()
	}
	return &CSVStore{f: f, w: w}, nil
}

func (c *CSVStore) AppendSnapshotCSV(sn Snapshot) error {
	rec := []string{
		sn.Timestamp.UTC().Format(time.RFC3339),
		fmt.Sprintf("%.2f", sn.System.CPUPercent),
		fmt.Sprintf("%.2f", sn.System.MemoryUsedMB),
		fmt.Sprintf("%.2f", sn.System.DiskUsedMB),
		strconv.FormatUint(sn.System.NetBytesSent, 10),
		strconv.FormatUint(sn.System.NetBytesRecv, 10),
		fmt.Sprintf("%.4f", sn.System.UploadSpeedMB),
		fmt.Sprintf("%.4f", sn.System.DownloadSpeedMB),
	}
	if err := c.w.Write(rec); err != nil {
		return err
	}
	c.w.Flush()
	return nil
}

func (c *CSVStore) Close() error {
	c.w.Flush()
	return c.f.Close()
}
