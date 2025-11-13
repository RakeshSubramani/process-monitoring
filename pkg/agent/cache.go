package agent

import (
	"sync"
	"time"
)

// Snapshot holds the data
type Snapshot struct {
	Timestamp   time.Time     `json:"timestamp"`
	System      Metrics       `json:"system"`
	Processes   []ProcessInfo `json:"processes"`
	Initialized bool          `json:"initialized"`
}

// Cache stores latest snapshot (thread-safe)
type Cache struct {
	mu       sync.RWMutex
	latest   Snapshot
	interval time.Duration

	prevSent uint64
	prevRecv uint64
	prevTime time.Time
}

var globalCache *Cache

// StartCachePoller begins sampling metrics at given interval and keeping latest snapshot
func StartCachePoller(interval time.Duration) (*Cache, error) {
	c := &Cache{
		interval: interval,
	}
	globalCache = c

	// initial collect with zero prevTime
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			sys, sent, recv, now, _ := CollectSystem(c.prevSent, c.prevRecv, c.prevTime)
			c.prevSent = sent
			c.prevRecv = recv
			c.prevTime = now

			procs, _ := CollectTopProcesses(0)

			snap := Snapshot{
				Timestamp:   time.Now(),
				System:      sys,
				Processes:   procs,
				Initialized: true,
			}

			c.mu.Lock()
			c.latest = snap
			c.mu.Unlock()

			<-ticker.C
		}
	}()

	return c, nil
}

// GetLatest returns the latest snapshot (copy)
func GetLatest() Snapshot {
	if globalCache == nil {
		return Snapshot{}
	}
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	return globalCache.latest
}
