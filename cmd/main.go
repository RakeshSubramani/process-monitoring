package main

import (
	"time"

	"github.com/RakeshSubramani/process-monitoring/pkg/agent"
)

func main() {
	agent.StartMonitor(9090, 5*time.Second)
}
