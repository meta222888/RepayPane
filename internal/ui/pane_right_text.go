package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	paneFileSizeColWidth      float32 = 76
	paneFileModifiedColWidth  float32 = 132
	paneFileEdgeRightGap      float32 = 6
)

// paneRightText right-aligns a single-line canvas.Text within its width.
type paneRightText struct {
	widget.BaseWidget
	text       *canvas.Text
	rightInset float32
}

func newPaneRightText(text string, c color.Color, size float32) *paneRightText {
	return newPaneRightTextInset(text, c, size, 0)
}

func newPaneModifiedText(text string, c color.Color, size float32) *paneRightText {
	inset := paneFileEdgeRightGap - paneRowPadH
	if inset < 0 {
		inset = 0
	}
	return newPaneRightTextInset(text, c, size, inset)
}

func newPaneRightTextInset(text string, c color.Color, size float32, rightInset float32) *paneRightText {
	t := canvas.NewText(text, c)
	t.TextSize = size
	p := &paneRightText{text: t, rightInset: rightInset}
	p.ExtendBaseWidget(p)
	return p
}

func (p *paneRightText) SetText(text string) {
	if p.text.Text == text {
		return
	}
	p.text.Text = text
	p.Refresh()
}

func (p *paneRightText) SetColor(c color.Color) {
	p.text.Color = c
	canvas.Refresh(p.text)
}

func (p *paneRightText) MinSize() fyne.Size {
	return fyne.NewSize(0, p.text.TextSize)
}

type paneRightTextRenderer struct {
	box  *paneRightText
	text *canvas.Text
}

func (r *paneRightTextRenderer) Layout(size fyne.Size) {
	r.box.Resize(size)
	sz, _ := fyne.CurrentApp().Driver().RenderedTextSize(r.text.Text, r.text.TextSize, r.text.TextStyle, r.text.FontSource)
	if sz.Height < r.text.TextSize {
		sz.Height = r.text.TextSize
	}
	r.text.Move(fyne.NewPos(size.Width-sz.Width-r.box.rightInset, (size.Height-sz.Height)/2))
}

func (r *paneRightTextRenderer) MinSize() fyne.Size {
	return r.box.MinSize()
}

func (r *paneRightTextRenderer) Refresh() {
	canvas.Refresh(r.text)
}

func (r *paneRightTextRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *paneRightTextRenderer) Destroy() {}

func (p *paneRightText) CreateRenderer() fyne.WidgetRenderer {
	return &paneRightTextRenderer{box: p, text: p.text}
}

func paneFileMetaColumns(sizeCol, modifiedCol fyne.CanvasObject) fyne.CanvasObject {
	return container.NewHBox(
		fixedWidth(sizeCol, paneFileSizeColWidth),
		fixedWidth(modifiedCol, paneFileModifiedColWidth),
	)
}

func paneFileMetaHeader(sizeLabel, modifiedLabel string) fyne.CanvasObject {
	sizeCol := newPaneRightText(sizeLabel, colorMuted, paneRowMetaSize)
	modifiedCol := newPaneModifiedText(modifiedLabel, colorMuted, paneRowMetaSize)
	return paneFileMetaColumns(sizeCol, modifiedCol)
}
