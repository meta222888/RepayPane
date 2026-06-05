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
	b.pane.dismissContextMenu()
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
	return objects[0].MinSize()
}

func (l paneListStackLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	list, underlay := objects[0], objects[1]
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
	return container.New(&paneListStackLayout{pane: p}, list, newPaneListUnderlay(p))
}
