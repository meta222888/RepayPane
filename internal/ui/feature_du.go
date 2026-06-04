package ui

import (
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type duEntry struct {
	size  string
	name  string
	path  string
	isDir bool
}

func (a *App) showDiskUsageTree() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	title := i18n.T(i18n.KeyFeatDu)
	curPath := "/"
	pathLbl := widget.NewLabel(curPath)
	pathLbl.TextStyle = fyne.TextStyle{Bold: true}

	var entries []duEntry

	list := widget.NewList(
		func() int { return len(entries) },
		func() fyne.CanvasObject {
			left := widget.NewLabel("")
			right := widget.NewLabel("")
			right.Alignment = fyne.TextAlignTrailing
			return container.NewBorder(nil, nil, left, right, nil)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			var nameLbl, sizeLbl *widget.Label
			for _, o := range c.Objects {
				if l, ok := o.(*widget.Label); ok {
					if nameLbl == nil {
						nameLbl = l
					} else {
						sizeLbl = l
					}
				}
			}
			if int(id) >= len(entries) {
				return
			}
			e := entries[id]
			icon := "📄 "
			if e.isDir {
				icon = "📁 "
			}
			nameLbl.SetText(icon + e.name)
			sizeLbl.SetText(e.size)
		},
	)

	var loadDu func(string)
	loadDu = func(dir string) {
		curPath = dir
		pathLbl.SetText(dir)
		entries = nil
		list.Refresh()
		go func() {
			cmd := `du -sh "` + shellQuote(dir) + `"/* 2>/dev/null`
			out, err := client.RunCombined(cmd)
			fyne.Do(func() {
				if err != nil && strings.TrimSpace(out) == "" {
					entries = []duEntry{{size: "—", name: err.Error(), path: dir}}
					list.Refresh()
					return
				}
				parsed := parseDuLines(out, dir)
				for i := range parsed {
					st, statErr := client.Stat(parsed[i].path)
					if statErr == nil {
						parsed[i].isDir = st.IsDir
					}
				}
				entries = parsed
				list.Refresh()
			})
		}()
	}

	list.OnSelected = func(id widget.ListItemID) {
		idx := int(id)
		if idx < 0 || idx >= len(entries) {
			return
		}
		e := entries[idx]
		if e.isDir {
			loadDu(e.path)
		}
	}

	upBtn := newAccentButton(i18n.T(i18n.KeyUp), func() {
		if curPath == "/" {
			return
		}
		loadDu(path.Dir(curPath))
	})
	refreshBtn := newAccentButton(i18n.T(i18n.KeyRefresh), func() { loadDu(curPath) })
	toolbar := container.NewHBox(upBtn, refreshBtn)
	header := container.NewBorder(nil, nil, pathLbl, toolbar, nil)
	body := container.NewBorder(header, nil, nil, nil, list)
	showThemedFeature(a, title, fyne.NewSize(640, 520), body)
	loadDu("/")
}

func parseDuLines(out, parent string) []duEntry {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	outEntries := make([]duEntry, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tab := strings.IndexByte(line, '\t')
		if tab < 0 {
			continue
		}
		size := strings.TrimSpace(line[:tab])
		p := strings.TrimSpace(line[tab+1:])
		name := path.Base(p)
		if name == "" {
			name = p
		}
		outEntries = append(outEntries, duEntry{size: size, name: name, path: p, isDir: strings.HasSuffix(p, "/")})
	}
	sort.Slice(outEntries, func(i, j int) bool {
		return duSizeRank(outEntries[i].size) > duSizeRank(outEntries[j].size)
	})
	_ = parent
	return outEntries
}

func duSizeRank(s string) float64 {
	s = strings.TrimSpace(s)
	mult := 1.0
	switch {
	case strings.HasSuffix(s, "T"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "T")
	case strings.HasSuffix(s, "G"):
		mult = 1024
		s = strings.TrimSuffix(s, "G")
	case strings.HasSuffix(s, "M"):
		mult = 1
		s = strings.TrimSuffix(s, "M")
	case strings.HasSuffix(s, "K"):
		mult = 1 / 1024
		s = strings.TrimSuffix(s, "K")
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v * mult
}

func shellQuote(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
