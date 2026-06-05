package ui

import (
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const netTrafficCmd = `echo "=== Network Interfaces ==="
cat /proc/net/dev 2>/dev/null | awk 'NR>2 {printf "%-12s RX: %10s  TX: %10s\n", $1, $2, $10}'
echo
echo "=== Routing ==="
(ip route 2>/dev/null || route -n 2>/dev/null) | head -8
`

const netPortsCmd = `echo "=== Listening Ports ==="
(ss -tulnp 2>/dev/null || netstat -tulnp 2>/dev/null || netstat -tuln 2>/dev/null)
`

func (a *App) showNetworkInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	title := i18n.T(i18n.KeyFeatNetwork)
	trafficLbl, trafficScroll := scrollLabel()
	setPorts, portsScroll := scrollLineList()

	autoRefresh := widget.NewCheck(i18n.T(i18n.KeyFeatNetAutoRefresh), nil)
	autoRefresh.SetChecked(false)

	var stopCh chan struct{}
	autoRefresh.OnChanged = func(checked bool) {
		if stopCh != nil {
			close(stopCh)
			stopCh = nil
		}
		if checked {
			stopCh = make(chan struct{})
			go netTrafficLoop(client, trafficLbl, stopCh)
		}
	}

	refreshTraffic := newAccentButton(i18n.T(i18n.KeyFeatRefreshTraffic), func() {
		loadNetTraffic(client, trafficLbl)
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
	showThemedFeature(a, title, fyne.NewSize(720, 560), split)

	loadNetTraffic(client, trafficLbl)
	loadNetPorts(client, setPorts)
}

func loadNetTraffic(client *remote.Client, lbl *widget.Label) {
	lbl.SetText(i18n.T(i18n.KeyFeatLoading))
	go func() {
		out, err := client.RunCombined(netTrafficCmd)
		fyne.Do(func() {
			if err != nil && strings.TrimSpace(out) == "" {
				lbl.SetText(err.Error())
				return
			}
			lbl.SetText(strings.TrimSpace(out))
		})
	}()
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

func netTrafficLoop(client *remote.Client, lbl *widget.Label, stop <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			loadNetTraffic(client, lbl)
		}
	}
}
