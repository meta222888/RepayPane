package ui

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"
	"github.com/relaypane/relaypane/internal/textencoding"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PaneKind int

const (
	PaneLocal PaneKind = iota
	PaneRemote

	paneDoubleClickInterval = 450 * time.Millisecond
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

	pathEntry      *widget.Entry
	panelPrefixLbl *widget.Label
	localDriveLbl     fyne.CanvasObject
	localDriveLblText *canvas.Text
	list           *widget.List
	listArea       fyne.CanvasObject
	listHeader     fyne.CanvasObject
	root           fyne.CanvasObject

	selectedRow int
	lastTapRow  int
	lastTapTime time.Time
	renamingRow int
	listPointerDown int
	pendingListRefresh bool
	lastDragAbs fyne.Position
	dragReady   bool
	listGen     int

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
	rowFactory := func() fyne.CanvasObject {
		return newPaneFileListRow(p.kind == PaneRemote)
	}
	p.list = widget.NewList(
		func() int { return p.rowCount() },
		rowFactory,
		p.updateListRow,
	)
	p.list.OnSelected = p.handleListSelect

	if p.kind == PaneLocal {
		p.listHeader = p.buildLocalListHeader()
		p.listArea = newPaneListArea(p, p.list)
		p.root = container.NewBorder(p.buildLocalChrome(), nil, nil, nil,
			container.NewBorder(p.listHeader, nil, nil, nil, p.listArea))
	} else {
		p.listHeader = p.buildRemoteListHeader()
		p.listArea = newPaneListArea(p, p.list)
		p.root = container.NewBorder(p.buildRemoteChrome(), nil, nil, nil,
			container.NewBorder(p.listHeader, nil, nil, nil, p.listArea))
	}
	p.ApplyLanguage()
	p.refreshPathDisplay()
}

func (p *FilePane) buildLocalListHeader() fyne.CanvasObject {
	nameCol := labelCText(strings.ToUpper(i18n.T(i18n.KeyColName)), colorMuted, 11)
	rightCol := labelCText(strings.ToUpper(i18n.T(i18n.KeyColSize))+"       "+strings.ToUpper(i18n.T(i18n.KeyColModified)), colorMuted, 11)
	row := container.NewBorder(nil, nil, nil, rightCol, nameCol)
	return panelBand(row, 28)
}

func (p *FilePane) buildLocalChrome() fyne.CanvasObject {
	p.localNav = NewLocalNav(p)
	t := canvas.NewText(i18n.T(i18n.KeyPanelLocal), colorMuted)
	t.TextSize = 12
	p.localDriveLblText = t
	p.localDriveLbl = wrapCanvasText(t)

	up := widget.NewButtonWithIcon("", theme.MoveUpIcon(), p.goUp)
	newFolder := widget.NewButtonWithIcon("", theme.FolderNewIcon(), p.promptNewFolder)
	refresh := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), p.RefreshListing)
	for _, b := range []*widget.Button{up, newFolder, refresh} {
		b.Importance = widget.LowImportance
	}
	actions := container.NewHBox(up, newFolder, refresh)

	p.pathEntry = widget.NewEntry()
	p.pathEntry.OnSubmitted = func(text string) {
		text = strings.TrimSpace(text)
		if text != "" {
			p.Navigate(text)
		}
	}
	left := container.NewHBox(p.localDriveLbl, p.localNav.Button())
	pathRow := container.NewBorder(nil, nil, left, actions, p.pathEntry)
	return panelBand(pathRow, 36)
}

func (p *FilePane) buildRemoteListHeader() fyne.CanvasObject {
	nameCol := labelCText(strings.ToUpper(i18n.T(i18n.KeyColName)), colorMuted, 11)
	sizeCol := labelCText(strings.ToUpper(i18n.T(i18n.KeyColSize)), colorMuted, 11)
	metaCol := labelCText(strings.ToUpper(i18n.T(i18n.KeyColModified)), colorMuted, 11)
	right := container.NewHBox(fixedWidth(metaCol, 128), fixedWidth(sizeCol, 72))
	row := container.NewBorder(nil, nil, nil, right, nameCol)
	return panelBand(row, 28)
}

