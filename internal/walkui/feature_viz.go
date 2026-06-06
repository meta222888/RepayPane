package walkui

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const netIfaceCmd = `awk 'NR>2 {
  iface=$1; gsub(":", "", iface)
  if (iface=="lo") next
  print iface "\t" $2 "\t" $10
}' /proc/net/dev 2>/dev/null`

const netRouteCmd = `(ip -o route show 2>/dev/null || ip route 2>/dev/null || route -n 2>/dev/null) | head -12`

const netPortsCmd = `echo "=== Listening Ports ==="
(ss -tulnp 2>/dev/null || netstat -tulnp 2>/dev/null || netstat -tuln 2>/dev/null)
`

const resourcesCmd = `free -b 2>/dev/null | awk '/^Mem:/ {print "MEM_TOTAL="$2; print "MEM_USED="$3; print "MEM_FREE="$4}'
grep '^cpu ' /proc/stat | awk '{idle=$5+$6; total=0; for(i=2;i<=NF;i++) total+=$i; print "CPU_IDLE=" idle; print "CPU_TOTAL=" total}'
uptime
ps -eo pid,user,comm,%cpu,%mem --sort=-%mem 2>/dev/null | head -10
`

type diskRow struct {
	source string
	size   string
	used   string
	avail  string
	pcent  string
	mount  string
}

type netIfaceStat struct {
	name string
	rx   int64
	tx   int64
}

func (a *App) showDiskSpace() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	var dlg *walk.Dialog
	var scroll *walk.ScrollView
	var content *walk.Composite
	var loading *walk.Label

	refresh := func() {
		if loading != nil {
			loading.SetVisible(true)
		}
		clearComposite(content)
		go func() {
			out, err := client.RunCombined(`df -hP 2>/dev/null || df -h 2>/dev/null`)
			a.syncUI(func() {
				if loading != nil {
					loading.SetVisible(false)
				}
				clearComposite(content)
				rows := parseDfOutput(out)
				if len(rows) == 0 {
					msg := i18n.T(i18n.KeyFeatNoData)
					if err != nil && strings.TrimSpace(out) == "" {
						msg = err.Error()
					} else if t := strings.TrimSpace(out); t != "" {
						msg = t
					}
					addLabel(content, msg)
					return
				}
				sortDiskRowsByUsed(rows)
				for _, row := range rows {
					addDiskCard(content, row)
				}
			})
		}()
	}

	if err := (Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyFeatDisk),
		MinSize:  Size{640, 480},
		Layout:   VBox{},
		Children: []Widget{
			Label{AssignTo: &loading, Text: i18n.T(i18n.KeyFeatReading)},
			ScrollView{
				AssignTo: &scroll,
				Layout:   VBox{},
				Children: []Widget{
					Composite{AssignTo: &content, Layout: VBox{}},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyRefresh), OnClicked: refresh},
					PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}).Create(a.mw); err != nil {
		return
	}
	a.ownDialog(dlg)
	refresh()
	dlg.Run()
}

func (a *App) showResourceUsage() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	var dlg *walk.Dialog
	var content *walk.Composite
	var loading *walk.Label

	refresh := func() {
		if loading != nil {
			loading.SetVisible(true)
		}
		clearComposite(content)
		go func() {
			out, err := client.RunCombined(resourcesCmd + "\n" + `ps -eo pid,user,comm,%cpu,%mem --sort=-%mem 2>/dev/null | head -10`)
			a.syncUI(func() {
				if loading != nil {
					loading.SetVisible(false)
				}
				clearComposite(content)
				if err != nil && strings.TrimSpace(out) == "" {
					addLabel(content, err.Error())
					return
				}
				memTotal, memUsed := parseMem(out)
				cpuPct := parseCPU(out)
				uptimeLine := parseUptime(out)

				addTitleLabel(content, i18n.T(i18n.KeyFeatResCPU))
				addUsageBar(content, cpuPct)
				addLabel(content, i18n.Tf(i18n.KeyFeatResCPUPct, cpuPct))

				memPct := 0.0
				if memTotal > 0 {
					memPct = float64(memUsed) / float64(memTotal) * 100
				}
				addTitleLabel(content, i18n.T(i18n.KeyFeatResMemory))
				addUsageBar(content, memPct)
				addLabel(content, i18n.Tf(i18n.KeyFeatResMemDetail, formatBytes(memUsed), formatBytes(memTotal), memPct))

				if uptimeLine != "" {
					addTitleLabel(content, i18n.T(i18n.KeyFeatResUptime))
					addLabel(content, uptimeLine)
				}
				addTitleLabel(content, i18n.T(i18n.KeyFeatResProcesses))
				addReadOnlyTextBlock(content, extractProcessTable(out))
			})
		}()
	}

	if err := (Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyFeatResources),
		MinSize:  Size{680, 520},
		Layout:   VBox{},
		Children: []Widget{
			Label{AssignTo: &loading, Text: i18n.T(i18n.KeyFeatLoading)},
			ScrollView{
				Layout: VBox{},
				Children: []Widget{
					Composite{AssignTo: &content, Layout: VBox{}},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyRefresh), OnClicked: refresh},
					PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}).Create(a.mw); err != nil {
		return
	}
	a.ownDialog(dlg)
	refresh()
	dlg.Run()
}

