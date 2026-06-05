package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

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

func (d *dragRegion) MouseDown(e *desktop.MouseEvent) {
	if e.Button != desktop.MouseButtonPrimary {
		return
	}
	winStartCaptionDrag(d.win)
}

func (d *dragRegion) MouseUp(*desktop.MouseEvent)   {}
func (d *dragRegion) MouseIn(*desktop.MouseEvent)   {}
func (d *dragRegion) MouseOut()                     {}

func (d *dragRegion) DoubleTapped(*fyne.PointEvent) {
	toggleMaximizeWindow(d.win)
}

var _ desktop.Mouseable = (*dragRegion)(nil)
var _ fyne.DoubleTappable = (*dragRegion)(nil)
