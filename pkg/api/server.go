package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/RakeshSubramani/process-monitoring/pkg/agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	addr    string
	metrics *prometheus.GaugeVec
}

// NewServer returns a server that exposes endpoints and Prometheus metrics
func NewServer(addr string) *Server {
	g := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_monitor_metric_value",
			Help: "Generic metric (labels vary)",
		},
		[]string{"metric"},
	)
	prometheus.MustRegister(g)

	return &Server{
		addr:    addr,
		metrics: g,
	}
}

func (s *Server) Start() {
	// Prometheus handler
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/api/metrics", s.handleMetrics)
	http.HandleFunc("/api/processes", s.handleProcesses)
	http.HandleFunc("/api/health", s.handleHealth)
	http.HandleFunc("/api/version", s.handleVersion)

	log.Printf("API server listening on %s", s.addr)
	go func() {
		if err := http.ListenAndServe(s.addr, nil); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// optional: background exporter to update basic gauges (example)
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for range t.C {
			s.pushPrometheus()
		}
	}()
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	latest := agent.GetLatest()
	if !latest.Initialized {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, latest.System)
}

func (s *Server) handleProcesses(w http.ResponseWriter, r *http.Request) {
	latest := agent.GetLatest()
	if !latest.Initialized {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}

	limit := 0
	// optional query param ?limit=10
	q := r.URL.Query().Get("limit")
	if q != "" {
		// parse
		if v, err := strconv.Atoi(q); err == nil {
			limit = v
		}
	}

	procs := latest.Processes
	if limit > 0 && len(procs) > limit {
		procs = procs[:limit]
	}
	writeJSON(w, procs)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok", "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"version": "v1.0.0", "build": "local"})
}

func (s *Server) pushPrometheus() {
	latest := agent.GetLatest()
	if !latest.Initialized {
		return
	}
	l := latest.System
	// set a few metrics by name
	s.metrics.WithLabelValues("cpu_percent").Set(l.CPUPercent)
	s.metrics.WithLabelValues("memory_percent").Set(l.MemoryPercent)
	s.metrics.WithLabelValues("disk_used_mb").Set(l.DiskUsedMB)
	s.metrics.WithLabelValues("net_upload_mbps").Set(l.UploadSpeedMBs)
	s.metrics.WithLabelValues("net_download_mbps").Set(l.DownloadSpeedMBs)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
