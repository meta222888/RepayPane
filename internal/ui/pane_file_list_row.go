package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	paneRowMinHeight = PaneRowHeight
)

type paneRenameEntry struct {
	widget.Entry
	onBlur func()
}

func newPaneRenameEntry() *paneRenameEntry {
	e := &paneRenameEntry{}
	e.Entry.ExtendBaseWidget(&e.Entry)
	return e
}

func (e *paneRenameEntry) FocusLost() {
	e.Entry.FocusLost()
	if e.onBlur != nil {
		e.onBlur()
	}
}

type paneFileListRow struct {
	widget.BaseWidget

	remote   bool
	rowIndex int
	selected bool
	hovered  bool
	renaming bool

	bg        *canvas.Rectangle
	nameCell  *paneFileNameCell
	nameEntry *paneRenameEntry
	sizeT     *paneRightText
	metaT     *paneRightText

	onSecondary    func(*fyne.PointEvent)
	onPrimary      func(ctrl bool)
	onDragged      func(*fyne.DragEvent)
	onDragEnd      func()
	onMouseDown    func()
	onMouseUp      func()
	onRenameCommit func(string)
	onRenameCancel func()

	dragActive bool
	ctrlDown   bool
}

func newPaneFileListRow(remote bool) *paneFileListRow {
	r := &paneFileListRow{remote: remote}
	r.ExtendBaseWidget(r)
	return r
}

func (r *paneFileListRow) renameText() string {
	if r.nameEntry == nil {
		return ""
	}
	return r.nameEntry.Text
}

func (r *paneFileListRow) startRename(name string, onCommit func(string), onCancel func(), onBlur func()) {
	r.onRenameCommit = onCommit
	r.onRenameCancel = onCancel
	r.renaming = true
	if r.nameEntry == nil {
		return
	}
	r.nameEntry.SetText(name)
	r.nameEntry.onBlur = func() {
		if !r.renaming {
			return
		}
		if onBlur != nil {
			onBlur()
		}
	}
	r.nameCell.Hide()
	r.nameEntry.Show()
	r.nameCell.Refresh()
	canvas.Refresh(r.nameEntry)
	r.Refresh()
	fyne.Do(func() {
		c := fyne.CurrentApp().Driver().CanvasForObject(r.nameEntry)
		if c == nil {
			return
		}
		c.Focus(r.nameEntry)
	})
}

func (r *paneFileListRow) endRename() {
	r.renaming = false
	if r.nameEntry == nil {
		return
	}
	r.nameEntry.onBlur = nil
	r.nameEntry.Hide()
	r.nameCell.Show()
	r.nameCell.Refresh()
	canvas.Refresh(r.nameEntry)
}

func (r *paneFileListRow) updateLocal(rowIndex int, name, size, modified string, isDir, isParent, selected bool) {
	r.rowIndex = rowIndex
	r.selected = selected
	if r.nameCell == nil {
		return
	}
	if r.renaming {
		return
	}
	if isParent {
		r.nameCell.SetContent("↩", "..")
		r.sizeT.SetText("—")
		r.metaT.SetText("—")
	} else {
		r.nameCell.SetContent(fileIcon(isDir), name)
		r.sizeT.SetText(size)
		r.metaT.SetText(modified)
	}
	r.refreshStyle()
}

func (r *paneFileListRow) updateRemote(rowIndex int, name, size, modified string, isDir, isParent, selected bool) {
	r.rowIndex = rowIndex
	r.selected = selected
	if r.nameCell == nil {
		return
	}
	if r.renaming {
		return
	}
	if isParent {
		r.nameCell.SetContent("↩", "..")
		r.sizeT.SetText("—")
		r.metaT.SetText("—")
	} else {
		r.nameCell.SetContent(fileIcon(isDir), name)
		r.sizeT.SetText(size)
		r.metaT.SetText(modified)
	}
	r.refreshStyle()
}

func (r *paneFileListRow) rowBgColor() color.Color {
	if r.selected {
		return colorRowSelected
	}
	if r.hovered {
		return colorRowHover
	}
	if r.rowIndex%2 == 0 {
		return colorPanel
	}
	return colorRowAlt
}

