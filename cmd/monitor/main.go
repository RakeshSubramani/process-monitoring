package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RakeshSubramani/process-monitoring/pkg/agent"
	alerts "github.com/RakeshSubramani/process-monitoring/pkg/alert"
	server "github.com/RakeshSubramani/process-monitoring/pkg/api"
	models "github.com/RakeshSubramani/process-monitoring/pkg/model"
	storage "github.com/RakeshSubramani/process-monitoring/pkg/store"
	"github.com/RakeshSubramani/process-monitoring/pkg/ui"
)

func main() {
	interval := flag.Duration("interval", 10*time.Second, "sampling interval (e.g. 2s, 5s)")
	sqlitePath := flag.String("sqlite", "data/monitor.db", "sqlite db path")
	csvPath := flag.String("csv", "data/metrics.csv", "csv export path")
	addr := flag.String("addr", ":9090", "http listen address")
	enableUI := flag.Bool("ui", true, "enable terminal dashboard print")
	enableSQL := flag.Bool("sql", true, "enable sqlite persistence")
	enableCSV := flag.Bool("enablecsv", true, "enable csv persistence")
	flag.Parse()

	// ensure data folder exists
	_ = os.MkdirAll("data", 0755)
	fmt.Println("sqlitePath", sqlitePath)
	// initialize storage (only open db here; agent will create table)
	if *enableSQL {
		if err := storage.InitSQLite(*sqlitePath); err != nil {
			log.Fatalf("sqlite init: %v", err)
		}
	}

	if *enableCSV {
		if err := storage.InitCSV(*csvPath); err != nil {
			log.Fatalf("csv init: %v", err)
		}
	}
	path := *sqlitePath
	// Initialize SQLite store
	sqlStore, err := storage.NewSQLiteStore(path)
	if err != nil {
		log.Fatalf("failed to initialize sqlite: %v", err)
	}
	// start cache poller (collects metrics and persists)
	cache, err := agent.StartCachePoller(*interval, *enableSQL, *enableCSV, *sqlitePath, *csvPath)
	if err != nil {
		log.Fatalf("start cache poller: %v", err)
	}
	_ = cache

	// start alert manager
	alertMgr := alerts.NewManager()
	// example rule: CPU > 85%
	alertMgr.AddRule(alerts.Rule{
		Name:     "High CPU",
		Interval: 5 * time.Second,
		CheckFn: func(s models.Metrics) bool {
			return s.CPUPercent > 85.0
		},
		ActionFn: func(name string, s models.Metrics) {
			log.Printf("[ALERT] %s fired: CPU=%.2f%%", name, s.CPUPercent)
		},
	})
	go alertMgr.Start(2 * time.Second)

	// start http server (API + prometheus)
	srv := server.NewServer(*addr)
	go srv.Start()

	// start a simple terminal dashboard print (optional)
	if *enableUI {
		go ui.RunTUI(interval)
	}

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	// close stores
	if *enableSQL {
		storage.CloseSQLite(sqlStore)
	}
}
