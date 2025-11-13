package agent

import (
	"fmt"
	stdnet "net"
	"sync"
	"time"

	models "github.com/RakeshSubramani/process-monitoring/pkg/model"
	storage "github.com/RakeshSubramani/process-monitoring/pkg/store"
	gnet "github.com/shirou/gopsutil/v4/net"
)

type Cache struct {
	Mu       sync.RWMutex
	Latest   models.Snapshot
	Interval time.Duration
	PrevSent uint64
	PrevRecv uint64
	PrevTime time.Time
	SqlStore bool
	CsvStore bool
	Sql      *storage.SQLiteStore
	Csv      *storage.CSVStore
}

var globalCache *Cache
var domainCache = map[string]string{}
var domainLock sync.Mutex

func StartCachePoller(interval time.Duration, enableSQL bool, enableCSV bool, sqlitePath, csvPath string) (*Cache, error) {
	c := &Cache{Interval: interval, SqlStore: enableSQL, CsvStore: enableCSV}
	globalCache = c

	if enableSQL {
		s, err := storage.NewSQLiteStore(sqlitePath)
		if err != nil {
			return nil, fmt.Errorf("sqlite init: %w", err)
		}
		c.Sql = s
	}
	if enableCSV {
		s, err := storage.NewCSVStore(csvPath)
		if err != nil {
			return nil, fmt.Errorf("csv init: %w", err)
		}
		c.Csv = s
	}

	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			sys, sent, recv, now, _ := CollectSystem(c.PrevSent, c.PrevRecv, c.PrevTime)
			c.PrevSent = sent
			c.PrevRecv = recv
			c.PrevTime = now

			procs, _ := CollectTopProcesses(0)

			var conns []models.ConnInfo
			connsStats, err := gnet.Connections("inet")
			if err == nil {
				for _, con := range connsStats {
					if con.Status == "ESTABLISHED" && con.Raddr.IP != "" {
						d := resolveDomainCached(con.Raddr.IP)
						conns = append(conns, models.ConnInfo{
							Pid:    con.Pid,
							Local:  fmt.Sprintf("%s:%d", con.Laddr.IP, con.Laddr.Port),
							Remote: fmt.Sprintf("%s:%d", con.Raddr.IP, con.Raddr.Port),
							Domain: d,
							Status: con.Status,
						})
					}
				}
			}

			snap := models.Snapshot{
				Timestamp:   time.Now(),
				System:      sys,
				Processes:   procs,
				Connections: conns,
				Ready:       true,
			}

			c.Mu.Lock()
			c.Latest = snap
			c.Mu.Unlock()

			// persist
			if c.Sql != nil && c.SqlStore {
				_ = c.Sql.InsertSnapshot(snap)
			}
			if c.Csv != nil && c.CsvStore {
				_ = c.Csv.AppendSnapshotCSV(snap)
			}

			<-t.C
		}
	}()
	return c, nil
}

func GetLatest() models.Snapshot {
	if globalCache == nil {
		return models.Snapshot{}
	}
	globalCache.Mu.RLock()
	defer globalCache.Mu.RUnlock()
	return globalCache.Latest
}

func GetRecentSnapshots(limit int) ([]models.Snapshot, error) {
	if globalCache == nil || globalCache.Sql == nil {
		return nil, fmt.Errorf("sqlite not enabled")
	}
	return globalCache.Sql.GetRecentSnapshots(limit)
}

func resolveDomainCached(ip string) string {
	domainLock.Lock()
	defer domainLock.Unlock()
	if d, ok := domainCache[ip]; ok {
		return d
	}
	names, err := stdnet.LookupAddr(ip)
	if err == nil && len(names) > 0 {
		domainCache[ip] = names[0]
		return names[0]
	}
	domainCache[ip] = ip
	return ip
}
