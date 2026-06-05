package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const paneBlankMinHeight float32 = 48

type paneBlankPad struct {
	widget.BaseWidget
	pane *FilePane
}

func newPaneBlankPad(p *FilePane) *paneBlankPad {
	b := &paneBlankPad{pane: p}
	b.ExtendBaseWidget(b)
	return b
}

func (b *paneBlankPad) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(colorPanel)
	return widget.NewSimpleRenderer(bg)
}

func (b *paneBlankPad) MinSize() fyne.Size {
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

func newPaneListArea(p *FilePane, list *widget.List) fyne.CanvasObject {
	return container.NewBorder(nil, newPaneBlankPad(p), nil, nil, list)
}