func (a *App) showNetworkInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	var dlg *walk.Dialog
	var trafficBox *walk.Composite
	var summaryLbl *walk.Label
	var bootLbl *walk.Label
	var portsEdit *walk.TextEdit
	var autoCheck *walk.CheckBox
	var stopCh chan struct{}
	panel := (*netTrafficPanel)(nil)

	refreshTraffic := func(full bool) {
		go func() {
			ifaceOut, ifaceErr := client.RunCombined(netIfaceCmd)
			var routeOut string
			if full {
				routeOut, _ = client.RunCombined(netRouteCmd)
			}
			a.syncUI(func() {
				if panel == nil {
					return
				}
				stats := parseNetIfaces(ifaceOut)
				if len(stats) == 0 {
					msg := i18n.T(i18n.KeyFeatNoData)
					if ifaceErr != nil && strings.TrimSpace(ifaceOut) == "" {
						msg = ifaceErr.Error()
					}
					if full {
						clearComposite(trafficBox)
						addLabel(trafficBox, msg)
					}
					return
				}
				var bootRx, bootTx int64
				for _, s := range stats {
					bootRx += s.rx
					bootTx += s.tx
				}
				if bootLbl != nil {
					bootLbl.SetText(i18n.Tf(i18n.KeyFeatNetBootTotals, formatBytes(bootRx), formatBytes(bootTx)))
				}
				if full {
					panel.update(stats, routeOut, true)
				} else {
					panel.updateInPlace(stats)
				}
			})
		}()
	}

	refreshPorts := func() {
		go func() {
			out, err := client.RunCombined(netPortsCmd)
			a.syncUI(func() {
				if portsEdit == nil {
					return
				}
				if err != nil && strings.TrimSpace(out) == "" {
					setMultilineText(portsEdit, err.Error())
					return
				}
				setMultilineText(portsEdit, strings.TrimSpace(out))
			})
		}()
	}

	if err := (Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyFeatNetwork),
		MinSize:  Size{720, 560},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{AssignTo: &summaryLbl, Text: i18n.T(i18n.KeyFeatNetRatePending)},
							Label{AssignTo: &bootLbl, Text: ""},
						},
					},
					HSpacer{},
					CheckBox{
						AssignTo: &autoCheck,
						Text:     i18n.T(i18n.KeyFeatNetAutoRefresh),
						Checked:  true,
						OnCheckedChanged: func() {
							if stopCh != nil {
								close(stopCh)
								stopCh = nil
							}
							if autoCheck.Checked() {
								stopCh = make(chan struct{})
								go a.netAutoRefreshLoop(client, panel, summaryLbl, autoCheck, stopCh)
							} else if summaryLbl != nil {
								summaryLbl.SetText(i18n.T(i18n.KeyFeatNetRateOff))
							}
						},
					},
				},
			},
			ScrollView{
				Layout: VBox{},
				Children: []Widget{
					Composite{AssignTo: &trafficBox, Layout: VBox{}},
				},
			},
			Label{Text: i18n.T(i18n.KeyFeatNetPorts), Font: Font{Bold: true}},
			TextEdit{AssignTo: &portsEdit, ReadOnly: true, VScroll: true, MinSize: Size{0, 120}, Font: Font{Family: "Consolas", PointSize: 9}},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyFeatRefreshTraffic), OnClicked: func() { refreshTraffic(true) }},
					PushButton{Text: i18n.T(i18n.KeyFeatRefreshPorts), OnClicked: refreshPorts},
					PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}).Create(a.mw); err != nil {
		return
	}
	panel = newNetTrafficPanel(trafficBox)
	a.ownDialog(dlg)
	dlg.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if stopCh != nil {
			close(stopCh)
			stopCh = nil
		}
	})
	stopCh = make(chan struct{})
	go a.netAutoRefreshLoop(client, panel, summaryLbl, autoCheck, stopCh)
	refreshTraffic(true)
	refreshPorts()
	dlg.Run()
}

