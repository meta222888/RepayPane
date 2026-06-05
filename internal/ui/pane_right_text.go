package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	paneFileSizeColWidth     float32 = 76
	paneFileModifiedColWidth float32 = 138
	paneFileEdgeRightGap     float32 = 6
	paneFileListScrollGutter float32 = 12
)

func paneFileListRightPad() float32 {
	return paneFileEdgeRightGap + paneFileListScrollGutter
}

// paneRightText right-aligns a single-line canvas.Text within its width.
type paneRightText struct {
	widget.BaseWidget
	text *canvas.Text
}

func newPaneRightText(text string, c color.Color, size float32) *paneRightText {
	t := canvas.NewText(text, c)
	t.TextSize = size
	p := &paneRightText{text: t}
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
	x := size.Width - sz.Width
	if x < 0 {
		x = 0
	}
	r.text.Move(fyne.NewPos(x, (size.Height-sz.Height)/2))
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
	modifiedCol := newPaneRightText(modifiedLabel, colorMuted, paneRowMetaSize)
	return paneFileMetaColumns(sizeCol, modifiedCol)
}

func paneFileListHeaderRow(nameCol, metaCols fyne.CanvasObject) fyne.CanvasObject {
	row := container.NewBorder(nil, nil, nil, metaCols, nameCol)
	extra := paneFileListRightPad() - paneRowPadH
	if extra > 0 {
		row = container.New(layout.NewCustomPaddedLayout(0, 0, extra, 0), row)
	}
	return row
}
