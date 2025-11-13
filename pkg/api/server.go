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

func NewServer(addr string) *Server {
	g := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_monitor_metric_value",
			Help: "metric value with label metric",
		},
		[]string{"metric"},
	)
	prometheus.MustRegister(g)
	return &Server{addr: addr, metrics: g}
}

func (s *Server) Start() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/metrics", s.handleMetrics)
	http.HandleFunc("/api/processes", s.handleProcesses)
	http.HandleFunc("/api/history", s.handleHistory)
	http.HandleFunc("/api/health", s.handleHealth)
	log.Printf("HTTP server listening on %s", s.addr)
	if err := http.ListenAndServe(s.addr, nil); err != nil {
		log.Fatalf("http listen: %v", err)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	latest := agent.GetLatest()
	if !latest.Ready {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	encodeJSON(w, latest.System)
}

func (s *Server) handleProcesses(w http.ResponseWriter, r *http.Request) {
	latest := agent.GetLatest()
	if !latest.Ready {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	limit := 0
	if q := r.URL.Query().Get("limit"); q != "" {
		if v, err := strconv.Atoi(q); err == nil {
			limit = v
		}
	}
	procs := latest.Processes
	if limit > 0 && len(procs) > limit {
		procs = procs[:limit]
	}
	encodeJSON(w, procs)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	n := 50
	if q := r.URL.Query().Get("n"); q != "" {
		if v, err := strconv.Atoi(q); err == nil {
			n = v
		}
	}
	snaps, err := agent.GetRecentSnapshots(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	encodeJSON(w, snaps)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	encodeJSON(w, map[string]interface{}{"status": "ok", "time": time.Now()})
}

func (s *Server) pushPrometheus() {
	latest := agent.GetLatest()
	if !latest.Ready {
		return
	}
	l := latest.System
	s.metrics.WithLabelValues("cpu_percent").Set(l.CPUPercent)
	s.metrics.WithLabelValues("memory_percent").Set(l.MemoryPercent)
	s.metrics.WithLabelValues("disk_used_mb").Set(l.DiskUsedMB)
	s.metrics.WithLabelValues("net_upload_mbps").Set(l.UploadSpeedMBs)
	s.metrics.WithLabelValues("net_download_mbps").Set(l.DownloadSpeedMBs)
}

func encodeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
