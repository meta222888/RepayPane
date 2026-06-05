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
	box         *fyne.Container
	summary     *widget.Label
	bootTotals  *widget.Label
	prev        map[string]netIfaceStat
	prevAt      time.Time
	showRate    bool
	hasContent  bool
	cards       map[string]*netIfaceCardView
	cardOrder   []string
	routeObjs   []fyne.CanvasObject
}

type netIfaceCardView struct {
	root   fyne.CanvasObject
	title  *widget.Label
	bar    *usageProgressBar
	detail *widget.Label
	rate   *widget.Label
}

func (a *App) showNetworkInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	title := i18n.T(i18n.KeyFeatNetwork)
	trafficState := &netTrafficState{
		box:        container.NewVBox(),
		summary:    widget.NewLabel(i18n.T(i18n.KeyFeatNetRatePending)),
		bootTotals: widget.NewLabel(""),
		cards:      make(map[string]*netIfaceCardView),
	}
	trafficState.summary.Importance = widget.MediumImportance
	trafficState.bootTotals.Importance = widget.LowImportance
	trafficScroll := container.NewVScroll(trafficState.box)
	setPorts, portsScroll := scrollSelectableText()

	autoRefresh := widget.NewCheck(i18n.T(i18n.KeyFeatNetAutoRefresh), nil)
	autoRefresh.SetChecked(true)

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
			trafficState.summary.SetText(i18n.T(i18n.KeyFeatNetRatePending))
			stopCh = make(chan struct{})
			go netTrafficLoop(client, trafficState, stopCh)
		} else {
			trafficState.summary.SetText(i18n.T(i18n.KeyFeatNetRateOff))
		}
	}

	refreshTraffic := newAccentButton(i18n.T(i18n.KeyFeatRefreshTraffic), func() {
		loadNetTraffic(client, trafficState, true)
	})
	refreshPorts := newAccentButton(i18n.T(i18n.KeyFeatRefreshPorts), func() {
		loadNetPorts(client, setPorts)
	})

	trafficLeft := container.NewVBox(trafficState.summary, trafficState.bootTotals)
	trafficHeader := container.NewBorder(nil, nil, trafficLeft, autoRefresh, nil)
	trafficBox := container.NewBorder(trafficHeader, refreshTraffic, nil, nil, trafficScroll)

	portsHeader := titleLabel(i18n.T(i18n.KeyFeatNetPorts))
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

	trafficState.showRate = true
	stopCh = make(chan struct{})
	go netTrafficLoop(client, trafficState, stopCh)
	loadNetPorts(client, setPorts)
}

func loadNetTraffic(client *remote.Client, state *netTrafficState, full bool) {
	if full || !state.hasContent {
		state.box.Objects = nil
		state.box.Add(featureLoadingLabel())
		state.box.Refresh()
	}

	go func() {
		ifaceOut, ifaceErr := client.RunCombined(netIfaceCmd)
		var routeOut string
		var routeErr error
		if full || !state.hasContent {
			routeOut, routeErr = client.RunCombined(netRouteCmd)
		}
		fyne.Do(func() {
			if !full && state.hasContent {
				updateNetTraffic(state, ifaceOut, ifaceErr)
				return
			}
			renderNetTraffic(state, ifaceOut, routeOut, ifaceErr, routeErr)
			state.hasContent = true
		})
	}()
}

func updateNetTraffic(state *netTrafficState, ifaceOut string, ifaceErr error) {
	stats := filterNetIfaces(parseNetIfaces(ifaceOut))
	if len(stats) == 0 {
		if ifaceErr != nil && strings.TrimSpace(ifaceOut) == "" {
			state.summary.SetText(ifaceErr.Error())
		}
		return
	}

	showRates, rxRates, txRates := applyNetSample(state, stats)

	orderChanged := len(stats) != len(state.cardOrder)
	if !orderChanged {
		for i, stat := range stats {
			if state.cardOrder[i] != stat.name {
				orderChanged = true
				break
			}
		}
	}

	for name := range state.cards {
		found := false
		for _, stat := range stats {
			if stat.name == name {
				found = true
				break
			}
		}
		if !found {
			delete(state.cards, name)
			orderChanged = true
		}
	}

	for _, stat := range stats {
		card, ok := state.cards[stat.name]
		if !ok {
			card = newNetIfaceCardView(state.showRate)
			state.cards[stat.name] = card
			orderChanged = true
		}
		card.update(stat, rxRates[stat.name], txRates[stat.name], showRates)
	}

	state.cardOrder = statNames(stats)
	if orderChanged {
		rebuildTrafficBox(state, stats)
		return
	}
	for _, stat := range stats {
		state.cards[stat.name].refresh()
	}
}

func applyNetSample(state *netTrafficState, stats []netIfaceStat) (showRates bool, rxRates, txRates map[string]float64) {
	now := time.Now()
	elapsed := now.Sub(state.prevAt).Seconds()
	if state.prevAt.IsZero() || elapsed <= 0 {
		elapsed = 0
	}
	showRates = state.showRate && state.prev != nil && elapsed > 0
	rxRates = make(map[string]float64, len(stats))
	txRates = make(map[string]float64, len(stats))
	var totalRxRate, totalTxRate float64
	for _, stat := range stats {
		rxRate, txRate := 0.0, 0.0
		if showRates {
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
			totalRxRate += rxRate
			totalTxRate += txRate
		}
		rxRates[stat.name] = rxRate
		txRates[stat.name] = txRate
	}
	if state.showRate {
		if showRates {
			state.summary.SetText(i18n.Tf(i18n.KeyFeatNetBandwidthTotal,
				formatBytesPerSec(totalRxRate), formatBytesPerSec(totalTxRate)))
		} else {
			state.summary.SetText(i18n.T(i18n.KeyFeatNetRatePending))
		}
		next := make(map[string]netIfaceStat, len(stats))
		for _, stat := range stats {
			next[stat.name] = stat
		}
		state.prev = next
		state.prevAt = now
	}
	setNetBootTotals(state, stats)
	return showRates, rxRates, txRates
}

