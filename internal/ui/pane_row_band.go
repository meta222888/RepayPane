package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// paneRowBand fixes chrome/header rows to PaneRowHeight (same as file list rows).
type paneRowBand struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	padLeft  float32
	padRight float32
}

func newPaneRowBand(content fyne.CanvasObject, padLeft, padRight float32) *paneRowBand {
	b := &paneRowBand{content: content, padLeft: padLeft, padRight: padRight}
	b.ExtendBaseWidget(b)
	return b
}

func (b *paneRowBand) MinSize() fyne.Size {
	w := b.content.MinSize().Width
	if w < 1 {
		w = 1
	}
	return fyne.NewSize(w, PaneRowHeight)
}

type paneRowBandRenderer struct {
	band    *paneRowBand
	bg      *canvas.Rectangle
	line    *canvas.Rectangle
	content fyne.CanvasObject
}

func (r *paneRowBandRenderer) Layout(size fyne.Size) {
	size.Height = PaneRowHeight
	r.band.Resize(size)

	bodyH := PaneRowHeight - paneBandLineH
	r.bg.Resize(fyne.NewSize(size.Width, bodyH))
	r.bg.Move(fyne.NewPos(0, 0))

	r.line.Resize(fyne.NewSize(size.Width, paneBandLineH))
	r.line.Move(fyne.NewPos(0, bodyH))

	innerW := size.Width - paneRowPadH*2
	innerH := paneBandInnerHeight()
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
	return r.band.MinSize()
}

func (r *paneRowBandRenderer) Refresh() {
	r.bg.FillColor = colorPanelHeader
	canvas.Refresh(r.bg)
	canvas.Refresh(r.line)
}

func (r *paneRowBandRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.content, r.line}
}

func (r *paneRowBandRenderer) Destroy() {}

func (b *paneRowBand) CreateRenderer() fyne.WidgetRenderer {
	return &paneRowBandRenderer{
		band:    b,
		bg:      canvas.NewRectangle(colorPanelHeader),
		line:    canvas.NewRectangle(colorBorder),
		content: b.content,
	}
}

func paneBand(content fyne.CanvasObject) fyne.CanvasObject {
	return newPaneRowBand(content, paneRowPadH, paneRowPadH)
}

func paneFileListBand(content fyne.CanvasObject) fyne.CanvasObject {
	return newPaneRowBand(content, paneFileListLeftPad, paneRowPadH)
}
