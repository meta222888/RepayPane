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
	paneFileSizeColWidth        float32 = 76
	paneFileModifiedColWidth    float32 = 138
	paneFileEdgeRightGap        float32 = 8
	paneFileListScrollGutter    float32 = 24
	paneFileListLeftPad         float32 = 0
	paneFileListLeftNudge       float32 = 16
	paneModifiedColRightInset   float32 = 8
)

func paneFileListRightPad() float32 {
	return paneFileEdgeRightGap + paneFileListScrollGutter
}

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
	return newPaneRightTextInset(text, c, size, paneModifiedColRightInset)
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
	x := size.Width - sz.Width - r.box.rightInset
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
	modifiedCol := newPaneModifiedText(modifiedLabel, colorMuted, paneRowMetaSize)
	return paneFileMetaColumns(sizeCol, modifiedCol)
}

func paneFileListHeaderRow(nameCol, metaCols fyne.CanvasObject) fyne.CanvasObject {
	row := container.NewBorder(nil, nil, nil, metaCols, nameCol)
	extra := paneFileListRightPad() - paneFileListLeftPad
	if extra > 0 {
		row = container.New(layout.NewCustomPaddedLayout(0, 0, extra, 0), row)
	}
	return row
}

type paneFileListNudgeLayout struct{}

func (paneFileListNudgeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	return objects[0].MinSize()
}

func (paneFileListNudgeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	child := objects[0]
	child.Resize(fyne.NewSize(size.Width+paneFileListLeftNudge, size.Height))
	child.Move(fyne.NewPos(-paneFileListLeftNudge, 0))
}

func paneFileListNudgeWrap(obj fyne.CanvasObject) fyne.CanvasObject {
	return container.New(paneFileListNudgeLayout{}, obj)
}
