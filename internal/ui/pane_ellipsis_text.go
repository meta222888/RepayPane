package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	paneLocalMetaColWidth  float32 = 200
	paneRemoteMetaColWidth float32 = 128
	paneRemoteSizeColWidth float32 = 72
)

// paneEllipsisText draws a single-line label truncated with "…" to fit its width.
type paneEllipsisText struct {
	widget.BaseWidget
	fullText string
	textSize float32
	textCol  color.Color
	lastW    float32
}

func newPaneEllipsisText(textSize float32, c color.Color) *paneEllipsisText {
	t := &paneEllipsisText{textSize: textSize, textCol: c}
	t.ExtendBaseWidget(t)
	return t
}

func (t *paneEllipsisText) SetText(text string) {
	if t.fullText == text {
		return
	}
	t.fullText = text
	t.lastW = 0
	t.Refresh()
}

func (t *paneEllipsisText) SetColor(c color.Color) {
	t.textCol = c
	t.Refresh()
}

func (t *paneEllipsisText) MinSize() fyne.Size {
	return fyne.NewSize(0, t.textSize)
}

type paneEllipsisTextRenderer struct {
	label *paneEllipsisText
	text  *canvas.Text
}

func (r *paneEllipsisTextRenderer) Layout(size fyne.Size) {
	r.label.Resize(size)
	if size.Width != r.label.lastW {
		r.label.lastW = size.Width
		r.text.Text = ellipsizeText(r.label.fullText, size.Width, r.label.textSize, r.text.TextStyle, r.text.FontSource)
	}
	r.text.Color = r.label.textCol
	textH := r.text.MinSize().Height
	if textH < r.label.textSize {
		textH = r.label.textSize
	}
	r.text.Resize(fyne.NewSize(size.Width, textH))
	r.text.Move(fyne.NewPos(0, (size.Height-textH)/2))
}

func (r *paneEllipsisTextRenderer) MinSize() fyne.Size {
	return r.label.MinSize()
}

func (r *paneEllipsisTextRenderer) Refresh() {
	r.label.lastW = 0
	r.text.Color = r.label.textCol
	canvas.Refresh(r.text)
}

func (r *paneEllipsisTextRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *paneEllipsisTextRenderer) Destroy() {}

func (t *paneEllipsisText) CreateRenderer() fyne.WidgetRenderer {
	txt := canvas.NewText("", t.textCol)
	txt.TextSize = t.textSize
	return &paneEllipsisTextRenderer{label: t, text: txt}
}

func ellipsizeText(text string, maxWidth float32, textSize float32, style fyne.TextStyle, font fyne.Resource) string {
	if text == "" || maxWidth <= 0 {
		return ""
	}
	driver := fyne.CurrentApp().Driver()
	full, _ := driver.RenderedTextSize(text, textSize, style, font)
	if full.Width <= maxWidth {
		return text
	}
	const ellipsis = "…"
	ell, _ := driver.RenderedTextSize(ellipsis, textSize, style, font)
	limit := maxWidth - ell.Width
	if limit <= 0 {
		return ellipsis
	}
	runes := []rune(text)
	lo, hi := 0, len(runes)
	best := 0
	for lo <= hi {
		mid := (lo + hi) / 2
		prefix := string(runes[:mid])
		w, _ := driver.RenderedTextSize(prefix, textSize, style, font)
		if w.Width <= limit {
			best = mid
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return string(runes[:best]) + ellipsis
}
