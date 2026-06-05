package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	paneRowNameSize  float32 = 14
	paneRowMetaSize  float32 = 13
	paneRowMinHeight float32 = 30
)

type paneFileListRow struct {
	widget.BaseWidget

	remote   bool
	rowIndex int
	selected bool
	hovered  bool

	bg     *canvas.Rectangle
	nameT  *canvas.Text
	sizeT  *canvas.Text
	metaT  *canvas.Text
	rightT *canvas.Text

	onSecondary func(*fyne.PointEvent)
	onPrimary   func()
	onDragged   func(*fyne.DragEvent)
	onDragEnd   func()
	onMouseDown func()
	onMouseUp   func()

	dragActive bool
}

func newPaneFileListRow(remote bool) *paneFileListRow {
	r := &paneFileListRow{remote: remote}
	r.ExtendBaseWidget(r)
	return r
}

func (r *paneFileListRow) updateLocal(rowIndex int, name, size, modified string, isDir, isParent, selected bool) {
	r.rowIndex = rowIndex
	r.selected = selected
	if r.nameT == nil {
		return
	}
	if isParent {
		r.nameT.Text = "↩  .."
		r.rightT.Text = "—   —"
	} else {
		r.nameT.Text = fileIcon(isDir) + "  " + name
		if size != "—" {
			r.rightT.Text = size + "   " + modified
		} else {
			r.rightT.Text = "—   " + modified
		}
	}
	r.refreshStyle()
}

func (r *paneFileListRow) updateRemote(rowIndex int, name, size, modified string, isDir, isParent, selected bool) {
	r.rowIndex = rowIndex
	r.selected = selected
	if r.nameT == nil {
		return
	}
	if isParent {
		r.nameT.Text = "↩  .."
		r.sizeT.Text = "—"
		r.metaT.Text = "—"
	} else {
		r.nameT.Text = fileIcon(isDir) + "  " + name
		r.sizeT.Text = size
		r.metaT.Text = modified
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
		r.nameT.Color = colorTextHighlight
	} else {
		r.nameT.Color = colorForeground
	}
	canvas.Refresh(r.bg)
	canvas.Refresh(r.nameT)
	if r.remote {
		canvas.Refresh(r.sizeT)
		canvas.Refresh(r.metaT)
	} else {
		canvas.Refresh(r.rightT)
	}
}

func (r *paneFileListRow) Tapped(*fyne.PointEvent) {
	if r.onPrimary != nil {
		r.onPrimary()
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

func (r *paneFileListRow) MouseDown(*desktop.MouseEvent) {
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
	r.nameT = canvas.NewText("", colorForeground)
	r.nameT.TextSize = paneRowNameSize

	var row fyne.CanvasObject
	if r.remote {
		r.sizeT = canvas.NewText("", colorMuted)
		r.sizeT.TextSize = paneRowMetaSize
		r.metaT = canvas.NewText("", colorMuted)
		r.metaT.TextSize = paneRowMetaSize
		right := container.NewHBox(fixedWidth(r.metaT, 128), fixedWidth(r.sizeT, 72))
		row = container.NewBorder(nil, nil, r.nameT, right, nil)
	} else {
		r.rightT = canvas.NewText("", colorMuted)
		r.rightT.TextSize = paneRowMetaSize
		row = container.NewBorder(nil, nil, nil, r.rightT, r.nameT)
	}

	content := container.NewStack(r.bg, container.NewPadded(row))
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
