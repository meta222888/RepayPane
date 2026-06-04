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
	nameLbl *widget.Label
	subLbl  *widget.Label
}

func newConnectPickerRow() *connectPickerRow {
	r := &connectPickerRow{
		bg:      canvas.NewRectangle(colorBG),
		nameLbl: widget.NewLabel(""),
		subLbl:  widget.NewLabel(""),
	}
	r.subLbl.Importance = widget.MediumImportance
	r.ExtendBaseWidget(r)
	return r
}

func (r *connectPickerRow) update(name, subtitle string, selected bool) {
	r.nameLbl.SetText(name)
	r.subLbl.SetText(subtitle)
	if selected {
		r.bg.FillColor = colorRowSelected
	} else {
		r.bg.FillColor = colorBG
	}
	canvas.Refresh(r.bg)
}

func (r *connectPickerRow) CreateRenderer() fyne.WidgetRenderer {
	text := container.NewVBox(r.nameLbl, r.subLbl)
	content := container.NewPadded(text)
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
	return fyne.NewSize(0, rr.objects[1].MinSize().Height)
}

func (rr *connectPickerRowRenderer) Refresh() {
	canvas.Refresh(rr.objects[0])
	canvas.Refresh(rr.objects[1])
}

func (rr *connectPickerRowRenderer) Objects() []fyne.CanvasObject { return rr.objects }
func (rr *connectPickerRowRenderer) Destroy()                       {}
