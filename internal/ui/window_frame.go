package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
)

// NewMainWindow creates a standard decorated window (native title bar and resize on Windows).
func NewMainWindow(a fyne.App) fyne.Window {
	w := a.NewWindow(i18n.T(i18n.KeyAppTitle))
	w.Resize(fyne.NewSize(1280, 760))
	w.SetPadded(false)
	w.SetMaster()
	return w
}

func closeWindow(w fyne.Window) {
	w.Close()
}
