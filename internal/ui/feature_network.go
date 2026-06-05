package ui

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

type netIfaceStat struct {
	name string
	rx   int64
	tx   int64
}

type netTrafficState struct {
	box      *fyne.Container
	prev     map[string]netIfaceStat
	prevAt   time.Time
	showRate bool
}

func (a *App) showNetworkInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	title := i18n.T(i18n.KeyFeatNetwork)
	trafficState := &netTrafficState{box: container.NewVBox()}
	trafficScroll := container.NewVScroll(trafficState.box)
	setPorts, portsScroll := scrollLineList()

	autoRefresh := widget.NewCheck(i18n.T(i18n.KeyFeatNetAutoRefresh), nil)
	autoRefresh.SetChecked(false)

	var stopCh chan struct{}
	autoRefresh.OnChanged = func(checked bool) {
		if stopCh != nil {
			close(stopCh)
			stopCh = nil
		}
		trafficState.showRate = checked
		trafficState.prev = nil
		trafficState.prevAt = time.Time{}
		if checked {
			stopCh = make(chan struct{})
			go netTrafficLoop(client, trafficState, stopCh)
		}
	}

	refreshTraffic := newAccentButton(i18n.T(i18n.KeyFeatRefreshTraffic), func() {
		loadNetTraffic(client, trafficState)
	})
	refreshPorts := newAccentButton(i18n.T(i18n.KeyFeatRefreshPorts), func() {
		loadNetPorts(client, setPorts)
	})

	trafficHeader := container.NewHBox(
		widget.NewLabelWithStyle(i18n.T(i18n.KeyFeatNetTraffic), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		autoRefresh,
	)
	trafficBox := container.NewBorder(trafficHeader, refreshTraffic, nil, nil, trafficScroll)

	portsHeader := widget.NewLabelWithStyle(i18n.T(i18n.KeyFeatNetPorts), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	portsBox := container.NewBorder(portsHeader, refreshPorts, nil, nil, portsScroll)

	split := container.NewVSplit(trafficBox, portsBox)
	split.SetOffset(0.5)
	dlg := showThemedFeature(a, title, fyne.NewSize(720, 560), split)
	dlg.SetOnClose(func() {
		if stopCh != nil {
			close(stopCh)
			stopCh = nil
		}
	})

	loadNetTraffic(client, trafficState)
	loadNetPorts(client, setPorts)
}

func loadNetTraffic(client *remote.Client, state *netTrafficState) {
	state.box.Objects = nil
	state.box.Add(featureLoadingLabel())
	state.box.Refresh()

	go func() {
		ifaceOut, ifaceErr := client.RunCombined(netIfaceCmd)
		routeOut, routeErr := client.RunCombined(netRouteCmd)
		fyne.Do(func() {
			renderNetTraffic(state, ifaceOut, routeOut, ifaceErr, routeErr)
		})
	}()
}

func renderNetTraffic(state *netTrafficState, ifaceOut, routeOut string, ifaceErr, routeErr error) {
	state.box.Objects = nil

	stats := filterNetIfaces(parseNetIfaces(ifaceOut))
	if len(stats) == 0 {
		if ifaceErr != nil && strings.TrimSpace(ifaceOut) == "" {
			state.box.Add(widget.NewLabel(ifaceErr.Error()))
		} else {
			state.box.Add(widget.NewLabel(i18n.T(i18n.KeyFeatNoData)))
		}
	} else {
		now := time.Now()
		elapsed := now.Sub(state.prevAt).Seconds()
		if state.prevAt.IsZero() || elapsed <= 0 {
			elapsed = 0
		}

		hint := widget.NewLabel(i18n.T(i18n.KeyFeatNetSinceBoot))
		hint.Importance = widget.LowImportance
		state.box.Add(hint)

		for _, stat := range stats {
			var rxRate, txRate float64
			if state.showRate && state.prev != nil && elapsed > 0 {
				if prev, ok := state.prev[stat.name]; ok {
					rxRate = float64(stat.rx-prev.rx) / elapsed
					txRate = float64(stat.tx-prev.tx) / elapsed
					if rxRate < 0 {
						rxRate = 0
					}
					if txRate < 0 {
						txRate = 0
					}
				}
			}
			state.box.Add(netIfaceCard(stat, rxRate, txRate, state.showRate && elapsed > 0))
		}

		if state.showRate {
			next := make(map[string]netIfaceStat, len(stats))
			for _, stat := range stats {
				next[stat.name] = stat
			}
			state.prev = next
			state.prevAt = now
		}
	}

	routes := parseNetRoutes(routeOut)
	if len(routes) > 0 || (routeErr != nil && strings.TrimSpace(routeOut) == "") {
		state.box.Add(widget.NewLabel(""))
		state.box.Add(widget.NewLabelWithStyle(i18n.T(i18n.KeyFeatNetRouting), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		if len(routes) == 0 {
			if routeErr != nil {
				state.box.Add(widget.NewLabel(routeErr.Error()))
			}
		} else {
			for _, route := range routes {
				lbl := widget.NewLabel(route)
				lbl.Wrapping = fyne.TextWrapWord
				state.box.Add(lbl)
			}
		}
	}

	state.box.Refresh()
}

func parseNetIfaces(out string) []netIfaceStat {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	stats := make([]netIfaceStat, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			parts = strings.Fields(line)
		}
		if len(parts) < 3 {
			continue
		}
		rx, errRx := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		tx, errTx := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if errRx != nil || errTx != nil {
			continue
		}
		stats = append(stats, netIfaceStat{
			name: strings.TrimSpace(parts[0]),
			rx:   rx,
			tx:   tx,
		})
	}
	return stats
}

func filterNetIfaces(stats []netIfaceStat) []netIfaceStat {
	filtered := make([]netIfaceStat, 0, len(stats))
	for _, stat := range stats {
		if stat.name == "lo" {
			continue
		}
		if isVirtualNetIface(stat.name) && stat.rx+stat.tx == 0 {
			continue
		}
		filtered = append(filtered, stat)
	}
	sort.Slice(filtered, func(i, j int) bool {
		ti := filtered[i].rx + filtered[i].tx
		tj := filtered[j].rx + filtered[j].tx
		if ti != tj {
			return ti > tj
		}
		return filtered[i].name < filtered[j].name
	})
	return filtered
}

func isVirtualNetIface(name string) bool {
	return strings.HasPrefix(name, "veth") ||
		strings.HasPrefix(name, "br-") ||
		strings.HasPrefix(name, "docker") ||
		strings.HasPrefix(name, "virbr")
}

func parseNetRoutes(out string) []string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	routes := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "===") {
			continue
		}
		routes = append(routes, line)
	}
	return routes
}

