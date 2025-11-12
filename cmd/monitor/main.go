package main

import (
	"log"
	"time"

	"github.com/RakeshSubramani/process-monitoring/pkg/agent"
	"github.com/RakeshSubramani/process-monitoring/pkg/api"
	"github.com/RakeshSubramani/process-monitoring/pkg/ui"
)

func main() {
	// start cache poller: update every 5s, persist to sqlite + csv
	cache, err := agent.StartCachePoller(5*time.Second, true, true, "./data/snapshots.db", "./data/snapshots.csv")
	if err != nil {
		log.Fatalf("failed to start cache: %v", err)
	}
	_ = cache

	// register and start prometheus metrics (you should still call RegisterPromMetrics somewhere)
	agent.RegisterPromMetrics()

	// start API server (includes /metrics)
	go func() {
		srv := api.NewServer(":9090")	
		srv.Run()
	}()

	// start UI (blocks)
	ui.RunTUI(5 * time.Second)
}
