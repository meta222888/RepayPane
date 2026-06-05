package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type fileListRow struct {
	widget.BaseWidget

	nameLbl *widget.Label
	sizeLbl *widget.Label
	metaLbl *widget.Label
	bg      *canvas.Rectangle

	onSecondary func(*fyne.PointEvent)
	onPrimary   func()
	onDouble    func()
}

func newFileListRow() *fileListRow {
	r := &fileListRow{
		nameLbl: widget.NewLabel(""),
		sizeLbl: widget.NewLabel(""),
		metaLbl: widget.NewLabel(""),
		bg:      canvas.NewRectangle(colorPanel),
	}
	r.sizeLbl.Alignment = fyne.TextAlignTrailing
	r.metaLbl.Alignment = fyne.TextAlignTrailing
	r.ExtendBaseWidget(r)
	return r
}

func (r *fileListRow) Tapped(*fyne.PointEvent) {
	if r.onPrimary != nil {
		r.onPrimary()
	}
}

func (r *fileListRow) DoubleTapped(*fyne.PointEvent) {
	if r.onDouble != nil {
		r.onDouble()
	}
}

func (r *fileListRow) TappedSecondary(ev *fyne.PointEvent) {
	if r.onSecondary != nil {
		r.onSecondary(ev)
	}
}

func (r *fileListRow) setRowStyle(rowIndex int, selected bool) {
	if selected {
		r.bg.FillColor = colorRowSelected
	} else if rowIndex%2 == 0 {
		r.bg.FillColor = colorPanel
	} else {
		r.bg.FillColor = colorRowAlt
	}
	canvas.Refresh(r.bg)
}

func (r *fileListRow) setSelected(selected bool) {
	if selected {
		r.bg.FillColor = colorRowSelected
	} else {
		r.bg.FillColor = colorPanel
	}
	canvas.Refresh(r.bg)
}

func (r *fileListRow) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewBorder(nil, nil, r.nameLbl, container.NewHBox(fixedWidth(r.metaLbl, 128), fixedWidth(r.sizeLbl, 72)), nil)
	return &fileListRowRenderer{
		row:     r,
		objects: []fyne.CanvasObject{r.bg, content},
	}
}

type fileListRowRenderer struct {
	row     *fileListRow
	objects []fyne.CanvasObject
}

func (rr *fileListRowRenderer) Layout(size fyne.Size) {
	rr.objects[0].Resize(size)
	rr.objects[0].Move(fyne.NewPos(0, 0))
	rr.objects[1].Resize(size)
	rr.objects[1].Move(fyne.NewPos(0, 0))
}

func (rr *fileListRowRenderer) MinSize() fyne.Size {
	h := float32(26)
	if mh := rr.objects[1].MinSize().Height; mh > h {
		h = mh
	}
	return fyne.NewSize(0, h)
}

func (rr *fileListRowRenderer) Refresh() {
	canvas.Refresh(rr.objects[0])
	canvas.Refresh(rr.objects[1])
}

func (rr *fileListRowRenderer) Objects() []fyne.CanvasObject { return rr.objects }
func (rr *fileListRowRenderer) Destroy()                       {}

var _ fyne.SecondaryTappable = (*fileListRow)(nil)
var _ fyne.Tappable = (*fileListRow)(nil)
var _ fyne.DoubleTappable = (*fileListRow)(nil)
