package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Draggable = (*dragRegion)(nil)

type dragRegion struct {
	widget.BaseWidget
	win   fyne.Window
	inner fyne.CanvasObject
}

func newDragRegion(win fyne.Window, inner fyne.CanvasObject) *dragRegion {
	d := &dragRegion{win: win, inner: inner}
	d.ExtendBaseWidget(d)
	return d
}

func (d *dragRegion) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.inner)
}

func (d *dragRegion) Dragged(e *fyne.DragEvent) {
	moveWindowBy(d.win, d.win.Canvas().Scale(), e.Dragged)
}

func (d *dragRegion) DragEnd() {}

func (d *dragRegion) MouseDown(*desktop.MouseEvent) {}

func (d *dragRegion) MouseUp(*desktop.MouseEvent) {}
