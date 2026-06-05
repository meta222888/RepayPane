package ui

import (
	"runtime"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// NewMainWindow creates a borderless main window when the desktop driver supports it.
func NewMainWindow(a fyne.App) fyne.Window {
	if drv, ok := a.Driver().(desktop.Driver); ok {
		w := drv.CreateSplashWindow()
		w.Resize(fyne.NewSize(1280, 760))
		w.SetTitle(i18n.T(i18n.KeyAppTitle))
		w.SetPadded(false)
		w.SetMaster()
		return w
	}
	w := a.NewWindow(i18n.T(i18n.KeyAppTitle))
	w.Resize(fyne.NewSize(1280, 760))
	w.SetPadded(false)
	w.SetMaster()
	return w
}

func minimizeWindow(w fyne.Window) {
	if runtime.GOOS == "windows" {
		winMinimizeWindows(w)
	}
}

func toggleMaximizeWindow(w fyne.Window) {
	if runtime.GOOS == "windows" {
		winToggleMaximize(w)
	}
}

func closeWindow(w fyne.Window) {
	w.Close()
}

func wrapWindowResize(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	return wrapWindowResizePlatform(w, content)
}
