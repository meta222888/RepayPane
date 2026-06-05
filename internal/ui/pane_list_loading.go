package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func newPaneLoadingHint() fyne.CanvasObject {
	lbl := canvas.NewText(i18n.T(i18n.KeyPaneListingLoading), colorMuted)
	lbl.TextSize = paneRowMetaSize
	bg := canvas.NewRectangle(colorInput)
	inner := container.New(layout.NewCustomPaddedLayout(6, 0, 6, 0), lbl)
	chip := container.NewStack(bg, inner)
	chip.Hide()
	return chip
}

func setPaneLoadingHint(hint fyne.CanvasObject, text string, visible bool) {
	if hint == nil {
		return
	}
	if lbl, ok := findCanvasText(hint); ok {
		lbl.Text = text
		canvas.Refresh(lbl)
	} else if wl, ok := findWidgetLabel(hint); ok {
		wl.SetText(text)
	}
	if visible {
		hint.Show()
	} else {
		hint.Hide()
	}
	canvas.Refresh(hint)
}

func findCanvasText(obj fyne.CanvasObject) (*canvas.Text, bool) {
	switch v := obj.(type) {
	case *canvas.Text:
		return v, true
	case *fyne.Container:
		for _, child := range v.Objects {
			if t, ok := findCanvasText(child); ok {
				return t, true
			}
		}
	}
	return nil, false
}

func findWidgetLabel(obj fyne.CanvasObject) (*widget.Label, bool) {
	switch v := obj.(type) {
	case *widget.Label:
		return v, true
	case *fyne.Container:
		for _, child := range v.Objects {
			if lbl, ok := findWidgetLabel(child); ok {
				return lbl, true
			}
		}
	}
	return nil, false
}
