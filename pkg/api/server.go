package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/RakeshSubramani/process-monitoring/pkg/agent"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) Run() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/metrics", s.handleMetrics)
	http.HandleFunc("/api/processes", s.handleProcesses)
	log.Printf("API server listening on %s", s.addr)
	log.Fatal(http.ListenAndServe(s.addr, nil))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	snap := agent.GetLatest()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp":   snap.Timestamp,
		"system":      snap.System,
		"connections": snap.Connections,
	})
}

func (s *Server) handleProcesses(w http.ResponseWriter, r *http.Request) {
	snap := agent.GetLatest()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap.Processes)
}
