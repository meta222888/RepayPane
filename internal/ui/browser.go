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

type PaneClipboard struct {
	Kind  PaneKind
	Path  string
	Name  string
	IsDir bool
}

type FilePane struct {
	app  *App
	kind PaneKind

	path      string
	local     []localEntry
	remote    []remote.FileInfo
	connected bool

	history   []string
	histIndex int

	pathEntry     *widget.Entry
	panelPrefixLbl *widget.Label
	list          *widget.List
	listHeader    fyne.CanvasObject
	toolbar       fyne.CanvasObject
	panelHdr      fyne.CanvasObject
	root          fyne.CanvasObject

	selectedRow int

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

func (p *FilePane) build() {
	p.list = widget.NewList(
		func() int { return p.rowCount() },
		func() fyne.CanvasObject { return newFileListRow() },
		p.updateListRow,
	)
	p.list.OnSelected = p.handleListSelect

	p.listHeader = p.buildListHeader()
	p.root = container.NewBorder(p.listHeader, nil, nil, nil, p.list)

	p.toolbar = p.buildToolbar()
	p.panelHdr = p.buildPanelHeader()
	p.ApplyLanguage()
	p.refreshPathDisplay()
}

func (p *FilePane) buildListHeader() fyne.CanvasObject {
	name := widget.NewLabel(strings.ToUpper(i18n.T(i18n.KeyColName)))
	size := widget.NewLabel(strings.ToUpper(i18n.T(i18n.KeyColSize)))
	meta := widget.NewLabel(strings.ToUpper(i18n.T(i18n.KeyColModified)))
	name.Importance = widget.MediumImportance
	size.Importance = widget.MediumImportance
	meta.Importance = widget.MediumImportance
	size.Alignment = fyne.TextAlignTrailing
	meta.Alignment = fyne.TextAlignTrailing
	row := container.NewBorder(nil, nil, name, container.NewHBox(fixedWidth(meta, 128), fixedWidth(size, 72)), nil)
	return withBackground(container.NewPadded(row), colorPanelHeader)
}

func (p *FilePane) buildToolbar() fyne.CanvasObject {
	up := widget.NewButtonWithIcon("", theme.MoveUpIcon(), p.goUp)
	newFolder := widget.NewButtonWithIcon("", theme.FolderNewIcon(), p.promptNewFolder)
	refresh := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), p.RefreshListing)
	for _, b := range []*widget.Button{up, newFolder, refresh} {
		b.Importance = widget.LowImportance
	}
	btns := container.NewHBox(up, newFolder, refresh)

	if p.kind == PaneLocal {
		p.localNav = NewLocalNav(p)
		left := container.NewHBox(p.localNav.Button(), btns)
		return withPanelHeader(left)
	}

	left := container.NewHBox(widget.NewLabel("🖥"), btns)
	return withPanelHeader(left)
}

func (p *FilePane) buildPanelHeader() fyne.CanvasObject {
	icon := "💻"
	if p.kind == PaneRemote {
		icon = "🖥"
	}
	iconLbl := widget.NewLabel(icon)
	p.panelPrefixLbl = widget.NewLabel("")
	p.panelPrefixLbl.Importance = widget.MediumImportance
	p.pathEntry = widget.NewEntry()
	p.pathEntry.OnSubmitted = func(text string) {
		text = strings.TrimSpace(text)
		if text != "" {
			p.Navigate(text)
		}
	}
	titleRow := container.NewHBox(iconLbl, p.panelPrefixLbl)
	row := container.NewBorder(nil, nil, titleRow, nil, p.pathEntry)
	return withPanelLabel(row)
}

func (p *FilePane) hasParentRow() bool {
	if p.kind == PaneLocal {
		return filepath.Dir(p.path) != p.path
	}
	return p.path != "/"
}

func (p *FilePane) isParentRow(row int) bool {
	return p.hasParentRow() && row == 0
}

func (p *FilePane) dataRowIndex(row int) int {
	if p.isParentRow(row) {
		return -1
	}
	if p.hasParentRow() {
		return row - 1
	}
	return row
}

func (p *FilePane) rowCount() int {
	n := len(p.local)
	if p.kind == PaneRemote {
		n = len(p.remote)
	}
	if p.hasParentRow() {
		n++
	}
	return n
}

