package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
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

func (r *paneEllipsisTextRenderer) applyDisplayedText(width float32) {
	if width <= 0 {
		return
	}
	r.text.Text = ellipsizeText(r.label.fullText, width, r.label.textSize, r.text.TextStyle, r.text.FontSource)
	r.label.lastW = width
}

func (r *paneEllipsisTextRenderer) Layout(size fyne.Size) {
	r.label.Resize(size)
	r.applyDisplayedText(size.Width)
	r.text.Color = r.label.textCol
	sz, _ := fyne.CurrentApp().Driver().RenderedTextSize(r.text.Text, r.label.textSize, r.text.TextStyle, r.text.FontSource)
	if sz.Height < r.label.textSize {
		sz.Height = r.label.textSize
	}
	if sz.Width < 1 {
		sz.Width = 1
	}
	r.text.Resize(sz)
	r.text.Move(fyne.NewPos(0, paneRowTextCenterY(size.Height, sz.Height)))
}

func (r *paneEllipsisTextRenderer) MinSize() fyne.Size {
	return r.label.MinSize()
}

func (r *paneEllipsisTextRenderer) Refresh() {
	r.text.Color = r.label.textCol
	r.applyDisplayedText(r.label.Size().Width)
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
