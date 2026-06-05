package ui

import (
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const dragStartThreshold = 4

type dragRegion struct {
	widget.BaseWidget
	win         fyne.Window
	inner       fyne.CanvasObject
	lastUp      time.Time
	pressed     bool
	dragStarted bool
	dragMoved   float32
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
	now := time.Now()
	if !d.lastUp.IsZero() && now.Sub(d.lastUp) <= doubleClickInterval() {
		d.lastUp = time.Time{}
		d.resetPress()
		toggleMaximizeWindow(d.win)
		return
	}
	d.pressed = true
	d.dragStarted = false
	d.dragMoved = 0
}

func (d *dragRegion) MouseUp(*desktop.MouseEvent) {
	d.resetPress()
	d.lastUp = time.Now()
}

func (d *dragRegion) Dragged(e *fyne.DragEvent) {
	if !d.pressed || d.dragStarted {
		return
	}
	d.dragMoved += float32(math.Max(math.Abs(float64(e.Dragged.DX)), math.Abs(float64(e.Dragged.DY))))
	if d.dragMoved < dragStartThreshold {
		return
	}
	d.dragStarted = true
	winStartCaptionDrag(d.win)
}

func (d *dragRegion) DragEnd() {
	d.resetPress()
}

func (d *dragRegion) resetPress() {
	d.pressed = false
	d.dragStarted = false
	d.dragMoved = 0
}

func (d *dragRegion) MouseIn(*desktop.MouseEvent) {}
func (d *dragRegion) MouseOut()                   {}

var _ desktop.Mouseable = (*dragRegion)(nil)
var _ fyne.Draggable = (*dragRegion)(nil)