func (p *FilePane) updateListRow(i widget.ListItemID, obj fyne.CanvasObject) {
	row := obj.(*fileListRow)
	idx := int(i)
	row.onPrimary = func() { p.selectRow(idx) }
	row.onDouble = func() {
		p.selectRow(idx)
		p.activateRow(idx)
	}
	row.onSecondary = func(ev *fyne.PointEvent) {
		p.selectRow(idx)
		p.showContextMenu(ev.AbsolutePosition)
	}
	row.setRowStyle(idx, idx == p.selectedRow)

	if p.isParentRow(idx) {
		row.nameLbl.SetText("↩  ..")
		row.sizeLbl.SetText("—")
		row.metaLbl.SetText("—")
		return
	}

	dataIdx := p.dataRowIndex(idx)
	if p.kind == PaneLocal {
		if dataIdx < 0 || dataIdx >= len(p.local) {
			return
		}
		e := p.local[dataIdx]
		row.nameLbl.SetText(fileIcon(e.isDir) + "  " + e.name)
		row.sizeLbl.SetText(formatSize(e.size, e.isDir))
		row.metaLbl.SetText(formatTime(e.mod))
		return
	}
	if dataIdx < 0 || dataIdx >= len(p.remote) {
		return
	}
	e := p.remote[dataIdx]
	row.nameLbl.SetText(fileIcon(e.IsDir) + "  " + e.Name)
	row.sizeLbl.SetText(formatSize(e.Size, e.IsDir))
	row.metaLbl.SetText(formatTime(e.ModTime))
}

func (p *FilePane) Toolbar() fyne.CanvasObject     { return p.toolbar }
func (p *FilePane) PanelHeader() fyne.CanvasObject { return p.panelHdr }
func (p *FilePane) Container() fyne.CanvasObject   { return p.root }

func (p *FilePane) CurrentPath() string { return p.path }

func (p *FilePane) SetConnected(v bool) {
	p.connected = v
	if !v {
		p.remote = nil
		p.list.Refresh()
	}
}

func (p *FilePane) ApplyLanguage() {
	if p.localNav != nil {
		p.localNav.ApplyLanguage()
	}
	p.refreshPathDisplay()
	p.list.Refresh()
}

func (p *FilePane) refreshPathDisplay() {
	if p.panelPrefixLbl != nil {
		if p.kind == PaneLocal {
			p.panelPrefixLbl.SetText(i18n.T(i18n.KeyPanelLocal))
		} else {
			p.panelPrefixLbl.SetText(i18n.T(i18n.KeyPanelRemote))
		}
	}
	if p.pathEntry != nil {
		p.pathEntry.SetText(p.path)
	}
}

func (p *FilePane) Navigate(path string) {
	if p.kind == PaneLocal {
		path = filepath.Clean(path)
		st, err := os.Stat(path)
		if err != nil || !st.IsDir() {
			dialog.ShowError(fmt.Errorf(i18n.Tf(i18n.KeyInvalidLocalPath, path)), p.app.window)
			p.refreshPathDisplay()
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
	p.selectedRow = -1
	if p.kind == PaneLocal && p.localNav != nil {
		p.localNav.syncFromPath(path)
	}
	if p.kind == PaneRemote {
		if sess := p.app.activeSession(); sess != nil {
			sess.remotePath = path
		}
	}
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

func (p *FilePane) RefreshListing() {
	if p.kind == PaneLocal {
		entries, err := listLocal(p.path)
		if err != nil {
			dialog.ShowError(err, p.app.window)
			return
		}
		p.local = entries
		p.list.Refresh()
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
	p.list.Refresh()
}

func (p *FilePane) selectRow(row int) {
	if row < 0 || row >= p.rowCount() {
		return
	}
	prev := p.selectedRow
	p.selectedRow = row
	p.list.Select(widget.ListItemID(row))
	if prev >= 0 && prev != row {
		p.list.RefreshItem(widget.ListItemID(prev))
	}
	p.list.RefreshItem(widget.ListItemID(row))
}

func (p *FilePane) handleListSelect(id widget.ListItemID) {
	p.selectRow(int(id))
}

func (p *FilePane) activateRow(row int) {
	if p.isParentRow(row) {
		p.goUp()
		return
	}
	dataIdx := p.dataRowIndex(row)
	if p.kind == PaneLocal {
		if dataIdx < 0 || dataIdx >= len(p.local) {
			return
		}
		e := p.local[dataIdx]
		if e.isDir {
			p.Navigate(e.path)
		}
		return
	}
	if dataIdx < 0 || dataIdx >= len(p.remote) {
		return
	}
	e := p.remote[dataIdx]
	if e.IsDir {
		p.Navigate(e.Path)
		return
	}
	p.app.openRemoteEditor(e)
}

func (p *FilePane) showContextMenu(at fyne.Position) {
	copyItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxCopy), p.ctxCopy)
	pasteItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxPaste), p.ctxPaste)
	newFolderItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxNewFolder), p.promptNewFolder)
	newFileItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxNewFile), p.promptNewFile)
	deleteItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxDelete), p.ctxDelete)

	if p.selectedRow < 0 || p.selectedRow >= p.rowCount() || p.isParentRow(p.selectedRow) {
		copyItem.Disabled = true
		deleteItem.Disabled = true
	}
	clip := p.app.clipboard
	if clip == nil || clip.Kind != p.kind {
		pasteItem.Disabled = true
	}
	if p.kind == PaneRemote && (!p.connected || p.app.activeClient() == nil) {
		pasteItem.Disabled = true
		newFolderItem.Disabled = true
		newFileItem.Disabled = true
		deleteItem.Disabled = true
	}

	menu := fyne.NewMenu("",
		copyItem,
		pasteItem,
		fyne.NewMenuItemSeparator(),
		newFolderItem,
		newFileItem,
		fyne.NewMenuItemSeparator(),
		deleteItem,
	)
	popup := widget.NewPopUpMenu(menu, p.app.window.Canvas())
	popup.ShowAtPosition(at)
}

