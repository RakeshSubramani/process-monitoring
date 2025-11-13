package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	models "github.com/RakeshSubramani/process-monitoring/pkg/model"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	Db *sql.DB
}
type CSVStore struct {
	Path string
}

func InitSQLite(path string) error {
	s, err := NewSQLiteStore(path)
	if err != nil {
		return err
	}
	// we close immediately because agent opens another instance — keep this as health-check
	_ = s
	return nil
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path+"?_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	s := &SQLiteStore{Db: db}
	if err := s.InitSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) InitSchema() error {
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
	_, err := s.Db.Exec(schema)
	return err
}

func (s *SQLiteStore) InsertSnapshot(sn models.Snapshot) error {
	sysj, _ := json.Marshal(sn.System)
	procj, _ := json.Marshal(sn.Processes)
	connj, _ := json.Marshal(sn.Connections)
	_, err := s.Db.Exec("INSERT INTO snapshots(ts, system_json, processes_json, conns_json) VALUES (?, ?, ?, ?)",
		sn.Timestamp.UTC().Format(time.RFC3339), string(sysj), string(procj), string(connj))
	return err
}

func (s *SQLiteStore) GetRecentSnapshots(n int) ([]models.Snapshot, error) {
	rows, err := s.Db.Query("SELECT ts, system_json, processes_json, conns_json FROM snapshots ORDER BY ts DESC LIMIT ?", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Snapshot
	for rows.Next() {
		var tsStr, sysj, procj, connj string
		if err := rows.Scan(&tsStr, &sysj, &procj, &connj); err != nil {
			return nil, err
		}
		var sn models.Snapshot
		ts, _ := time.Parse(time.RFC3339, tsStr)
		sn.Timestamp = ts
		json.Unmarshal([]byte(sysj), &sn.System)
		json.Unmarshal([]byte(procj), &sn.Processes)
		json.Unmarshal([]byte(connj), &sn.Connections)
		out = append(out, sn)
	}
	return out, nil
}

func CloseSQLite(s *SQLiteStore) error {
	if s == nil || s.Db == nil {
		return nil
	}
	return s.Db.Close()
}