func (p *FilePane) buildRemoteChrome() fyne.CanvasObject {
	p.panelPrefixLbl = widget.NewLabel("")
	p.panelPrefixLbl.Importance = widget.MediumImportance

	up := widget.NewButtonWithIcon("", theme.MoveUpIcon(), p.goUp)
	newFolder := widget.NewButtonWithIcon("", theme.FolderNewIcon(), p.promptNewFolder)
	refresh := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), p.RefreshListing)
	for _, b := range []*widget.Button{up, newFolder, refresh} {
		b.Importance = widget.LowImportance
	}
	actions := container.NewHBox(up, newFolder, refresh)

	p.pathEntry = widget.NewEntry()
	p.pathEntry.OnSubmitted = func(text string) {
		text = strings.TrimSpace(text)
		if text != "" {
			p.Navigate(text)
		}
	}
	left := container.NewHBox(widget.NewLabel("🖥"), p.panelPrefixLbl)
	pathRow := container.NewBorder(nil, nil, left, actions, p.pathEntry)
	return panelBand(pathRow, 36)
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

func (p *FilePane) listContentHeight() float32 {
	n := p.rowCount()
	if n == 0 {
		return 0
	}
	// Match widget.List row spacing (item height + theme padding between rows).
	pad := fyne.CurrentApp().Settings().Theme().Size(theme.SizeNamePadding)
	return float32(n)*paneRowMinHeight + float32(n-1)*pad
}

func (p *FilePane) updateListRow(i widget.ListItemID, obj fyne.CanvasObject) {
	idx := int(i)
	row := obj.(*paneFileListRow)
	selected := idx == p.selectedRow

	row.onPrimary = func() {
		if p.renamingRow == idx {
			return
		}
		p.tapRow(idx)
	}
	row.onSecondary = func(ev *fyne.PointEvent) {
		if p.renamingRow >= 0 {
			p.cancelRename()
		}
		p.selectRow(idx)
		p.showContextMenu(ev.AbsolutePosition, idx)
	}
	row.onMouseDown = func() { p.noteListPointerDown() }
	row.onMouseUp = func() { p.noteListPointerUp() }
	row.onDragged = func(e *fyne.DragEvent) {
		if !p.dragReady {
			if !p.setClipboardFromRow(idx) {
				return
			}
			p.dragReady = true
		}
		p.lastDragAbs = e.AbsolutePosition
	}
	row.onDragEnd = func() {
		p.dragReady = false
		if p.app.clipboard != nil {
			p.app.completePaneDrop(p, p.lastDragAbs)
		}
	}

	if p.isParentRow(idx) {
		row.onDragged = nil
		row.onDragEnd = nil
		if p.kind == PaneLocal {
			row.updateLocal(idx, "..", "—", "—", true, true, selected)
		} else {
			row.updateRemote(idx, "..", "—", "—", true, true, selected)
		}
		row.endRename()
		return
	}

	dataIdx := p.dataRowIndex(idx)
	if p.kind == PaneLocal {
		if dataIdx < 0 || dataIdx >= len(p.local) {
			return
		}
		e := p.local[dataIdx]
		row.updateLocal(idx, e.name, formatSize(e.size, e.isDir), formatTime(e.mod), e.isDir, false, selected)
		p.applyRowRenameState(row, idx)
		return
	}
	if dataIdx < 0 || dataIdx >= len(p.remote) {
		return
	}
	e := p.remote[dataIdx]
	row.updateRemote(idx, e.Name, formatSize(e.Size, e.IsDir), formatTime(e.ModTime), e.IsDir, false, selected)
	p.applyRowRenameState(row, idx)
}

func (p *FilePane) applyRowRenameState(row *paneFileListRow, idx int) {
	if idx != p.renamingRow {
		row.endRename()
		return
	}
	if row.renaming {
		return
	}
	name := p.nameForRow(idx)
	if name == "" {
		return
	}
	row.startRename(name, func(newName string) {
		p.commitRename(idx, newName)
	}, func() {
		p.cancelRename()
	})
}
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
		if p.localDriveLblText != nil {
			p.localDriveLblText.Text = i18n.T(i18n.KeyPanelLocal)
			canvas.Refresh(p.localDriveLblText)
		}
		if p.pathEntry != nil {
			p.pathEntry.SetText(p.path)
		}
		if p.localNav != nil {
			p.localNav.syncFromPath(p.path)
		}
		return
	}
	if p.panelPrefixLbl != nil {
		p.panelPrefixLbl.SetText(i18n.T(i18n.KeyPanelRemote))
	}
	if p.pathEntry != nil {
		p.pathEntry.SetText(p.path)
	}
}

