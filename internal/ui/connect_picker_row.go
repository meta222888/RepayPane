package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type connectPickerRow struct {
	widget.BaseWidget

	bg      *canvas.Rectangle
	lineLbl *widget.Label

	onPrimary func()
	onDouble  func()
}

func newConnectPickerRow() *connectPickerRow {
	r := &connectPickerRow{
		bg:      canvas.NewRectangle(colorPanel),
		lineLbl: widget.NewLabel(""),
	}
	r.ExtendBaseWidget(r)
	return r
}

func (r *connectPickerRow) update(name, subtitle string, selected bool) {
	if subtitle != "" {
		r.lineLbl.SetText(name + "  ·  " + subtitle)
	} else {
		r.lineLbl.SetText(name)
	}
	if selected {
		r.bg.FillColor = colorRowSelected
	} else {
		r.bg.FillColor = colorPanel
	}
	canvas.Refresh(r.bg)
}

func (r *connectPickerRow) Tapped(*fyne.PointEvent) {
	if r.onPrimary != nil {
		r.onPrimary()
	}
}

func (r *connectPickerRow) DoubleTapped(*fyne.PointEvent) {
	if r.onDouble != nil {
		r.onDouble()
	}
}

func (r *connectPickerRow) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewPadded(r.lineLbl)
	return &connectPickerRowRenderer{
		row:     r,
		objects: []fyne.CanvasObject{r.bg, content},
	}
}

type connectPickerRowRenderer struct {
	row     *connectPickerRow
	objects []fyne.CanvasObject
}

func (rr *connectPickerRowRenderer) Layout(size fyne.Size) {
	rr.objects[0].Resize(size)
	rr.objects[0].Move(fyne.NewPos(0, 0))
	rr.objects[1].Resize(size)
	rr.objects[1].Move(fyne.NewPos(0, 0))
}

func (rr *connectPickerRowRenderer) MinSize() fyne.Size {
	h := rr.objects[1].MinSize().Height
	return fyne.NewSize(0, h)
}

func (rr *connectPickerRowRenderer) Refresh() {
	canvas.Refresh(rr.objects[0])
	canvas.Refresh(rr.objects[1])
}

func (rr *connectPickerRowRenderer) Objects() []fyne.CanvasObject { return rr.objects }
func (rr *connectPickerRowRenderer) Destroy()                       {}

var _ fyne.Tappable = (*connectPickerRow)(nil)
var _ fyne.DoubleTappable = (*connectPickerRow)(nil)