func (p *FilePane) selectedName() string {
	if p.selectedRow < 0 || p.isParentRow(p.selectedRow) {
		return ""
	}
	dataIdx := p.dataRowIndex(p.selectedRow)
	if p.kind == PaneLocal {
		if dataIdx < 0 || dataIdx >= len(p.local) {
			return ""
		}
		return p.local[dataIdx].name
	}
	if dataIdx < 0 || dataIdx >= len(p.remote) {
		return ""
	}
	return p.remote[dataIdx].Name
}

func (p *FilePane) selectedFullPath() string {
	if p.selectedRow < 0 || p.isParentRow(p.selectedRow) {
		return ""
	}
	dataIdx := p.dataRowIndex(p.selectedRow)
	if p.kind == PaneLocal {
		if dataIdx < 0 || dataIdx >= len(p.local) {
			return ""
		}
		return p.local[dataIdx].path
	}
	if dataIdx < 0 || dataIdx >= len(p.remote) {
		return ""
	}
	return p.remote[dataIdx].Path
}

func (p *FilePane) selectedIsDir() bool {
	if p.selectedRow < 0 || p.isParentRow(p.selectedRow) {
		return false
	}
	dataIdx := p.dataRowIndex(p.selectedRow)
	if p.kind == PaneLocal {
		if dataIdx < 0 || dataIdx >= len(p.local) {
			return false
		}
		return p.local[dataIdx].isDir
	}
	if dataIdx < 0 || dataIdx >= len(p.remote) {
		return false
	}
	return p.remote[dataIdx].IsDir
}

func (p *FilePane) ctxCopy() {
	path := p.selectedFullPath()
	if path == "" {
		return
	}
	p.app.clipboard = &PaneClipboard{
		Kind:  p.kind,
		Path:  path,
		Name:  p.selectedName(),
		IsDir: p.selectedIsDir(),
	}
}

func (p *FilePane) ctxPaste() {
	clip := p.app.clipboard
	if clip == nil || clip.Kind != p.kind {
		return
	}
	dst := p.joinPath(clip.Name)
	if p.pathExistsAt(dst) {
		dialog.ShowError(fmt.Errorf(i18n.Tf(i18n.KeyFileExists, clip.Name)), p.app.window)
		return
	}
	if p.kind == PaneLocal {
		if err := copyPathLocal(clip.Path, dst); err != nil {
			dialog.ShowError(err, p.app.window)
			return
		}
		p.RefreshListing()
		return
	}
	client := p.app.activeClient()
	if client == nil {
		return
	}
	if err := client.CopyPath(clip.Path, dst); err != nil {
		dialog.ShowError(err, p.app.window)
		return
	}
	p.RefreshListing()
}