func (p *FilePane) Navigate(path string) {
	p.cancelRename()
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
		path = cleanRemotePath(path)
	}
	p.path = path
	p.selectedRow = -1
	p.lastTapRow = -1
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
	parent := remoteParentPath(p.path)
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
		p.refreshListIfAllowed()
		return
	}
	if !p.connected || p.app.activeClient() == nil {
		return
	}
	client := p.app.activeClient()
	dir := p.path
	p.listGen++
	gen := p.listGen
	go func() {
		entries, err := client.ListDir(dir)
		if err == nil {
			sort.Slice(entries, func(i, j int) bool {
				if entries[i].IsDir != entries[j].IsDir {
					return entries[i].IsDir
				}
				return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
			})
		}
		fyne.Do(func() {
			if gen != p.listGen || p.path != dir {
				return
			}
			if err != nil {
				dialog.ShowError(err, p.app.window)
				return
			}
			p.remote = entries
			p.refreshListIfAllowed()
		})
	}()
}

func (p *FilePane) noteListPointerDown() {
	p.listPointerDown++
}

func (p *FilePane) noteListPointerUp() {
	if p.listPointerDown > 0 {
		p.listPointerDown--
	}
	if p.listPointerDown == 0 && p.pendingListRefresh {
		p.refreshListIfAllowed()
	}
}

func (p *FilePane) refreshListIfAllowed() {
	if p.list == nil {
		return
	}
	if p.listPointerDown > 0 {
		p.pendingListRefresh = true
		return
	}
	p.pendingListRefresh = false
	p.list.Refresh()
}

func (p *FilePane) tapRow(row int) {
	if row < 0 || row >= p.rowCount() {
		return
	}
	if p.renamingRow >= 0 && row != p.renamingRow {
		p.cancelRename()
	}
	now := time.Now()
	elapsed := now.Sub(p.lastTapTime)
	if row == p.lastTapRow && elapsed <= paneDoubleClickInterval {
		p.lastTapRow = -1
		p.cancelRename()
		p.activateRow(row)
		return
	}
	if row == p.selectedRow && row == p.lastTapRow && elapsed > paneDoubleClickInterval && !p.isParentRow(row) {
		p.lastTapRow = -1
		p.startRename(row)
		return
	}
	p.lastTapRow = row
	p.lastTapTime = now
	p.selectRow(row)
}

func validRenameName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return false
	}
	return !strings.ContainsAny(name, `/\:*?"<>|`)
}

func (p *FilePane) startRename(row int) {
	if p.isParentRow(row) {
		return
	}
	if p.kind == PaneRemote && (!p.connected || p.app.activeClient() == nil) {
		return
	}
	if p.renamingRow >= 0 {
		p.cancelRename()
	}
	p.renamingRow = row
	p.selectRow(row)
	p.list.RefreshItem(widget.ListItemID(row))
}

func (p *FilePane) cancelRename() {
	if p.renamingRow < 0 {
		return
	}
	row := p.renamingRow
	p.renamingRow = -1
	p.list.RefreshItem(widget.ListItemID(row))
}

func (p *FilePane) commitRename(row int, newName string) {
	newName = strings.TrimSpace(newName)
	oldName := p.nameForRow(row)
	if !validRenameName(newName) {
		dialog.ShowError(fmt.Errorf(i18n.T(i18n.KeyRenameInvalidName)), p.app.window)
		return
	}
	if newName == oldName {
		p.cancelRename()
		return
	}
	newPath := p.joinPath(newName)
	if p.pathExistsAt(newPath) {
		dialog.ShowError(fmt.Errorf(i18n.Tf(i18n.KeyFileExists, newName)), p.app.window)
		return
	}
	oldPath := p.fullPathForRow(row)
	if oldPath == "" {
		p.cancelRename()
		return
	}

	if p.kind == PaneLocal {
		if err := os.Rename(oldPath, newPath); err != nil {
			dialog.ShowError(err, p.app.window)
			return
		}
		p.cancelRename()
		p.RefreshListing()
		return
	}
	client := p.app.activeClient()
	if client == nil {
		p.cancelRename()
		return
	}
	if err := client.Rename(oldPath, newPath); err != nil {
		dialog.ShowError(err, p.app.window)
		return
	}
	p.cancelRename()
	p.RefreshListing()
}

func (p *FilePane) clearSelection() {
	p.clearSelectionQuiet()
}

func (p *FilePane) clearSelectionQuiet() {
	p.cancelRename()
	prev := p.selectedRow
	if prev < 0 {
		return
	}
	p.selectedRow = -1
	p.list.RefreshItem(widget.ListItemID(prev))
}

func (p *FilePane) selectRow(row int) {
	if row < 0 || row >= p.rowCount() {
		return
	}
	prev := p.selectedRow
	if prev == row {
		return
	}
	p.selectedRow = row
	if prev >= 0 {
		p.list.RefreshItem(widget.ListItemID(prev))
	}
	p.list.RefreshItem(widget.ListItemID(row))
	p.list.Select(widget.ListItemID(row))
}

