package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type paneListUnderlay struct {
	widget.BaseWidget
	pane *FilePane
}

func newPaneListUnderlay(p *FilePane) *paneListUnderlay {
	b := &paneListUnderlay{pane: p}
	b.ExtendBaseWidget(b)
	return b
}

func (b *paneListUnderlay) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.NRGBA{A: 0})
	return widget.NewSimpleRenderer(bg)
}

func (b *paneListUnderlay) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (b *paneListUnderlay) Tapped(*fyne.PointEvent) {
	// Primary taps on blank list area only clear selection; must not sit above rows (see list z-order).
	b.pane.clearSelectionQuiet()
}

func (b *paneListUnderlay) TappedSecondary(ev *fyne.PointEvent) {
	b.pane.showContextMenu(ev.AbsolutePosition, -1)
}

var _ fyne.Tappable = (*paneListUnderlay)(nil)
var _ fyne.SecondaryTappable = (*paneListUnderlay)(nil)

type paneListStackLayout struct {
	pane *FilePane
}

func (l paneListStackLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	if len(objects) == 1 {
		return objects[0].MinSize()
	}
	return objects[1].MinSize()
}

func (l paneListStackLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	underlay, list := objects[0], objects[1]
	list.Resize(size)
	list.Move(fyne.NewPos(0, 0))

	contentH := l.pane.listContentHeight()
	gap := size.Height - contentH
	if gap < 1 {
		underlay.Hide()
		return
	}
	underlay.Show()
	underlay.Resize(fyne.NewSize(size.Width, gap))
	underlay.Move(fyne.NewPos(0, contentH))
}

func newPaneListArea(p *FilePane, list *widget.List) fyne.CanvasObject {
	// Underlay first, list second — hit-testing prefers later children so list receives row clicks.
	return container.New(&paneListStackLayout{pane: p}, newPaneListUnderlay(p), list)
}
