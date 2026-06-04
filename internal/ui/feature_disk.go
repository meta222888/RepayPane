package ui

import (
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showDiskSpace() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	title := i18n.T(i18n.KeyFeatDisk)
	listBox := container.NewVBox()
	scroll := container.NewVScroll(listBox)
	refresh := newAccentButton(i18n.T(i18n.KeyRefresh), func() {
		loadDiskSpace(client, listBox)
	})
	body := container.NewBorder(nil, refresh, nil, nil, scroll)
	showThemedFeature(a, title, fyne.NewSize(640, 480), body)
	loadDiskSpace(client, listBox)
}

func loadDiskSpace(client *remote.Client, listBox *fyne.Container) {
	listBox.Objects = nil
	loading := widget.NewLabel(i18n.T(i18n.KeyFeatLoading))
	listBox.Add(loading)
	listBox.Refresh()

	go func() {
		out, err := client.RunCombined(`df -hP --output=source,size,used,avail,pcent,target 2>/dev/null | tail -n +2`)
		fyne.Do(func() {
			listBox.Objects = nil
			if err != nil && strings.TrimSpace(out) == "" {
				listBox.Add(widget.NewLabel(err.Error()))
				listBox.Refresh()
				return
			}
			lines := strings.Split(strings.TrimSpace(out), "\n")
			if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
				listBox.Add(widget.NewLabel(i18n.T(i18n.KeyFeatNoData)))
				listBox.Refresh()
				return
			}
			for _, line := range lines {
				fields := splitFields(line)
				if len(fields) < 6 {
					continue
				}
				mount := fields[5]
				total := fields[1]
				used := fields[2]
				avail := fields[3]
				pct := fields[4]
				listBox.Add(diskUsageCard(mount, total, used, avail, pct))
			}
			listBox.Refresh()
		})
	}()
}

func splitFields(s string) []string {
	return strings.Fields(s)
}
