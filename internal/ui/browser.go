package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PaneKind int

const (
	PaneLocal PaneKind = iota
	PaneRemote
)

type FilePane struct {
	app  *App
	kind PaneKind

	path      string
	local     []localEntry
	remote    []remote.FileInfo
	connected bool

	history   []string
	histIndex int

	pathLabel  *widget.Label
	panelHdrLabel *widget.Label
	breadcrumb *fyne.Container
	table      *widget.Table
	toolbar    fyne.CanvasObject
	panelHdr   fyne.CanvasObject
	root       fyne.CanvasObject

	selectedRow int
	lastTap     time.Time
	lastTapID   int

	localNav *LocalNav
}

func NewLocalPane(app *App) *FilePane {
	p := &FilePane{app: app, kind: PaneLocal, path: defaultLocalDir(), selectedRow: -1, histIndex: -1}
	p.build()
	p.pushHistory(p.path)
	p.RefreshListing()
	return p
}

func NewRemotePane(app *App) *FilePane {
	p := &FilePane{app: app, kind: PaneRemote, path: "/", selectedRow: -1}
	p.build()
	return p
}

func (p *FilePane) colCount() int {
	if p.kind == PaneRemote {
		return 4
	}
	return 3
}

func (p *FilePane) build() {
	p.breadcrumb = container.NewHBox()

	p.table = widget.NewTable(
		func() (int, int) { return p.rowCount() + 1, p.colCount() },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		p.updateCell,
	)
	p.table.SetColumnWidth(0, 240)
	p.table.SetColumnWidth(1, 90)
	if p.kind == PaneRemote {
		p.table.SetColumnWidth(2, 120)
		p.table.SetColumnWidth(3, 140)
	} else {
		p.table.SetColumnWidth(2, 140)
	}
	p.table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		row := id.Row - 1
		p.selectedRow = row
		p.handleRowSelect(id, row)
	}

	p.toolbar = p.buildToolbar()
	p.panelHdr = p.buildPanelHeader()
	p.root = p.table
	p.ApplyLanguage()
	p.refreshPathDisplay()
}

func (p *FilePane) buildToolbar() fyne.CanvasObject {
	up := widget.NewButtonWithIcon("", theme.MoveUpIcon(), p.goUp)
	newFolder := widget.NewButtonWithIcon("", theme.FolderNewIcon(), p.newFolderSoon)
	refresh := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), p.RefreshListing)
	for _, b := range []*widget.Button{up, newFolder, refresh} {
		b.Importance = widget.LowImportance
	}
	btns := container.NewHBox(up, newFolder, refresh)

	if p.kind == PaneLocal {
		p.localNav = NewLocalNav(p)
		p.pathLabel = widget.NewLabel("")
		p.pathLabel.TextStyle = fyne.TextStyle{Monospace: true}
		left := container.NewHBox(p.localNav.Button(), btns)
		row := container.NewBorder(nil, nil, left, p.pathLabel, nil)
		return withPanelHeader(row)
	}

	serverIcon := widget.NewLabel("⬡")
	pathBox := container.NewBorder(nil, nil, nil, nil, p.breadcrumb)
	left := container.NewHBox(serverIcon, pathBox)
	row := container.NewBorder(nil, nil, left, btns, nil)
	return withPanelHeader(row)
}

func (p *FilePane) buildPanelHeader() fyne.CanvasObject {
	p.panelHdrLabel = widget.NewLabel("")
	p.panelHdrLabel.TextStyle = fyne.TextStyle{Bold: true}
	return withPanelLabel(p.panelHdrLabel)
}

func (p *FilePane) rowCount() int {
	if p.kind == PaneLocal {
		return len(p.local)
	}
	return len(p.remote)
}

