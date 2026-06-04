package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type dragRegion struct {
	widget.BaseWidget
	win      fyne.Window
	inner    fyne.CanvasObject
	dragging bool
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
	if winBeginDrag(d) {
		d.dragging = true
	}
}

func (d *dragRegion) MouseUp(*desktop.MouseEvent) {
	if d.dragging {
		winEndDrag()
		d.dragging = false
	}
}

func (d *dragRegion) MouseIn(*desktop.MouseEvent) {}
func (d *dragRegion) MouseOut()                   {}
