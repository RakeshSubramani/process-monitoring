package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMonitor starts the metric collection and HTTP server
func StartMonitor(port int, interval time.Duration) {
	fmt.Printf("[agent] Starting Go Process Monitor at :%d...\n", port)
	RegisterMetrics()
	go AutoCollect(interval, UpdateMetrics)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
