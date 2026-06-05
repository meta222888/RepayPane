package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type localFileListRow struct {
	widget.BaseWidget

	rowIndex int
	selected bool
	hovered  bool
	isDir    bool

	bg     *canvas.Rectangle
	iconT  *canvas.Text
	nameT  *canvas.Text
	rightT *canvas.Text

	onSecondary func(*fyne.PointEvent)
	onPrimary   func()
	onDouble    func()
}

func newLocalFileListRow() *localFileListRow {
	r := &localFileListRow{}
	r.ExtendBaseWidget(r)
	return r
}

func (r *localFileListRow) update(rowIndex int, name, size, modified string, isDir, isParent, selected bool) {
	r.rowIndex = rowIndex
	r.selected = selected
	r.isDir = isDir || isParent

	if r.iconT == nil {
		return
	}
	if isParent {
		r.iconT.Text = "↩"
		r.iconT.Color = colorAccent
		r.nameT.Text = ".."
		r.rightT.Text = "—   —"
	} else {
		if isDir {
			r.iconT.Text = "[D]"
			r.iconT.Color = colorAccent
		} else {
			r.iconT.Text = "[F]"
			r.iconT.Color = colorMuted
		}
		r.nameT.Text = name
		if size != "—" {
			r.rightT.Text = size + "   " + modified
		} else {
			r.rightT.Text = "—   " + modified
		}
	}
	r.refreshStyle()
}

func (r *localFileListRow) rowBgColor() color.Color {
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

func (r *localFileListRow) refreshStyle() {
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
	canvas.Refresh(r.iconT)
	canvas.Refresh(r.nameT)
	canvas.Refresh(r.rightT)
}

func (r *localFileListRow) Tapped(*fyne.PointEvent) {
	if r.onPrimary != nil {
		r.onPrimary()
	}
}

func (r *localFileListRow) DoubleTapped(*fyne.PointEvent) {
	if r.onDouble != nil {
		r.onDouble()
	}
}

func (r *localFileListRow) TappedSecondary(ev *fyne.PointEvent) {
	if r.onSecondary != nil {
		r.onSecondary(ev)
	}
}

func (r *localFileListRow) MouseIn(_ *desktop.MouseEvent) {
	r.hovered = true
	r.refreshStyle()
}

func (r *localFileListRow) MouseMoved(_ *desktop.MouseEvent) {}

func (r *localFileListRow) MouseOut() {
	r.hovered = false
	r.refreshStyle()
}

func (r *localFileListRow) CreateRenderer() fyne.WidgetRenderer {
	r.bg = canvas.NewRectangle(r.rowBgColor())
	r.iconT = canvas.NewText("[F]", colorMuted)
	r.iconT.TextSize = 11
	r.nameT = canvas.NewText("", colorForeground)
	r.nameT.TextSize = 12
	r.rightT = canvas.NewText("", colorMuted)
	r.rightT.TextSize = 11

	nameBox := container.NewHBox(r.iconT, r.nameT)
	row := container.NewBorder(nil, nil, nil, r.rightT, nameBox)
	content := container.NewStack(r.bg, container.NewPadded(row))
	return widget.NewSimpleRenderer(content)
}

var _ fyne.Tappable = (*localFileListRow)(nil)
var _ fyne.DoubleTappable = (*localFileListRow)(nil)
var _ fyne.SecondaryTappable = (*localFileListRow)(nil)
var _ desktop.Hoverable = (*localFileListRow)(nil)
