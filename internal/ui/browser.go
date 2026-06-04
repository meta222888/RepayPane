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

	pathLabel     *widget.Label
	panelHdrLabel *widget.Label
	breadcrumb    *fyne.Container
	list          *widget.List
	listHeader    fyne.CanvasObject
	toolbar       fyne.CanvasObject
	panelHdr      fyne.CanvasObject
	root          fyne.CanvasObject

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

func (p *FilePane) build() {
	p.breadcrumb = container.NewHBox()

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
	meta := widget.NewLabel("")
	if p.kind == PaneRemote {
		meta.SetText(strings.ToUpper(i18n.T(i18n.KeyColModified)))
	} else {
		meta.SetText(strings.ToUpper(i18n.T(i18n.KeyColModified)))
	}
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
		p.pathLabel = widget.NewLabel("")
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

func (p *FilePane) updateListRow(i widget.ListItemID, obj fyne.CanvasObject) {
	row := obj.(*fileListRow)
	row.onSecondary = func(ev *fyne.PointEvent) {
		p.selectedRow = int(i)
		p.list.Select(i)
		p.showContextMenu(ev.AbsolutePosition)
	}

	if p.kind == PaneLocal {
		if int(i) >= len(p.local) {
			return
		}
		e := p.local[i]
		row.nameLbl.SetText(fileIcon(e.isDir) + "  " + e.name)
		row.sizeLbl.SetText(formatSize(e.size, e.isDir))
		row.metaLbl.SetText(formatTime(e.mod))
		return
	}
	if int(i) >= len(p.remote) {
		return
	}
	e := p.remote[i]
	row.nameLbl.SetText(fileIcon(e.IsDir) + "  " + e.Name)
	row.sizeLbl.SetText(formatSize(e.Size, e.IsDir))
	if p.kind == PaneRemote {
		row.metaLbl.SetText(formatTime(e.ModTime))
	} else {
		row.metaLbl.SetText(formatTime(e.ModTime))
	}
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
	p.selectedRow = -1
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

func (p *FilePane) handleListSelect(id widget.ListItemID) {
	row := int(id)
	now := time.Now()
	isDouble := row == p.lastTapID && now.Sub(p.lastTap) < 500*time.Millisecond
	p.lastTap = now
	p.lastTapID = row
	p.selectedRow = row
	if isDouble {
		p.activateRow(row)
		return
	}
	listID := id
	time.AfterFunc(100*time.Millisecond, func() {
		fyne.Do(func() { p.list.Unselect(listID) })
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

func (p *FilePane) showContextMenu(at fyne.Position) {
	copyItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxCopy), p.ctxCopy)
	pasteItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxPaste), p.ctxPaste)
	newFolderItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxNewFolder), p.promptNewFolder)
	newFileItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxNewFile), p.promptNewFile)
	deleteItem := fyne.NewMenuItem(i18n.T(i18n.KeyCtxDelete), p.ctxDelete)

	if p.selectedRow < 0 || p.selectedRow >= p.rowCount() {
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
	if p.selectedRow < 0 {
		return ""
	}
	if p.kind == PaneLocal {
		if p.selectedRow >= len(p.local) {
			return ""
		}
		return p.local[p.selectedRow].name
	}
	if p.selectedRow >= len(p.remote) {
		return ""
	}
	return p.remote[p.selectedRow].Name
}

func (p *FilePane) selectedFullPath() string {
	if p.selectedRow < 0 {
		return ""
	}
	if p.kind == PaneLocal {
		if p.selectedRow >= len(p.local) {
			return ""
		}
		return p.local[p.selectedRow].path
	}
	if p.selectedRow >= len(p.remote) {
		return ""
	}
	return p.remote[p.selectedRow].Path
}

func (p *FilePane) selectedIsDir() bool {
	if p.selectedRow < 0 {
		return false
	}
	if p.kind == PaneLocal {
		if p.selectedRow >= len(p.local) {
			return false
		}
		return p.local[p.selectedRow].isDir
	}
	if p.selectedRow >= len(p.remote) {
		return false
	}
	return p.remote[p.selectedRow].IsDir
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
