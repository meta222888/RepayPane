package ui

import (
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type diskRow struct {
	source string
	size   string
	used   string
	avail  string
	pcent  string
	mount  string
}

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
	listBox.Add(featureLoadingLabel())
	listBox.Refresh()

	go func() {
		out, err := client.RunCombined(`df -hP 2>/dev/null || df -h 2>/dev/null`)
		fyne.Do(func() {
			listBox.Objects = nil
			rows := parseDfOutput(out)
			if len(rows) == 0 {
				if err != nil && strings.TrimSpace(out) == "" {
					listBox.Add(widget.NewLabel(err.Error()))
				} else if trimmed := strings.TrimSpace(out); trimmed != "" {
					lbl := widget.NewLabel(trimmed)
					lbl.Wrapping = fyne.TextWrapWord
					listBox.Add(lbl)
				} else {
					listBox.Add(widget.NewLabel(i18n.T(i18n.KeyFeatNoData)))
				}
				listBox.Refresh()
				return
			}
			for _, row := range rows {
				listBox.Add(diskUsageCard(row.mount, row.size, row.used, row.avail, row.pcent))
			}
			listBox.Refresh()
		})
	}()
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

func featureLoadingLabel() fyne.CanvasObject {
	lbl := widget.NewLabel(i18n.T(i18n.KeyFeatLoading))
	return container.NewCenter(lbl)
}

func featureLoadingPanel() fyne.CanvasObject {
	bar := widget.NewProgressBarInfinite()
	bar.Start()
	lbl := widget.NewLabel(i18n.T(i18n.KeyFeatLoading))
	return container.NewCenter(container.NewVBox(
		bar,
		container.NewPadded(lbl),
	))
}