func setNetBootTotals(state *netTrafficState, stats []netIfaceStat) {
	if state.bootTotals == nil {
		return
	}
	var totalRx, totalTx int64
	for _, stat := range stats {
		totalRx += stat.rx
		totalTx += stat.tx
	}
	state.bootTotals.SetText(i18n.Tf(i18n.KeyFeatNetBootTotals, formatBytes(totalRx), formatBytes(totalTx)))
}

func statNames(stats []netIfaceStat) []string {
	names := make([]string, len(stats))
	for i, stat := range stats {
		names[i] = stat.name
	}
	return names
}

func rebuildTrafficBox(state *netTrafficState, stats []netIfaceStat) {
	objs := make([]fyne.CanvasObject, 0, len(stats)+len(state.routeObjs))
	for _, stat := range stats {
		objs = append(objs, state.cards[stat.name].root)
	}
	objs = append(objs, state.routeObjs...)
	state.box.Objects = objs
	state.box.Refresh()
}

func renderNetTraffic(state *netTrafficState, ifaceOut, routeOut string, ifaceErr, routeErr error) {
	state.box.Objects = nil
	state.cards = make(map[string]*netIfaceCardView)
	state.cardOrder = nil
	state.routeObjs = nil

	stats := filterNetIfaces(parseNetIfaces(ifaceOut))
	if len(stats) == 0 {
		if ifaceErr != nil && strings.TrimSpace(ifaceOut) == "" {
			state.box.Add(widget.NewLabel(ifaceErr.Error()))
		} else {
			state.box.Add(widget.NewLabel(i18n.T(i18n.KeyFeatNoData)))
		}
		state.bootTotals.SetText("")
		state.box.Refresh()
		return
	}

	showRates, rxRates, txRates := applyNetSample(state, stats)

	state.cardOrder = statNames(stats)
	for _, stat := range stats {
		card := newNetIfaceCardView(state.showRate)
		card.update(stat, rxRates[stat.name], txRates[stat.name], showRates)
		state.cards[stat.name] = card
		state.box.Add(card.root)
	}

	routes := parseNetRoutes(routeOut)
	if len(routes) > 0 || (routeErr != nil && strings.TrimSpace(routeOut) == "") {
		state.routeObjs = []fyne.CanvasObject{
			widget.NewLabel(""),
			titleLabel(i18n.T(i18n.KeyFeatNetRouting)),
		}
		if len(routes) == 0 {
			if routeErr != nil {
				state.routeObjs = append(state.routeObjs, widget.NewLabel(routeErr.Error()))
			}
		} else {
			for _, route := range routes {
				lbl := widget.NewLabel(route)
				lbl.Wrapping = fyne.TextWrapWord
				state.routeObjs = append(state.routeObjs, lbl)
			}
		}
		for _, obj := range state.routeObjs {
			state.box.Add(obj)
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

func newNetIfaceCardView(showRate bool) *netIfaceCardView {
	title := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	bar := newUsageProgressBar()
	detail := widget.NewLabel("")
	rows := []fyne.CanvasObject{wrapTitleLabel(title), bar, detail}
	var rate *widget.Label
	if showRate {
		rate = widget.NewLabel("")
		rate.Importance = widget.MediumImportance
		rows = append(rows, rate)
	}
	v := &netIfaceCardView{
		title:  title,
		bar:    bar,
		detail: detail,
		rate:   rate,
	}
	v.root = withBackground(container.NewPadded(container.NewVBox(rows...)), colorPanel)
	return v
}

func (v *netIfaceCardView) update(stat netIfaceStat, rxRate, txRate float64, showRate bool) {
	v.title.SetText(stat.name)
	total := stat.rx + stat.tx
	rxShare := 0.0
	if total > 0 {
		rxShare = float64(stat.rx) / float64(total) * 100
	}
	v.bar.SetUsage(rxShare)
	v.detail.SetText(i18n.Tf(i18n.KeyFeatNetIfaceDetail, formatBytes(stat.rx), formatBytes(stat.tx)))
	if v.rate != nil {
		if showRate {
			v.rate.SetText(i18n.Tf(i18n.KeyFeatNetRate, formatBytesPerSec(rxRate), formatBytesPerSec(txRate)))
			v.rate.Show()
		} else {
			v.rate.Hide()
		}
	}
}

func (v *netIfaceCardView) refresh() {
	v.bar.Refresh()
	v.title.Refresh()
	v.detail.Refresh()
	if v.rate != nil {
		v.rate.Refresh()
	}
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

func compactPortOutput(out string) string {
	lines := strings.Split(out, "\n")
	var buf strings.Builder
	first := true
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "===") {
			continue
		}
		if !first {
			buf.WriteByte('\n')
		}
		buf.WriteString(trimmed)
		first = false
	}
	return buf.String()
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
			setText(compactPortOutput(out))
		})
	}()
}

func netTrafficLoop(client *remote.Client, state *netTrafficState, stop <-chan struct{}) {
	loadNetTraffic(client, state, !state.hasContent)
	// Second sample soon so current bandwidth appears without waiting a full interval.
	quick := time.NewTimer(1 * time.Second)
	defer quick.Stop()
	select {
	case <-stop:
		return
	case <-quick.C:
		loadNetTraffic(client, state, false)
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			loadNetTraffic(client, state, false)
		}
	}
}
