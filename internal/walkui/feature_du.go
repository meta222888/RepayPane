package walkui

import (
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type duRow struct {
	name   string
	size   string
	sizeKB int64
	path   string
	isDir  bool
}

type duTableModel struct {
	walk.TableModelBase
	rows []duRow
}

func (m *duTableModel) RowCount() int { return len(m.rows) }
func (m *duTableModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.rows) {
		return nil
	}
	r := m.rows[row]
	switch col {
	case 0:
		if r.isDir {
			return "[DIR] " + r.name
		}
		return r.name
	case 1:
		return r.size
	default:
		return r.path
	}
}

func (a *App) showDiskUsageTree() {
	a.showDuDialog("/")
}

func (a *App) showDuDialog(dir string) {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	dir = normalizeDuPath(dir)

	var dlg *walk.Dialog
	var pathLbl *walk.Label
	var tv *walk.TableView
	model := &duTableModel{}
	curPath := dir

	load := func(d string) {
		curPath = normalizeDuPath(d)
		if pathLbl != nil {
			pathLbl.SetText(curPath)
		}
		model.rows = nil
		model.PublishRowsReset()
		go func() {
			out, err := client.RunCombined(duListCmd(curPath))
			rows := parseDuRows(out, curPath, err)
			a.syncUI(func() {
				model.rows = rows
				model.PublishRowsReset()
			})
		}()
	}

	if err := (Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyFeatDu),
		MinSize:  Size{640, 520},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: i18n.T(i18n.KeyUp), OnClicked: func() {
						if curPath == "/" {
							return
						}
						load(path.Dir(curPath))
					}},
					PushButton{Text: i18n.T(i18n.KeyRefresh), OnClicked: func() { load(curPath) }},
					Label{AssignTo: &pathLbl, Text: curPath},
				},
			},
			TableView{
				AssignTo:         &tv,
				AlternatingRowBG: true,
				Columns: []TableViewColumn{
					{Title: "Name", Width: 200},
					{Title: "Size", Width: 80},
					{Title: "Path", Width: 280},
				},
				Model: model,
				OnItemActivated: func() {
					idx := tv.CurrentIndex()
					if idx < 0 || idx >= len(model.rows) {
						return
					}
					r := model.rows[idx]
					if r.isDir {
						load(r.path)
					}
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}).Create(a.mw); err != nil {
		return
	}
	load(dir)
	dlg.Run()
}

func parseDuRows(out, parent string, err error) []duRow {
	if err != nil && strings.TrimSpace(out) == "" {
		return []duRow{{name: err.Error(), size: "—", path: parent}}
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	rows := make([]duRow, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		isDir := parts[0] == "D"
		sizeKB, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		p := strings.TrimSpace(parts[2])
		name := path.Base(p)
		rows = append(rows, duRow{
			name: name, size: formatKBHuman(sizeKB), sizeKB: sizeKB,
			path: p, isDir: isDir,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].sizeKB != rows[j].sizeKB {
			return rows[i].sizeKB > rows[j].sizeKB
		}
		return rows[i].name < rows[j].name
	})
	if len(rows) == 0 {
		return []duRow{{name: i18n.T(i18n.KeyFeatNoData), size: "—", path: parent}}
	}
	return rows
}

func formatKBHuman(kb int64) string {
	return formatSize(kb * 1024)
}

func duListCmd(dir string) string {
	quoted := `"` + strings.ReplaceAll(dir, `"`, `\"`) + `"`
	tab := "\t"
	return `du -sk ` + quoted + `/* 2>/dev/null | sort -rn | while read sz p; do
  [ -z "$p" ] && continue
  if [ -d "$p" ]; then t=D; else t=F; fi
  printf "%s` + tab + `%s` + tab + `%s\n" "$t" "$sz" "$p"
done`
}

func normalizeDuPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	cleaned := filepath.ToSlash(filepath.Clean(p))
	if cleaned == "." {
		return "/"
	}
	return cleaned
}
