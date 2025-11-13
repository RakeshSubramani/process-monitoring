package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore initializes SQLite and creates tables if not exist
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		system_cpu REAL,
		system_mem REAL,
		system_disk REAL,
		system_net_sent REAL,
		system_net_recv REAL,
		processes TEXT,
		connections TEXT
	);`

	if _, err := db.Exec(createTable); err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

// InsertSnapshot inserts a system snapshot into DB
func (s *SQLiteStore) InsertSnapshot(snap Snapshot) error {
	procJSON, _ := json.Marshal(snap.Processes)
	connJSON, _ := json.Marshal(snap.Connections)

	_, err := s.db.Exec(`
	INSERT INTO snapshots 
	(timestamp, system_cpu, system_mem, system_disk, system_net_sent, system_net_recv, processes, connections)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		snap.Timestamp,
		snap.System.CPUPercent,
		snap.System.MemoryPercent,
		// snap.System.DiskPercent,
		// snap.System.NetSentMB,
		// snap.System.NetRecvMB,
		string(procJSON),
		string(connJSON),
	)
	return err
}

// GetRecentSnapshots fetches last N records
func (s *SQLiteStore) GetRecentSnapshots(limit int) ([]Snapshot, error) {
	rows, err := s.db.Query(`
		SELECT timestamp, system_cpu, system_mem, system_disk, system_net_sent, system_net_recv, processes, connections
		FROM snapshots
		ORDER BY id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []Snapshot

	for rows.Next() {
		var ts time.Time
		var cpu, mem, disk, sent, recv float64
		var procJSON, connJSON string

		if err := rows.Scan(&ts, &cpu, &mem, &disk, &sent, &recv, &procJSON, &connJSON); err != nil {
			continue
		}

		var procs []ProcessInfo
		var conns []ConnInfo
		_ = json.Unmarshal([]byte(procJSON), &procs)
		_ = json.Unmarshal([]byte(connJSON), &conns)

		snapshots = append(snapshots, Snapshot{
			Timestamp: ts,
			System: Metrics{
				CPUPercent:    cpu,
				MemoryPercent: mem,
				// DiskPercent:   disk,
				// NetSentMB:     sent,
				// NetRecvMB:     recv,
			},
			Processes:   procs,
			Connections: conns,
		})
	}

	return snapshots, nil
}
