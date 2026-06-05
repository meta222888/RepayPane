package ui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const resourcesCmd = `free -b 2>/dev/null | awk '/^Mem:/ {print "MEM_TOTAL="$2; print "MEM_USED="$3; print "MEM_FREE="$4}'
grep '^cpu ' /proc/stat | awk '{idle=$5+$6; total=0; for(i=2;i<=NF;i++) total+=$i; print "CPU_IDLE=" idle; print "CPU_TOTAL=" total}'
uptime
ps -eo comm,%cpu,%mem --sort=-%mem 2>/dev/null | head -8
`

func (a *App) showResourceUsage() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	title := i18n.T(i18n.KeyFeatResources)
	contentBox := container.NewVBox()
	scroll := container.NewVScroll(contentBox)
	refresh := newAccentButton(i18n.T(i18n.KeyRefresh), func() {
		loadResources(client, contentBox)
	})
	body := container.NewBorder(nil, refresh, nil, nil, scroll)
	showThemedFeature(a, title, fyne.NewSize(680, 520), body)
	loadResources(client, contentBox)
}

func loadResources(client *remote.Client, box *fyne.Container) {
	box.Objects = nil
	box.Add(widget.NewLabel(i18n.T(i18n.KeyFeatLoading)))
	box.Refresh()

	go func() {
		out, err := client.RunCombined(resourcesCmd + "\n" + `ps -eo pid,user,comm,%cpu,%mem --sort=-%mem 2>/dev/null | head -10`)
		fyne.Do(func() {
			box.Objects = nil
			if err != nil && strings.TrimSpace(out) == "" {
				box.Add(widget.NewLabel(err.Error()))
				box.Refresh()
				return
			}
			memTotal, memUsed := parseMem(out)
			cpuPct := parseCPU(out)
			uptimeLine := parseUptime(out)

			box.Add(widget.NewLabelWithStyle(i18n.T(i18n.KeyFeatResCPU), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			cpuBar := newUsageProgressBar()
			cpuBar.SetUsage(cpuPct)
			box.Add(cpuBar)
			box.Add(widget.NewLabel(i18n.Tf(i18n.KeyFeatResCPUPct, cpuPct)))

			box.Add(widget.NewLabelWithStyle(i18n.T(i18n.KeyFeatResMemory), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			memPct := 0.0
			if memTotal > 0 {
				memPct = float64(memUsed) / float64(memTotal) * 100
			}
			memBar := newUsageProgressBar()
			memBar.SetUsage(memPct)
			box.Add(memBar)
			box.Add(widget.NewLabel(i18n.Tf(i18n.KeyFeatResMemDetail, formatBytes(memUsed), formatBytes(memTotal), memPct)))

			if uptimeLine != "" {
				box.Add(widget.NewLabelWithStyle(i18n.T(i18n.KeyFeatResUptime), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
				uptimeLbl := widget.NewLabel(uptimeLine)
				uptimeLbl.Wrapping = fyne.TextWrapWord
				box.Add(paddedWidgetLabel(uptimeLbl))
			}

			box.Add(widget.NewLabelWithStyle(i18n.T(i18n.KeyFeatResProcesses), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			procLbl := widget.NewLabel(extractProcessTable(out))
			procLbl.Wrapping = fyne.TextWrapWord
			box.Add(paddedWidgetLabel(procLbl))
			box.Refresh()
		})
	}()
}

func parseMem(out string) (total, used int64) {
	reTotal := regexp.MustCompile(`MEM_TOTAL=(\d+)`)
	reUsed := regexp.MustCompile(`MEM_USED=(\d+)`)
	if m := reTotal.FindStringSubmatch(out); len(m) > 1 {
		total, _ = strconv.ParseInt(m[1], 10, 64)
	}
	if m := reUsed.FindStringSubmatch(out); len(m) > 1 {
		used, _ = strconv.ParseInt(m[1], 10, 64)
	}
	return total, used
}

func parseCPU(out string) float64 {
	reIdle := regexp.MustCompile(`CPU_IDLE=(\d+)`)
	reTotal := regexp.MustCompile(`CPU_TOTAL=(\d+)`)
	var idle, total int64
	if m := reIdle.FindStringSubmatch(out); len(m) > 1 {
		idle, _ = strconv.ParseInt(m[1], 10, 64)
	}
	if m := reTotal.FindStringSubmatch(out); len(m) > 1 {
		total, _ = strconv.ParseInt(m[1], 10, 64)
	}
	if total <= 0 {
		return 0
	}
	return float64(total-idle) / float64(total) * 100
}

func parseUptime(out string) string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "load average") || strings.HasPrefix(line, "up ") {
			return line
		}
	}
	return ""
}

func extractProcessTable(out string) string {
	lines := strings.Split(out, "\n")
	start := -1
	for i, line := range lines {
		if strings.Contains(line, "COMM") || strings.Contains(line, "%CPU") {
			start = i
			break
		}
	}
	if start < 0 {
		return strings.TrimSpace(out)
	}
	return strings.Join(lines[start:], "\n")
}

func formatBytes(n int64) string {
	if n < 1024 {
		return strconv.FormatInt(n, 10) + " B"
	}
	if n < 1024*1024 {
		return strconv.FormatFloat(float64(n)/1024, 'f', 1, 64) + " KB"
	}
	if n < 1024*1024*1024 {
		return strconv.FormatFloat(float64(n)/(1024*1024), 'f', 1, 64) + " MB"
	}
	return strconv.FormatFloat(float64(n)/(1024*1024*1024), 'f', 2, 64) + " GB"
}
