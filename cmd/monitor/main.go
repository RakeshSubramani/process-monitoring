package main

import (
	"flag"
	"log"
	"time"

	"github.com/RakeshSubramani/process-monitoring/pkg/agent"
	server "github.com/RakeshSubramani/process-monitoring/pkg/api"
	"github.com/RakeshSubramani/process-monitoring/pkg/ui"
)

func main() {
	interval := flag.Duration("interval", 2*time.Second, "sampling interval")
	enableUI := flag.Bool("ui", true, "enable terminal UI dashboard")
	addr := flag.String("addr", ":9090", "http listen address")
	flag.Parse()

	// start cache poller
	_, err := agent.StartCachePoller(*interval)
	if err != nil {
		log.Fatalf("failed to start cache: %v", err)
	}

	// start server
	s := server.NewServer(*addr)
	s.Start()

	// start UI (optional)
	if *enableUI {
		go ui.RunTUI(*interval)
	}

	// block forever
	select {}
}