func (p *FilePane) updateCell(id widget.TableCellID, obj fyne.CanvasObject) {
	label := obj.(*widget.Label)
	label.Alignment = fyne.TextAlignLeading
	label.TextStyle = fyne.TextStyle{}
	if id.Col == 1 {
		label.Alignment = fyne.TextAlignTrailing
	}
	if id.Row == 0 {
		label.TextStyle = fyne.TextStyle{Bold: true}
		label.Importance = widget.MediumImportance
		switch id.Col {
		case 0:
			label.SetText(strings.ToUpper(i18n.T(i18n.KeyColName)))
		case 1:
			label.SetText(strings.ToUpper(i18n.T(i18n.KeyColSize)))
		case 2:
			if p.kind == PaneRemote {
				label.SetText(strings.ToUpper(i18n.T(i18n.KeyColPermissions)))
			} else {
				label.SetText(strings.ToUpper(i18n.T(i18n.KeyColModified)))
			}
		case 3:
			label.SetText(strings.ToUpper(i18n.T(i18n.KeyColModified)))
		}
		return
	}
	row := id.Row - 1
	if p.kind == PaneLocal {
		if row >= len(p.local) {
			return
		}
		e := p.local[row]
		switch id.Col {
		case 0:
			label.SetText(fileIcon(e.isDir) + "  " + e.name)
		case 1:
			label.SetText(formatSize(e.size, e.isDir))
		case 2:
			label.SetText(formatTime(e.mod))
		}
		return
	}
	if row >= len(p.remote) {
		return
	}
	e := p.remote[row]
	switch id.Col {
	case 0:
		label.SetText(fileIcon(e.IsDir) + "  " + e.Name)
	case 1:
		label.SetText(formatSize(e.Size, e.IsDir))
	case 2:
		label.SetText(formatRemotePerm(e.Mode, e.IsDir))
	case 3:
		label.SetText(formatTime(e.ModTime))
	}
}

func (p *FilePane) Toolbar() fyne.CanvasObject  { return p.toolbar }
func (p *FilePane) PanelHeader() fyne.CanvasObject { return p.panelHdr }
func (p *FilePane) Container() fyne.CanvasObject { return p.root }

func (p *FilePane) CurrentPath() string { return p.path }

func (p *FilePane) SetConnected(v bool) {
	p.connected = v
	if !v {
		p.remote = nil
		p.table.Refresh()
	}
}

func (p *FilePane) ApplyLanguage() {
	if p.localNav != nil {
		p.localNav.ApplyLanguage()
	}
	p.refreshPathDisplay()
	p.table.Refresh()
}