func (p *FilePane) handleListSelect(id widget.ListItemID) {
	row := int(id)
	prev := p.selectedRow
	if prev == row {
		return
	}
	p.selectedRow = row
	if prev >= 0 {
		p.list.RefreshItem(widget.ListItemID(prev))
	}
	p.list.RefreshItem(widget.ListItemID(row))
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
			return
		}
		p.app.openLocalEditor(e.path, e.name, e.size)
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

// beginContextMenuSelection updates selection for a context menu without RefreshItem
// (avoids EnsureMinSize fighting the popup overlay). Returns row indices to refresh after dismiss.
func (p *FilePane) beginContextMenuSelection(row int) []int {
	if row < 0 {
		prev := p.selectedRow
		if prev < 0 {
			return nil
		}
		p.selectedRow = -1
		return []int{prev}
	}
	prev := p.selectedRow
	if prev == row {
		return nil
	}
	p.selectedRow = row
	if prev >= 0 {
		return []int{prev, row}
	}
	return []int{row}
}

func (p *FilePane) dismissContextMenu() {
	dismissPopUpMenus(p.app.window.Canvas())
}

func (p *FilePane) showContextMenu(at fyne.Position, row int) {
	deferRefresh := p.beginContextMenuSelection(row)

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
	if clip == nil {
		pasteItem.Disabled = true
	} else if clip.Kind == PaneLocal && p.kind == PaneRemote {
		if !p.connected || p.app.activeClient() == nil {
			pasteItem.Disabled = true
		}
	} else if clip.Kind == PaneRemote && p.kind == PaneLocal {
		if !p.connected || p.app.activeClient() == nil {
			pasteItem.Disabled = true
		}
	}
	if p.kind == PaneRemote && (!p.connected || p.app.activeClient() == nil) {
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
	showPopUpContextMenu(p.app.window, at, menu, func() {
		for _, idx := range deferRefresh {
			if idx >= 0 && idx < p.rowCount() {
				p.list.RefreshItem(widget.ListItemID(idx))
			}
		}
	})
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
	_ = p.setClipboardFromRow(p.selectedRow)
}

func (p *FilePane) setClipboardFromRow(row int) bool {
	if row < 0 || p.isParentRow(row) {
		return false
	}
	path := p.fullPathForRow(row)
	name := p.nameForRow(row)
	if path == "" || name == "" {
		return false
	}
	p.app.clipboard = &PaneClipboard{
		Kind:  p.kind,
		Path:  path,
		Name:  name,
		IsDir: p.isDirForRow(row),
	}
	return true
}

func (p *FilePane) ctxPaste() {
	clip := p.app.clipboard
	if clip == nil {
		return
	}
	if clip.Kind != p.kind {
		p.app.transferClipboardToPane(clip, p)
		return
	}
	p.pasteClipboard(clip)
}

func (p *FilePane) pasteClipboard(clip *PaneClipboard) {
	if clip == nil {
		return
	}
	exists := func(name string) bool { return p.pathExistsAt(p.joinPath(name)) }
	if exists(clip.Name) {
		p.app.resolveFileConflict(clip.Name, exists, func(name string) {
			if name == "" {
				return
			}
			p.pasteClipboardAs(clip, name)
		})
		return
	}
	p.pasteClipboardAs(clip, clip.Name)
}

func (p *FilePane) pasteClipboardAs(clip *PaneClipboard, destName string) {
	dst := p.joinPath(destName)
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

func (p *FilePane) fullPathForRow(row int) string {
	if p.isParentRow(row) {
		return ""
	}
	dataIdx := p.dataRowIndex(row)
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

func (p *FilePane) nameForRow(row int) string {
	if p.isParentRow(row) {
		return ""
	}
	dataIdx := p.dataRowIndex(row)
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

func (p *FilePane) isDirForRow(row int) bool {
	if p.isParentRow(row) {
		return false
	}
	dataIdx := p.dataRowIndex(row)
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
		ShowLocalEditor(p.app, dst, name, "", textencoding.Info{Encoding: textencoding.UTF8})
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
	ShowEditor(p.app, entry, "", textencoding.Info{Encoding: textencoding.UTF8})
}

func (p *FilePane) joinPath(name string) string {
	if p.kind == PaneLocal {
		return filepath.Join(p.path, name)
	}
	return path.Join(cleanRemotePath(p.path), name)
}

func cleanRemotePath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if p == "" {
		return "/"
	}
	return path.Clean(p)
}

func remoteParentPath(p string) string {
	p = cleanRemotePath(p)
	if p == "/" {
		return "/"
	}
	parent := path.Dir(p)
	if parent == "." || parent == "" {
		return "/"
	}
	return parent
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
