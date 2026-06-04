package ui

import (
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const sysInfoCmd = `echo "=== Host ==="
hostname -f 2>/dev/null || hostname
echo
echo "=== OS ==="
(uname -srmo 2>/dev/null; cat /etc/os-release 2>/dev/null | grep -E '^(PRETTY_NAME|NAME|VERSION)=' ) | head -5
echo
echo "=== Kernel ==="
uname -a
echo
echo "=== Uptime ==="
uptime
echo
echo "=== CPU ==="
(lscpu 2>/dev/null | grep -E 'Model name|Architecture|CPU\\(s\\)|Thread|Core' || grep -m1 model /proc/cpuinfo 2>/dev/null || echo "n/a")
echo
echo "=== Memory ==="
free -h 2>/dev/null || cat /proc/meminfo | head -3
echo
echo "=== Disk (summary) ==="
df -hP --output=target,size,used,avail,pcent 2>/dev/null | head -10
echo
echo "=== User ==="
whoami; id 2>/dev/null
`

func (a *App) showSystemInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	title := i18n.T(i18n.KeyFeatSysInfo)
	lbl, scroll := scrollLabel()
	refresh := newAccentButton(i18n.T(i18n.KeyRefresh), func() {
		loadSysInfo(client, lbl)
	})
	body := container.NewBorder(nil, refresh, nil, nil, scroll)
	showThemedFeature(a, title, fyne.NewSize(640, 520), body)
	loadSysInfo(client, lbl)
}

func loadSysInfo(client *remote.Client, lbl *widget.Label) {
	lbl.SetText(i18n.T(i18n.KeyFeatLoading))
	go func() {
		out, err := client.RunCombined(sysInfoCmd)
		fyne.Do(func() {
			if err != nil && strings.TrimSpace(out) == "" {
				lbl.SetText(err.Error())
				return
			}
			lbl.SetText(strings.TrimSpace(out))
		})
	}()
}