func (a *App) netAutoRefreshLoop(client *remote.Client, panel *netTrafficPanel, summary *walk.Label, auto *walk.CheckBox, stopCh chan struct{}) {
	prev := map[string]netIfaceStat{}
	var prevAt time.Time
	for {
		select {
		case <-stopCh:
			return
		case <-time.After(5 * time.Second):
		}
		if auto == nil || !auto.Checked() {
			continue
		}
		out, err := client.RunCombined(netIfaceCmd)
		a.syncUI(func() {
			stats := parseNetIfaces(out)
			if len(stats) == 0 {
				if err != nil && summary != nil {
					summary.SetText(err.Error())
				}
				return
			}
			now := time.Now()
			var rxTotal, txTotal float64
			if !prevAt.IsZero() {
				dt := now.Sub(prevAt).Seconds()
				if dt > 0 {
					for _, s := range stats {
						if p, ok := prev[s.name]; ok {
							rxTotal += float64(s.rx-p.rx) / dt
							txTotal += float64(s.tx-p.tx) / dt
						}
					}
					if summary != nil {
						summary.SetText(i18n.Tf(i18n.KeyFeatNetRate, formatBytes(int64(rxTotal))+"/s", formatBytes(int64(txTotal))+"/s"))
					}
				}
			}
			for _, s := range stats {
				prev[s.name] = s
			}
			prevAt = now
			if panel != nil {
				panel.updateInPlace(stats)
			}
		})
	}
}

func addDiskCard(parent *walk.Composite, row diskRow) {
	pct, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(row.pcent), "%"), 64)
	addTitleLabel(parent, row.mount)
	addUsageBar(parent, pct)
	addLabel(parent, i18n.Tf(i18n.KeyFeatDiskDetail, row.used, row.size, row.avail, row.pcent))
}

func addUsageBar(parent *walk.Composite, pct float64) {
	if parent == nil {
		return
	}
	pb, err := walk.NewProgressBar(parent)
	if err != nil {
		return
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	pb.SetValue(int(pct))
	parent.Children().Add(pb)
}

func addReadOnlyTextBlock(parent *walk.Composite, text string) {
	if parent == nil {
		return
	}
	te, err := walk.NewTextEdit(parent)
	if err != nil {
		return
	}
	_ = te.SetReadOnly(true)
	font, _ := walk.NewFont("Consolas", 9, 0)
	if font != nil {
		te.SetFont(font)
	}
	setMultilineText(te, text)
}

func addLabel(parent *walk.Composite, text string) {
	if parent == nil {
		return
	}
	lbl, err := walk.NewLabel(parent)
	if err != nil {
		return
	}
	lbl.SetText(text)
	parent.Children().Add(lbl)
}

func addTitleLabel(parent *walk.Composite, text string) {
	if parent == nil {
		return
	}
	lbl, err := walk.NewLabel(parent)
	if err != nil {
		return
	}
	lbl.SetText(text)
	font, _ := walk.NewFont("Segoe UI", 9, walk.FontBold)
	if font != nil {
		lbl.SetFont(font)
	}
	parent.Children().Add(lbl)
}

func clearComposite(c *walk.Composite) {
	if c == nil {
		return
	}
	for c.Children().Len() > 0 {
		item := c.Children().At(0)
		c.Children().Remove(item)
		item.Dispose()
	}
}

func parseDfOutput(out string) []diskRow {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	rows := make([]diskRow, 0, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i == 0 && strings.Contains(strings.ToLower(line), "filesystem") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		rows = append(rows, diskRow{
			source: fields[0],
			size:   fields[1],
			used:   fields[2],
			avail:  fields[3],
			pcent:  fields[4],
			mount:  strings.Join(fields[5:], " "),
		})
	}
	return rows
}

func sortDiskRowsByUsed(rows []diskRow) {
	sort.Slice(rows, func(i, j int) bool {
		ui := parseHumanBytes(rows[i].used)
		uj := parseHumanBytes(rows[j].used)
		if ui != uj {
			return ui > uj
		}
		return rows[i].mount < rows[j].mount
	})
}

func parseHumanBytes(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	mult := int64(1)
	if len(s) > 1 {
		switch s[len(s)-1] {
		case 'k', 'K':
			mult = 1024
			s = strings.TrimSpace(s[:len(s)-1])
		case 'm', 'M':
			mult = 1024 * 1024
			s = strings.TrimSpace(s[:len(s)-1])
		case 'g', 'G':
			mult = 1024 * 1024 * 1024
			s = strings.TrimSpace(s[:len(s)-1])
		case 't', 'T':
			mult = 1024 * 1024 * 1024 * 1024
			s = strings.TrimSpace(s[:len(s)-1])
		}
	}
	if len(s) > 1 && (s[len(s)-1] == 'i' || s[len(s)-1] == 'B') {
		s = strings.TrimSpace(s[:len(s)-1])
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f * float64(mult))
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

func parseNetIfaces(out string) []netIfaceStat {
	var stats []netIfaceStat
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		rx, _ := strconv.ParseInt(fields[1], 10, 64)
		tx, _ := strconv.ParseInt(fields[2], 10, 64)
		stats = append(stats, netIfaceStat{name: fields[0], rx: rx, tx: tx})
	}
	return stats
}
