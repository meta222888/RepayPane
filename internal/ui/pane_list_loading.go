package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// paneListLoadingOverlay covers the remote list while listing; must not leave an opaque layer after Hide.
type paneListLoadingOverlay struct {
	widget.BaseWidget
	active bool
	lbl    *widget.Label
}

func newPaneListLoadingOverlay() *paneListLoadingOverlay {
	o := &paneListLoadingOverlay{
		lbl: widget.NewLabel(i18n.T(i18n.KeyPaneListingLoading)),
	}
	o.ExtendBaseWidget(o)
	o.Hide()
	return o
}

func (o *paneListLoadingOverlay) setActive(v bool) {
	o.active = v
	if v {
		o.Show()
	} else {
		o.Hide()
	}
	o.Refresh()
}

func (o *paneListLoadingOverlay) setText(text string) {
	o.lbl.SetText(text)
}

type paneListLoadingRenderer struct {
	overlay *paneListLoadingOverlay
	objects []fyne.CanvasObject
}

func (r *paneListLoadingRenderer) Layout(size fyne.Size) {
	if !r.overlay.active || len(r.objects) == 0 {
		return
	}
	for _, obj := range r.objects {
		obj.Resize(size)
		obj.Move(fyne.NewPos(0, 0))
	}
}

func (r *paneListLoadingRenderer) MinSize() fyne.Size {
	if !r.overlay.active {
		return fyne.NewSize(0, 0)
	}
	return r.overlay.lbl.MinSize()
}

func (r *paneListLoadingRenderer) Refresh() {
	if !r.overlay.active {
		r.objects = nil
		return
	}
	bg := canvas.NewRectangle(colorPanel)
	r.objects = []fyne.CanvasObject{
		container.NewStack(bg, container.NewCenter(r.overlay.lbl)),
	}
}

func (r *paneListLoadingRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *paneListLoadingRenderer) Destroy() {}

func (o *paneListLoadingOverlay) CreateRenderer() fyne.WidgetRenderer {
	r := &paneListLoadingRenderer{overlay: o}
	r.Refresh()
	return r
}
