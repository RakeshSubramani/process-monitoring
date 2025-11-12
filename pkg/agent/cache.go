package agent

import (
	"fmt"
	stdnet "net"
	"sync"
	"time"

	gnet "github.com/shirou/gopsutil/v4/net"
)

// Snapshot holds a single point-in-time snapshot of system + processes
type Snapshot struct {
	Timestamp time.Time     `json:"timestamp"`
	System    Metrics       `json:"system"`
	Processes []ProcessInfo `json:"processes"`
	// Connections optional: list of active connections with domain (optional)
	Connections []ConnInfo `json:"connections,omitempty"`
}

// ConnInfo minimal connection info
type ConnInfo struct {
	Pid    int32  `json:"pid"`
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Domain string `json:"domain,omitempty"`
	Status string `json:"status"`
}

// Cache is a thread-safe in-memory snapshot cache
type Cache struct {
	mu       sync.RWMutex
	latest   Snapshot
	interval time.Duration

	// last net totals for bandwidth calc
	prevSent uint64
	prevRecv uint64
	prevTime time.Time
}

var globalCache *Cache

// StartCachePoller creates and starts the background poller.
// storeToSQL / storeToCSV flags enable historical storage.
func StartCachePoller(interval time.Duration, storeToSQL bool, storeToCSV bool, sqlitePath string, csvPath string) (*Cache, error) {
	c := &Cache{
		interval: interval,
	}
	globalCache = c

	// init stores if requested
	var sqlStore *SQLiteStore
	var csvStore *CSVStore
	var err error
	if storeToSQL {
		sqlStore, err = NewSQLiteStore(sqlitePath)
		if err != nil {
			return nil, err
		}
	}
	if storeToCSV {
		csvStore, err = NewCSVStore(csvPath)
		if err != nil {
			return nil, err
		}
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			// collect snapshot
			snap := c.collectSnapshot()
			// write to cache
			c.mu.Lock()
			c.latest = snap
			c.mu.Unlock()

			// persist if enabled
			if sqlStore != nil {
				_ = sqlStore.InsertSnapshot(snap) // ignore error here; you can log
			}
			if csvStore != nil {
				_ = csvStore.AppendSnapshotCSV(snap)
			}

			<-ticker.C
		}
	}()

	return c, nil
}

// GetLatest returns the latest snapshot
func GetLatest() Snapshot {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	return globalCache.latest
}

// collectSnapshot uses your CollectSystem / CollectTopProcesses / connections and composes Snapshot
func (c *Cache) collectSnapshot() Snapshot {
	// ---- System metrics ----
	sys, sent, recv, now, _ := CollectSystem(c.prevSent, c.prevRecv, c.prevTime)
	c.prevSent = sent
	c.prevRecv = recv
	c.prevTime = now

	// ---- Process list ----
	procs, _ := CollectTopProcesses(0) // 0 → collect all processes
	// fmt.Println("procs", procs)
	// ---- Network connections ----
	var conns []ConnInfo
	connsStats, err := gnet.Connections("inet")
	if err == nil {
		for _, con := range connsStats {
			if con.Status == "ESTABLISHED" && con.Raddr.IP != "" {
				domain := resolveDomainCached(con.Raddr.IP, domainCache)
				conns = append(conns, ConnInfo{
					Pid:    con.Pid,
					Local:  fmt.Sprintf("%s:%d", con.Laddr.IP, con.Laddr.Port),
					Remote: fmt.Sprintf("%s:%d", con.Raddr.IP, con.Raddr.Port),
					Domain: domain,
					Status: con.Status,
				})
			}
		}
	}

	return Snapshot{
		Timestamp:   time.Now(),
		System:      sys,
		Processes:   procs,
		Connections: conns,
	}
}

var domainCache = make(map[string]string)
var domainLock sync.Mutex

func resolveDomainCached(ip string, cache map[string]string) string {
	domainLock.Lock()
	defer domainLock.Unlock()

	if domain, ok := cache[ip]; ok {
		return domain
	}

	names, err := stdnet.LookupAddr(ip)
	if err == nil && len(names) > 0 {
		cache[ip] = names[0]
		return names[0]
	}

	cache[ip] = ip
	return ip
}