func netIfaceCard(stat netIfaceStat, rxRate, txRate float64, showRate bool) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(stat.name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	total := stat.rx + stat.tx
	rxShare := 0.0
	if total > 0 {
		rxShare = float64(stat.rx) / float64(total) * 100
	}
	bar := newUsageProgressBar()
	bar.SetUsage(rxShare)

	detail := widget.NewLabel(i18n.Tf(i18n.KeyFeatNetIfaceDetail, formatBytes(stat.rx), formatBytes(stat.tx)))

	rows := []fyne.CanvasObject{title, bar, detail}
	if showRate {
		rateLbl := widget.NewLabel(i18n.Tf(i18n.KeyFeatNetRate, formatBytesPerSec(rxRate), formatBytesPerSec(txRate)))
		rateLbl.Importance = widget.MediumImportance
		rows = append(rows, rateLbl)
	}
	return withBackground(container.NewPadded(container.NewVBox(rows...)), colorPanel)
}

func formatBytesPerSec(bps float64) string {
	if bps < 0 {
		bps = 0
	}
	if bps < 1024 {
		return strconv.FormatFloat(bps, 'f', 0, 64) + " B/s"
	}
	if bps < 1024*1024 {
		return strconv.FormatFloat(bps/1024, 'f', 1, 64) + " KB/s"
	}
	if bps < 1024*1024*1024 {
		return strconv.FormatFloat(bps/(1024*1024), 'f', 1, 64) + " MB/s"
	}
	return strconv.FormatFloat(bps/(1024*1024*1024), 'f', 2, 64) + " GB/s"
}

func loadNetPorts(client *remote.Client, setText func(string)) {
	setText(i18n.T(i18n.KeyFeatLoading))
	go func() {
		out, err := client.RunCombined(netPortsCmd)
		fyne.Do(func() {
			if err != nil && strings.TrimSpace(out) == "" {
				setText(err.Error())
				return
			}
			setText(strings.TrimSpace(out))
		})
	}()
}

func netTrafficLoop(client *remote.Client, state *netTrafficState, stop <-chan struct{}) {
	loadNetTraffic(client, state)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			loadNetTraffic(client, state)
		}
	}
}
