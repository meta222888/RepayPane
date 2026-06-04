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
}

func newFileListRow() *fileListRow {
	r := &fileListRow{
		nameLbl: widget.NewLabel(""),
		sizeLbl: widget.NewLabel(""),
		metaLbl: widget.NewLabel(""),
		bg:      canvas.NewRectangle(colorBG),
	}
	r.sizeLbl.Alignment = fyne.TextAlignTrailing
	r.metaLbl.Alignment = fyne.TextAlignTrailing
	r.ExtendBaseWidget(r)
	return r
}

func (r *fileListRow) TappedSecondary(ev *fyne.PointEvent) {
	if r.onSecondary != nil {
		r.onSecondary(ev)
	}
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
	return fyne.NewSize(0, rr.objects[1].MinSize().Height)
}

func (rr *fileListRowRenderer) Refresh() {
	canvas.Refresh(rr.objects[0])
	canvas.Refresh(rr.objects[1])
}

func (rr *fileListRowRenderer) Objects() []fyne.CanvasObject { return rr.objects }
func (rr *fileListRowRenderer) Destroy()                       {}

var _ fyne.SecondaryTappable = (*fileListRow)(nil)
