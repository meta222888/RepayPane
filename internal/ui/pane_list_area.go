package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const paneBlankMinHeight float32 = 48

type paneBlankPad struct {
	widget.BaseWidget
	pane        *FilePane
	transparent bool
}

func newPaneBlankPad(p *FilePane) *paneBlankPad {
	b := &paneBlankPad{pane: p}
	b.ExtendBaseWidget(b)
	return b
}

func newPaneListUnderlay(p *FilePane) *paneBlankPad {
	b := &paneBlankPad{pane: p, transparent: true}
	b.ExtendBaseWidget(b)
	return b
}

func (b *paneBlankPad) CreateRenderer() fyne.WidgetRenderer {
	var fill color.Color = colorPanel
	if b.transparent {
		fill = color.NRGBA{A: 0}
	}
	bg := canvas.NewRectangle(fill)
	return widget.NewSimpleRenderer(bg)
}

func (b *paneBlankPad) MinSize() fyne.Size {
	if b.transparent {
		return fyne.NewSize(0, 0)
	}
	return fyne.NewSize(0, paneBlankMinHeight)
}

func (b *paneBlankPad) Tapped(*fyne.PointEvent) {
	b.pane.clearSelectionQuiet()
}

func (b *paneBlankPad) TappedSecondary(ev *fyne.PointEvent) {
	b.pane.clearSelectionQuiet()
	b.pane.showContextMenu(ev.AbsolutePosition)
}

var _ fyne.Tappable = (*paneBlankPad)(nil)
var _ fyne.SecondaryTappable = (*paneBlankPad)(nil)

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
	listStack := container.New(&paneListStackLayout{pane: p}, list, newPaneListUnderlay(p))
	return container.NewBorder(nil, newPaneBlankPad(p), nil, nil, listStack)
}