func (r *paneFileListRow) refreshStyle() {
	if r.bg == nil {
		return
	}
	r.bg.FillColor = r.rowBgColor()
	if r.selected {
		r.nameCell.SetColor(colorTextHighlight)
	} else {
		r.nameCell.SetColor(colorForeground)
	}
	canvas.Refresh(r.bg)
	r.nameCell.Refresh()
	r.sizeT.Refresh()
	r.metaT.Refresh()
}

func (r *paneFileListRow) Tapped(*fyne.PointEvent) {
	if r.onPrimary != nil {
		r.onPrimary(r.ctrlDown)
	}
}

func (r *paneFileListRow) TappedSecondary(ev *fyne.PointEvent) {
	if r.onSecondary != nil {
		r.onSecondary(ev)
	}
}

func (r *paneFileListRow) Dragged(e *fyne.DragEvent) {
	if r.onDragged == nil {
		return
	}
	r.dragActive = true
	r.onDragged(e)
}

func (r *paneFileListRow) DragEnd() {
	if r.onDragEnd != nil {
		r.onDragEnd()
	}
	r.dragActive = false
}

func (r *paneFileListRow) MouseDown(ev *desktop.MouseEvent) {
	r.ctrlDown = ev.Modifier&desktop.ControlModifier != 0
	if r.onMouseDown != nil {
		r.onMouseDown()
	}
}

func (r *paneFileListRow) MouseUp(*desktop.MouseEvent) {
	if r.onMouseUp != nil {
		r.onMouseUp()
	}
}

func (r *paneFileListRow) Cursor() desktop.Cursor {
	if r.onDragged == nil {
		return desktop.DefaultCursor
	}
	if r.dragActive {
		return desktop.CrosshairCursor
	}
	if r.hovered {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (r *paneFileListRow) MouseIn(_ *desktop.MouseEvent) {
	r.hovered = true
	r.refreshStyle()
}

func (r *paneFileListRow) MouseMoved(_ *desktop.MouseEvent) {}

func (r *paneFileListRow) MouseOut() {
	r.hovered = false
	r.refreshStyle()
}

func (r *paneFileListRow) CreateRenderer() fyne.WidgetRenderer {
	r.bg = canvas.NewRectangle(r.rowBgColor())
	r.bg.SetMinSize(fyne.NewSize(0, paneRowMinHeight))
	r.nameCell = newPaneFileNameCell(paneRowNameSize, colorForeground)
	r.nameEntry = newPaneRenameEntry()
	r.nameEntry.Hide()
	r.nameEntry.OnSubmitted = func(text string) {
		if r.renaming && r.onRenameCommit != nil {
			r.onRenameCommit(text)
		}
	}
	entryThemed := container.NewThemeOverride(r.nameEntry, newCompactEntryTheme(paneRowNameSize))
	nameCol := container.NewStack(r.nameCell, container.NewMax(entryThemed))

	r.sizeT = newPaneRightText("", colorMuted, paneRowMetaSize)
	r.metaT = newPaneModifiedText("", colorMuted, paneRowMetaSize)
	right := paneFileMetaColumns(r.sizeT, r.metaT)
	row := container.NewBorder(nil, nil, nil, right, nameCol)

	padded := container.New(layout.NewCustomPaddedLayout(paneFileListLeftPad, paneRowPadV, paneFileListRightPad(), paneRowPadV), row)
	content := container.NewStack(r.bg, padded)
	return widget.NewSimpleRenderer(content)
}

func (r *paneFileListRow) MinSize() fyne.Size {
	return fyne.NewSize(0, paneRowMinHeight)
}

var _ fyne.Tappable = (*paneFileListRow)(nil)
var _ fyne.SecondaryTappable = (*paneFileListRow)(nil)
var _ fyne.Draggable = (*paneFileListRow)(nil)
var _ desktop.Hoverable = (*paneFileListRow)(nil)
var _ desktop.Mouseable = (*paneFileListRow)(nil)
var _ desktop.Cursorable = (*paneFileListRow)(nil)
