# 🖥️ Process Monitoring Terminal UI

A beautiful, interactive **terminal-based system monitor** written in **Go** — built using [`tview`](https://github.com/rivo/tview) and [`gopsutil`](https://github.com/shirou/gopsutil).

It provides **real-time insights** into:
- 🧠 CPU usage  
- 💾 Memory consumption  
- 🧱 Disk utilization  
- 🌐 Network bandwidth  
- 🔥 Top active processes (scrollable view)

---

## ✨ Features

✅ Live system resource monitoring  
✅ Realtime process list sorted by CPU usage  
✅ Scrollable process table (top 10 visible, scroll for more)  
✅ Color-coded metrics (CPU load indicators)  
✅ SQLite persistence for storing snapshots  
✅ Prometheus metrics endpoint → `http://localhost:9090/metrics`  
✅ REST API endpoints for metrics and processes  

---

## 🧩 Tech Stack

| Component | Description |
|------------|--------------|
| **Go** | Core programming language |
| **tview** | Terminal UI framework |
| **gopsutil** | System metrics & process info |
| **sqlite3** | Lightweight database for snapshots |
| **Prometheus client** | Exposes metrics for external scraping |

---

## ⚙️ Installation

1. **Clone the repository**
   ```bash
   git clone git@github-personal:RakeshSubramani/process-monitoring.git
   cd process-monitoring/cmd/monitor
    