func (p *FilePane) refreshPathDisplay() {
	if p.kind == PaneLocal {
		if p.pathLabel != nil {
			p.pathLabel.SetText(p.path)
		}
		if p.panelHdrLabel != nil {
			drive := p.path
			if len(drive) >= 2 && drive[1] == ':' {
				drive = strings.ToUpper(drive[:2]) + `\`
			}
			p.panelHdrLabel.SetText("💾  " + i18n.T(i18n.KeyPanelLocal) + " — " + drive)
		}
		return
	}
	if p.panelHdrLabel != nil {
		p.panelHdrLabel.SetText("⬡  " + i18n.T(i18n.KeyPanelRemote) + " — " + p.path)
	}
}

func (p *FilePane) Navigate(path string) {
	if p.kind == PaneLocal {
		path = filepath.Clean(path)
		st, err := os.Stat(path)
		if err != nil || !st.IsDir() {
			dialog.ShowError(fmt.Errorf(i18n.Tf(i18n.KeyInvalidLocalPath, path)), p.app.window)
			return
		}
		p.pushHistory(path)
	} else {
		if !p.connected || p.app.activeClient() == nil {
			return
		}
		path = strings.ReplaceAll(path, "\\", "/")
		if path == "" {
			path = "/"
		}
	}
	p.path = path
	if p.kind == PaneLocal && p.localNav != nil {
		p.localNav.syncFromPath(path)
	}
	if p.kind == PaneRemote {
		if sess := p.app.activeSession(); sess != nil {
			sess.remotePath = path
		}
	}
	p.refreshBreadcrumb()
	p.refreshPathDisplay()
	p.RefreshListing()
}

func (p *FilePane) pushHistory(path string) {
	if p.histIndex >= 0 && p.histIndex < len(p.history) && p.history[p.histIndex] == path {
		return
	}
	if p.histIndex < len(p.history)-1 {
		p.history = p.history[:p.histIndex+1]
	}
	p.history = append(p.history, path)
	p.histIndex = len(p.history) - 1
}

func (p *FilePane) goUp() {
	if p.kind == PaneLocal {
		parent := filepath.Dir(p.path)
		if parent == p.path {
			return
		}
		p.Navigate(parent)
		return
	}
	if p.path == "/" {
		return
	}
	parent := filepath.Dir(strings.ReplaceAll(p.path, "\\", "/"))
	p.Navigate(parent)
}

func (p *FilePane) newFolderSoon() {
	dialog.ShowInformation(i18n.T(i18n.KeyMenuComingSoon), i18n.T(i18n.KeyMenuFeaturesSoon), p.app.window)
}

func (p *FilePane) refreshBreadcrumb() {
	if p.kind != PaneRemote {
		return
	}
	p.breadcrumb.Objects = nil
	parts := strings.Split(p.path, "/")
	acc := ""
	rootBtn := widget.NewButton("/", func() { p.Navigate("/") })
	rootBtn.Importance = widget.LowImportance
	p.breadcrumb.Add(rootBtn)
	for _, part := range parts {
		if part == "" {
			continue
		}
		if acc == "" || acc == "/" {
			acc = "/" + part
		} else {
			acc = acc + "/" + part
		}
		target := acc
		sep := widget.NewLabel(" › ")
		btn := widget.NewButton(part, func() { p.Navigate(target) })
		btn.Importance = widget.LowImportance
		p.breadcrumb.Add(sep)
		p.breadcrumb.Add(btn)
	}
	p.breadcrumb.Refresh()
}

func (p *FilePane) RefreshListing() {
	if p.kind == PaneLocal {
		entries, err := listLocal(p.path)
		if err != nil {
			dialog.ShowError(err, p.app.window)
			return
		}
		p.local = entries
		p.table.Refresh()
		return
	}
	if !p.connected || p.app.activeClient() == nil {
		return
	}
	entries, err := p.app.activeClient().ListDir(p.path)
	if err != nil {
		dialog.ShowError(err, p.app.window)
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	p.remote = entries
	p.table.Refresh()
}

func (p *FilePane) handleRowSelect(id widget.TableCellID, row int) {
	now := time.Now()
	isDouble := row == p.lastTapID && now.Sub(p.lastTap) < 500*time.Millisecond
	p.lastTap = now
	p.lastTapID = row
	if isDouble {
		p.activateRow(row)
		return
	}
	cellID := id
	time.AfterFunc(100*time.Millisecond, func() {
		fyne.Do(func() { p.table.Unselect(cellID) })
	})
}

func (p *FilePane) activateRow(row int) {
	if p.kind == PaneLocal {
		if row < 0 || row >= len(p.local) {
			return
		}
		e := p.local[row]
		if e.isDir {
			p.Navigate(e.path)
		}
		return
	}
	if row < 0 || row >= len(p.remote) {
		return
	}
	e := p.remote[row]
	if e.IsDir {
		p.Navigate(e.Path)
		return
	}
	p.app.openRemoteEditor(e)
}

func fileIcon(isDir bool) string {
	if isDir {
		return "📁"
	}
	return "📄"
}

func formatSize(size int64, isDir bool) string {
	if isDir {
		return "—"
	}
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("2006-01-02 15:04")
}

func formatRemotePerm(m os.FileMode, isDir bool) string {
	if m == 0 {
		if isDir {
			return "drwxr-xr-x"
		}
		return "—"
	}
	s := m.String()
	if len(s) >= 10 {
		return s
	}
	return s
}

func (p *FilePane) SelectedPath() string {
	if p.kind != PaneLocal || p.selectedRow < 0 || p.selectedRow >= len(p.local) {
		return ""
	}
	e := p.local[p.selectedRow]
	if e.isDir {
		return ""
	}
	return e.path
}

func (p *FilePane) SelectedEntry() *remote.FileInfo {
	if p.kind != PaneRemote || p.selectedRow < 0 || p.selectedRow >= len(p.remote) {
		return nil
	}
	e := p.remote[p.selectedRow]
	return &e
}
