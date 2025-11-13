# 🖥️ Process Monitoring Terminal UI

A beautiful, interactive **terminal-based system monitor** written in **Go** — powered by [`tview`](https://github.com/rivo/tview), [`gopsutil`](https://github.com/shirou/gopsutil), and [`prometheus/client_golang`](https://github.com/prometheus/client_golang).

It provides **real-time insights** into:
- 🧠 CPU usage  
- 💾 Memory consumption  
- 🧱 Disk utilization  
- 🌐 Network bandwidth  
- 🔥 Top active processes (scrollable + searchable + sortable view)

---

## ✨ Features

✅ Live system resource monitoring  
✅ Realtime process list sorted by CPU, Memory, or PID  
✅ Scrollable process table (↑↓ navigation)  
✅ Searchable processes (`/` to search, `Enter` to apply, `Esc` to reset)  
✅ Color-coded metrics (CPU load: 🟩 normal, 🟨 warning, 🟥 high)  
✅ Kill process with **Ctrl + K** (safe shortcut)  
✅ SQLite persistence (`monitor.db` stores historical snapshots)  
✅ Prometheus metrics endpoint → `http://localhost:9090/metrics`  
✅ REST API endpoints for metrics, processes, and history  
✅ System health endpoint for readiness/liveness checks  


## 🌐 API Endpoints

Below are the available REST and Prometheus endpoints exposed by the monitor:

| **Endpoint** | **Description** | **Example Output** |
|---------------|-----------------|--------------------|
| `/metrics` | Prometheus metrics endpoint | Exposes Prometheus-compatible metrics for external scraping. |
| `/api/metrics` | Returns current CPU, memory, disk, and network metrics | ```json { "cpu_usage": [23.5, 15.4, 12.1], "memory_used_percent": 42.3, "disk_used_percent": 60.7, "network": { "bytes_sent": 14523312, "bytes_recv": 234534123 } } ``` |
| `/api/processes` | Returns list of top running processes | ```json [ { "pid": 1342, "name": "chrome", "cpu": 32.5, "mem": 4.5 }, { "pid": 2011, "name": "code", "cpu": 12.3, "mem": 2.1 } ] ``` |
| `/api/history` | Returns stored snapshots from SQLite (`monitor.db`) | ```json [ { "timestamp": "2025-11-13T18:32:00Z", "cpu": 22.1, "mem": 48.5 }, { "timestamp": "2025-11-13T18:33:00Z", "cpu": 25.4, "mem": 49.1 } ] ``` |
| `/api/health` | Health check endpoint | ```json { "status": "ok", "uptime": "1m23s" } ``` |


---

## 🧩 Tech Stack

| Component | Description |
|------------|-------------|
| **Go** | Core language |
| **tview** | Terminal UI framework |
| **gopsutil** | System metrics and process info |
| **sqlite3** | Lightweight database for metric snapshots |
| **prometheus/client_golang** | Prometheus exporter |
| **net/http** | REST API server |

---

## ⚙️ Installation

1. **Clone the repository**
   ```bash
   git clone git@github-personal:RakeshSubramani/process-monitoring.git
   cd process-monitoring/cmd/monitor

## ⚙️ Sample Image


<p align="center">
   <img width="800" height="539" alt="image" src="https://github.com/user-attachments/assets/a882c944-3f27-4053-a345-65ebbb62dbca" />
</p>
