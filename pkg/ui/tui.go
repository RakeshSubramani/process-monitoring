package ui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

func RunTUI(refresh time.Duration) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	header := widgets.NewParagraph()
	header.Title = "💻 System Overview"
	header.TextStyle.Fg = ui.ColorYellow
	header.BorderStyle.Fg = ui.ColorCyan
	header.SetRect(0, 0, 120, 5)

	table := widgets.NewTable()
	table.Title = "🔥 Processes (↑↓ Scroll | / Search | s Sort | k Kill | q Quit)"
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.BorderStyle.Fg = ui.ColorGreen
	table.FillRow = true
	table.RowSeparator = false
	table.SetRect(0, 5, 120, 30)

	offset := 0
	maxVisible := 18
	sortKey := "cpu"
	filter := ""

	update := func() {
		// ─── System Info ────────────────────────────────────────────────
		cpuPercents, _ := cpu.Percent(0, true)
		memStats, _ := mem.VirtualMemory()
		diskStats, _ := disk.Usage("/")
		netStats, _ := net.IOCounters(false)

		cpuText := ""
		for i, c := range cpuPercents {
			cpuText += fmt.Sprintf("CPU%d: %.1f%%  ", i, c)
		}

		header.Text = fmt.Sprintf(
			"%s\nMEM: %.1f%% (%.1f GB / %.1f GB)\nDISK: %.1f%% (%.1f GB / %.1f GB)\nNET: ↑ %.1f MB ↓ %.1f MB",
			cpuText,
			memStats.UsedPercent,
			float64(memStats.Used)/1024/1024/1024, float64(memStats.Total)/1024/1024/1024,
			diskStats.UsedPercent,
			float64(diskStats.Used)/1024/1024/1024, float64(diskStats.Total)/1024/1024/1024,
			float64(netStats[0].BytesSent)/1024/1024,
			float64(netStats[0].BytesRecv)/1024/1024,
		)

		// ─── Process Table ──────────────────────────────────────────────
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
			if filter != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
				continue
			}
			cpuPct, _ := p.CPUPercent()
			memPct, _ := p.MemoryPercent()
			infos = append(infos, pInfo{PID: p.Pid, Name: name, CPU: cpuPct, MEM: memPct})
		}

		switch sortKey {
		case "cpu":
			sort.Slice(infos, func(i, j int) bool { return infos[i].CPU > infos[j].CPU })
		case "mem":
			sort.Slice(infos, func(i, j int) bool { return infos[i].MEM > infos[j].MEM })
		case "pid":
			sort.Slice(infos, func(i, j int) bool { return infos[i].PID < infos[j].PID })
		}

		rows := [][]string{{"PID", "NAME", "CPU (%)", "MEM (%)"}}
		for _, p := range infos {
			color := ui.ColorGreen
			switch {
			case p.CPU > 70:
				color = ui.ColorRed
			case p.CPU > 30:
				color = ui.ColorYellow
			}
			rows = append(rows, []string{
				fmt.Sprintf("%d", p.PID),
				p.Name,
				fmt.Sprintf("[%5.2f](fg:%s)", p.CPU, colorToString(color)),
				fmt.Sprintf("%.2f", p.MEM),
			})
		}

		start := offset + 1
		end := offset + maxVisible
		if end > len(rows) {
			end = len(rows)
		}
		if start < len(rows) {
			table.Rows = append(rows[:1], rows[start:end]...)
		} else {
			table.Rows = rows[:1]
		}
	}

	update()
	ui.Render(header, table)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	searchMode := false
	searchText := ""

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				ui.Close()
				fmt.Println("👋 Exiting process monitor. Bye!")
				os.Exit(0)

			case "<Down>":
				offset++
				update()

			case "<Up>":
				if offset > 0 {
					offset--
				}
				update()

			case "s":
				switch sortKey {
				case "cpu":
					sortKey = "mem"
				case "mem":
					sortKey = "pid"
				default:
					sortKey = "cpu"
				}
				update()

			case "/":
				searchMode = true
				searchText = ""
				header.Title = "🔍 Search mode: type to filter (Enter to apply, Esc to cancel)"
				ui.Render(header)

			case "<Escape>":
				if searchMode {
					searchMode = false
					filter = ""
					header.Title = "💻 System Overview"
					update()
				}

			case "<Enter>":
				if searchMode {
					filter = searchText
					searchMode = false
					header.Title = "💻 System Overview"
					update()
				}

			case "k":
				if len(table.Rows) > 1 {
					pidStr := table.Rows[1][0] // top visible row
					pid, _ := strconv.Atoi(pidStr)
					exec.Command("kill", "-9", fmt.Sprint(pid)).Run()
					update()
				}

			default:
				if searchMode && len(e.ID) == 1 {
					searchText += e.ID
					header.Text = fmt.Sprintf("Filter: %s", searchText)
					ui.Render(header)
				}
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
