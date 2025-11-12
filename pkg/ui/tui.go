package ui

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// RunTUI runs a compact scrollable live process monitor
func RunTUI(refresh time.Duration) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	header := widgets.NewParagraph()
	header.Title = "💻 System Metrics"
	header.TextStyle.Fg = ui.ColorYellow
	header.BorderStyle.Fg = ui.ColorCyan
	header.SetRect(0, 0, 100, 3) // ⬅️ Fixed height = 3 lines

	table := widgets.NewTable()
	table.Title = "🔥 Processes (Scroll ↑↓)"
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.BorderStyle.Fg = ui.ColorGreen
	table.FillRow = true
	table.RowSeparator = false
	table.SetRect(0, 3, 100, 18) // ⬅️ Fixed height = 15 lines (compact)

	offset := 0
	maxVisible := 10 // show 10 process rows max

	update := func() {
		cpuPercent, _ := cpu.Percent(0, false)
		memStats, _ := mem.VirtualMemory()
		diskStats, _ := disk.Usage("/")
		netStats, _ := gnet.IOCounters(false)

		header.Text = fmt.Sprintf(
			"CPU: %.2f%% | MEM: %.1fMB (%.1f%%) | DISK: %.1fMB | UP: %.1fMB | DOWN: %.1fMB",
			cpuPercent[0],
			float64(memStats.Used)/1024/1024, memStats.UsedPercent,
			float64(diskStats.Used)/1024/1024,
			float64(netStats[0].BytesSent)/1024/1024,
			float64(netStats[0].BytesRecv)/1024/1024,
		)

		procs, _ := process.Processes()
		type pInfo struct {
			PID  int32
			Name string
			CPU  float64
			MEM  float32
		}

		var infos []pInfo
		for _, p := range procs {
			name, _ := p.Name()
			cpuPct, _ := p.CPUPercent()
			memPct, _ := p.MemoryPercent()
			if name != "" {
				infos = append(infos, pInfo{PID: p.Pid, Name: name, CPU: cpuPct, MEM: memPct})
			}
		}
		sort.Slice(infos, func(i, j int) bool { return infos[i].CPU > infos[j].CPU })

		rows := [][]string{{"PID", "NAME", "CPU (%)", "MEM (%)"}}
		for _, p := range infos {
			color := ui.ColorGreen
			switch {
			case p.CPU > 50:
				color = ui.ColorRed
			case p.CPU > 20:
				color = ui.ColorYellow
			}
			row := []string{
				strconv.Itoa(int(p.PID)),
				p.Name,
				fmt.Sprintf("[%6.2f](fg:%s)", p.CPU, colorToString(color)),
				fmt.Sprintf("%.2f", p.MEM),
			}
			rows = append(rows, row)
		}

		// Scrolling slice
		start := offset + 1
		end := offset + maxVisible
		if start >= len(rows) {
			start = len(rows) - maxVisible
			if start < 1 {
				start = 1
			}
			end = len(rows)
		}
		if end > len(rows) {
			end = len(rows)
		}
		table.Rows = append(rows[:1], rows[start:end]...) // header + visible rows
	}

	update()
	ui.Render(header, table)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Down>":
				offset++
				update()
			case "<Up>":
				if offset > 0 {
					offset--
				}
				update()
			case "<PageDown>":
				offset += 5
				update()
			case "<PageUp>":
				if offset >= 5 {
					offset -= 5
				} else {
					offset = 0
				}
				update()
			}
			ui.Render(header, table)
		case <-ticker.C:
			update()
			ui.Render(header, table)
		}
	}
}

func colorToString(c ui.Color) string {
	switch c {
	case ui.ColorRed:
		return "red"
	case ui.ColorYellow:
		return "yellow"
	case ui.ColorGreen:
		return "green"
	default:
		return "white"
	}
}
