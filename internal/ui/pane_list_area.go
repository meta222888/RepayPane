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
	b.pane.clearSelection()
}

func (b *paneBlankPad) TappedSecondary(ev *fyne.PointEvent) {
	b.pane.clearSelection()
	b.pane.showContextMenu(ev.AbsolutePosition)
}

var _ fyne.Tappable = (*paneBlankPad)(nil)
var _ fyne.SecondaryTappable = (*paneBlankPad)(nil)

type paneListVBoxLayout struct {
	pane *FilePane
}

func (l paneListVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, paneBlankMinHeight)
	}
	w := float32(0)
	for _, o := range objects {
		m := o.MinSize()
		if m.Width > w {
			w = m.Width
		}
	}
	return fyne.NewSize(w, l.pane.listContentHeight()+paneBlankMinHeight)
}

func (l paneListVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	list, blank := objects[0], objects[1]
	contentH := l.pane.listContentHeight()
	blankH := size.Height - contentH
	if blankH < paneBlankMinHeight {
		blankH = paneBlankMinHeight
	}
	listH := size.Height - blankH
	if listH < 0 {
		listH = 0
	}
	list.Resize(fyne.NewSize(size.Width, listH))
	list.Move(fyne.NewPos(0, 0))
	blank.Resize(fyne.NewSize(size.Width, blankH))
	blank.Move(fyne.NewPos(0, listH))
}

func newPaneListArea(p *FilePane, list *widget.List) fyne.CanvasObject {
	return container.New(&paneListVBoxLayout{pane: p}, list, newPaneBlankPad(p))
}
