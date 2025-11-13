package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	models "github.com/RakeshSubramani/process-monitoring/pkg/model"
)

func InitCSV(path string) error {
	_, err := NewCSVStore(path)
	return err
}

func NewCSVStore(path string) (*CSVStore, error) {
	// create if not exists and write header
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	f.Close()
	c := &CSVStore{Path: path}

	// write header if file empty
	fi, _ := os.Stat(path)
	if fi.Size() == 0 {
		file, _ := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
		w := csv.NewWriter(file)
		w.Write([]string{"ts", "cpu_percent", "mem_percent", "disk_used_mb", "net_sent", "net_recv"})
		w.Flush()
		file.Close()
	}
	return c, nil
}

func (c *CSVStore) AppendSnapshotCSV(sn models.Snapshot) error {
	file, err := os.OpenFile(c.Path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	_ = w.Write([]string{
		sn.Timestamp.UTC().Format(time.RFC3339),
		fmt.Sprintf("%.3f", sn.System.CPUPercent),
		fmt.Sprintf("%.3f", sn.System.MemoryPercent),
		fmt.Sprintf("%.3f", sn.System.DiskUsedMB),
		strconv.FormatUint(sn.System.NetBytesSent, 10),
		strconv.FormatUint(sn.System.NetBytesRecv, 10),
	})
	w.Flush()
	return w.Error()
}
