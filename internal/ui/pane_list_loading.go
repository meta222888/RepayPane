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
	lbl := widget.NewLabel(i18n.T(i18n.KeyPaneListingLoading))
	lbl.Importance = widget.MediumImportance
	bg := canvas.NewRectangle(colorInput)
	inner := container.New(layout.NewCustomPaddedLayout(8, 2, 8, 2), lbl)
	chip := container.NewStack(bg, inner)
	chip.Hide()
	return chip
}

func setPaneLoadingHint(hint fyne.CanvasObject, text string, visible bool) {
	if hint == nil {
		return
	}
	if lbl, ok := findWidgetLabel(hint); ok {
		lbl.SetText(text)
	}
	if visible {
		hint.Show()
	} else {
		hint.Hide()
	}
	canvas.Refresh(hint)
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
