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
	var loadGen int

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

	loadingPanel := featureLoadingPanel()
	listArea := container.NewStack(list, loadingPanel)

	var loadDu func(string)
	loadDu = func(dir string) {
		loadGen++
		gen := loadGen
		curPath = dir
		pathLbl.SetText(dir)
		entries = nil
		list.Refresh()
		loadingPanel.Show()
		list.Hide()

		go func() {
			out, err := client.RunCombined(duListCmd(dir))
			parsed := parseDuLines(out, dir)
			fyne.Do(func() {
				if gen != loadGen {
					return
				}
				loadingPanel.Hide()
				list.Show()
				if err != nil && strings.TrimSpace(out) == "" {
					entries = []duEntry{{size: "—", name: err.Error(), path: dir}}
					list.Refresh()
					return
				}
				entries = parsed
				if len(entries) == 0 {
					entries = []duEntry{{size: "—", name: i18n.T(i18n.KeyFeatNoData), path: dir}}
				}
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
	body := container.NewBorder(header, nil, nil, nil, listArea)
	showThemedFeature(a, title, fyne.NewSize(640, 520), body)
	loadDu("/")
}

func duListCmd(dir string) string {
	quoted := `"` + shellQuote(dir) + `"`
	tab := "\t"
	return `du -sh ` + quoted + `/* 2>/dev/null | while IFS= read -r line; do
  sz="${line%%` + tab + `*}"
  p="${line#*` + tab + `}"
  [ -z "$p" ] && continue
  if [ -d "$p" ]; then t=D; else t=F; fi
  printf "%s` + tab + `%s` + tab + `%s\n" "$t" "$sz" "$p"
done`
}

func parseDuLines(out, parent string) []duEntry {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	outEntries := make([]duEntry, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		var isDir bool
		var size, p string
		switch {
		case len(parts) >= 3 && (parts[0] == "D" || parts[0] == "F"):
			isDir = parts[0] == "D"
			size = strings.TrimSpace(parts[1])
			p = strings.TrimSpace(parts[2])
		default:
			tab := strings.IndexByte(line, '\t')
			if tab < 0 {
				continue
			}
			size = strings.TrimSpace(line[:tab])
			p = strings.TrimSpace(line[tab+1:])
			isDir = strings.HasSuffix(p, "/")
		}
		name := path.Base(p)
		if name == "" {
			name = p
		}
		outEntries = append(outEntries, duEntry{size: size, name: name, path: p, isDir: isDir})
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
