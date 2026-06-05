package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// paneRowBand fixes chrome/header rows to PaneRowHeight so widgets cannot stretch the band taller.
type paneRowBand struct {
	widget.BaseWidget
	content fyne.CanvasObject
}

func newPaneRowBand(content fyne.CanvasObject) *paneRowBand {
	b := &paneRowBand{content: content}
	b.ExtendBaseWidget(b)
	return b
}

func (b *paneRowBand) MinSize() fyne.Size {
	return fyne.NewSize(0, PaneRowHeight)
}

type paneRowBandRenderer struct {
	band    *paneRowBand
	bg      *canvas.Rectangle
	content fyne.CanvasObject
}

func (r *paneRowBandRenderer) Layout(size fyne.Size) {
	size.Height = PaneRowHeight
	r.bg.Resize(size)
	innerW := size.Width - paneRowPadH*2
	innerH := size.Height - paneRowPadV*2
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}
	r.content.Resize(fyne.NewSize(innerW, innerH))
	r.content.Move(fyne.NewPos(paneRowPadH, paneRowPadV))
}

func (r *paneRowBandRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, PaneRowHeight)
}

func (r *paneRowBandRenderer) Refresh() {
	r.bg.FillColor = colorPanelHeader
	canvas.Refresh(r.bg)
}

func (r *paneRowBandRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.content}
}

func (r *paneRowBandRenderer) Destroy() {}

func (b *paneRowBand) CreateRenderer() fyne.WidgetRenderer {
	r := &paneRowBandRenderer{
		band:    b,
		bg:      canvas.NewRectangle(colorPanelHeader),
		content: b.content,
	}
	r.bg.SetMinSize(fyne.NewSize(0, PaneRowHeight))
	return r
}

func paneBand(content fyne.CanvasObject) fyne.CanvasObject {
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	return container.NewVBox(newPaneRowBand(content), line)
}