func (p *FilePane) ctxDelete() {
	name := p.selectedName()
	path := p.selectedFullPath()
	if path == "" {
		return
	}
	dialog.ShowConfirm(i18n.T(i18n.KeyDelete), i18n.Tf(i18n.KeyDeleteFileConfirm, name), func(ok bool) {
		if !ok {
			return
		}
		if p.kind == PaneLocal {
			if err := removePathLocal(path); err != nil {
				dialog.ShowError(err, p.app.window)
				return
			}
			p.selectedRow = -1
			p.RefreshListing()
			return
		}
		client := p.app.activeClient()
		if client == nil {
			return
		}
		var err error
		if p.selectedIsDir() {
			err = client.RemoveAll(path)
		} else {
			err = client.Remove(path)
		}
		if err != nil {
			dialog.ShowError(err, p.app.window)
			return
		}
		p.selectedRow = -1
		p.RefreshListing()
	}, p.app.window)
}

func (p *FilePane) promptNewFolder() {
	entry := widget.NewEntry()
	dialog.ShowForm(i18n.T(i18n.KeyCtxNewFolder), i18n.T(i18n.KeyOK), i18n.T(i18n.KeyCancel),
		[]*widget.FormItem{widget.NewFormItem(i18n.T(i18n.KeyFormName), entry)},
		func(ok bool) {
			if !ok {
				return
			}
			name := strings.TrimSpace(entry.Text)
			if name == "" {
				return
			}
			p.createFolder(name)
		}, p.app.window)
}

func (p *FilePane) promptNewFile() {
	entry := widget.NewEntry()
	dialog.ShowForm(i18n.T(i18n.KeyCtxNewFile), i18n.T(i18n.KeyOK), i18n.T(i18n.KeyCancel),
		[]*widget.FormItem{widget.NewFormItem(i18n.T(i18n.KeyFormName), entry)},
		func(ok bool) {
			if !ok {
				return
			}
			name := strings.TrimSpace(entry.Text)
			if name == "" {
				return
			}
			p.createFile(name)
		}, p.app.window)
}

func (p *FilePane) createFolder(name string) {
	dst := p.joinPath(name)
	if p.pathExistsAt(dst) {
		dialog.ShowError(fmt.Errorf(i18n.Tf(i18n.KeyFileExists, name)), p.app.window)
		return
	}
	if p.kind == PaneLocal {
		if err := os.Mkdir(dst, 0o755); err != nil {
			dialog.ShowError(err, p.app.window)
			return
		}
		p.RefreshListing()
		return
	}
	client := p.app.activeClient()
	if client == nil {
		return
	}
	if err := client.Mkdir(dst); err != nil {
		dialog.ShowError(err, p.app.window)
		return
	}
	p.RefreshListing()
}

func (p *FilePane) createFile(name string) {
	dst := p.joinPath(name)
	if p.pathExistsAt(dst) {
		dialog.ShowError(fmt.Errorf(i18n.Tf(i18n.KeyFileExists, name)), p.app.window)
		return
	}
	if p.kind == PaneLocal {
		if err := os.WriteFile(dst, nil, 0o644); err != nil {
			dialog.ShowError(err, p.app.window)
			return
		}
		p.RefreshListing()
		ShowLocalEditor(p.app, dst, name)
		return
	}
	client := p.app.activeClient()
	if client == nil {
		return
	}
	if err := client.WriteFile(dst, nil); err != nil {
		dialog.ShowError(err, p.app.window)
		return
	}
	p.RefreshListing()
	entry := remote.FileInfo{Name: name, Path: dst}
	ShowEditor(p.app, entry, "")
}

func (p *FilePane) joinPath(name string) string {
	if p.kind == PaneLocal {
		return filepath.Join(p.path, name)
	}
	return filepath.ToSlash(filepath.Join(p.path, name))
}

func (p *FilePane) pathExistsAt(fullPath string) bool {
	if p.kind == PaneLocal {
		_, err := os.Stat(fullPath)
		return err == nil
	}
	client := p.app.activeClient()
	if client == nil {
		return false
	}
	_, err := client.Stat(fullPath)
	return err == nil
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

func (p *FilePane) SelectedPath() string {
	if p.kind != PaneLocal {
		return ""
	}
	dataIdx := p.dataRowIndex(p.selectedRow)
	if dataIdx < 0 || dataIdx >= len(p.local) {
		return ""
	}
	e := p.local[dataIdx]
	if e.isDir {
		return ""
	}
	return e.path
}

func (p *FilePane) SelectedEntry() *remote.FileInfo {
	if p.kind != PaneRemote {
		return nil
	}
	dataIdx := p.dataRowIndex(p.selectedRow)
	if dataIdx < 0 || dataIdx >= len(p.remote) {
		return nil
	}
	e := p.remote[dataIdx]
	return &e
}
