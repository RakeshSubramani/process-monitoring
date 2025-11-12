package agent

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	// Ensure parent folder exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	s := &SQLiteStore{db: db}

	// Initialize schema
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return s, nil
}

func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ts DATETIME,
		system_json TEXT,
		processes_json TEXT,
		conns_json TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_snap_ts ON snapshots(ts);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) InsertSnapshot(sn Snapshot) error {
	sysj, _ := json.Marshal(sn.System)
	procj, _ := json.Marshal(sn.Processes)
	connj, _ := json.Marshal(sn.Connections)
	_, err := s.db.Exec("INSERT INTO snapshots(ts, system_json, processes_json, conns_json) VALUES (?, ?, ?, ?)",
		sn.Timestamp.UTC().Format(time.RFC3339), string(sysj), string(procj), string(connj))
	return err
}

func (s *SQLiteStore) GetLastSnapshots(n int) ([]Snapshot, error) {
	rows, err := s.db.Query("SELECT ts, system_json, processes_json, conns_json FROM snapshots ORDER BY ts DESC LIMIT ?", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Snapshot
	for rows.Next() {
		var tsStr, sysj, procj, connj string
		if err := rows.Scan(&tsStr, &sysj, &procj, &connj); err != nil {
			return nil, err
		}
		var sn Snapshot
		ts, _ := time.Parse(time.RFC3339, tsStr)
		sn.Timestamp = ts
		json.Unmarshal([]byte(sysj), &sn.System)
		json.Unmarshal([]byte(procj), &sn.Processes)
		json.Unmarshal([]byte(connj), &sn.Connections)
		out = append(out, sn)
	}
	return out, nil
}

// package agent

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"time"

// 	_ "github.com/mattn/go-sqlite3"
// )

// type SQLiteStore struct {
// 	db *sql.DB
// }

// // path should be something like "data/monitor.db"
// func NewSQLiteStore(path string) (*SQLiteStore, error) {
// 	// ✅ Ensure parent folder exists
// 	dir := filepath.Dir(path)
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		return nil, fmt.Errorf("failed to create data directory: %w", err)
// 	}

// 	// ✅ Open database
// 	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open sqlite: %w", err)
// 	}

// 	s := &SQLiteStore{db: db}

// 	// ✅ Initialize schema
// 	if err := s.initSchema(); err != nil {
// 		return nil, fmt.Errorf("failed to init schema: %w", err)
// 	}

// 	return s, nil
// }

// func (s *SQLiteStore) initSchema() error {
// 	schema := `
// 	CREATE TABLE IF NOT EXISTS snapshots (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		ts DATETIME,
// 		system_json TEXT,
// 		processes_json TEXT,
// 		conns_json TEXT
// 	);
// 	CREATE INDEX IF NOT EXISTS idx_snap_ts ON snapshots(ts);
// 	`
// 	_, err := s.db.Exec(schema)
// 	return err
// }

// func (s *SQLiteStore) InsertSnapshot(sn Snapshot) error {
// 	sysj, _ := json.Marshal(sn.System)
// 	procj, _ := json.Marshal(sn.Processes)
// 	connj, _ := json.Marshal(sn.Connections)
// 	_, err := s.db.Exec("INSERT INTO snapshots(ts, system_json, processes_json, conns_json) VALUES (?, ?, ?, ?)",
// 		sn.Timestamp.UTC().Format(time.RFC3339), string(sysj), string(procj), string(connj))
// 	return err
// }

// func (s *SQLiteStore) GetLastSnapshots(n int) ([]Snapshot, error) {
// 	rows, err := s.db.Query("SELECT ts, system_json, processes_json, conns_json FROM snapshots ORDER BY ts DESC LIMIT ?", n)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var out []Snapshot
// 	for rows.Next() {
// 		var tsStr, sysj, procj, connj string
// 		if err := rows.Scan(&tsStr, &sysj, &procj, &connj); err != nil {
// 			return nil, err
// 		}
// 		var sn Snapshot
// 		ts, _ := time.Parse(time.RFC3339, tsStr)
// 		sn.Timestamp = ts
// 		json.Unmarshal([]byte(sysj), &sn.System)
// 		json.Unmarshal([]byte(procj), &sn.Processes)
// 		json.Unmarshal([]byte(connj), &sn.Connections)
// 		out = append(out, sn)
// 	}
// 	return out, nil
// }
