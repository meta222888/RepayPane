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

func moveWindowBy(w fyne.Window, scale float32, delta fyne.Delta) {
	if runtime.GOOS == "windows" {
		winMoveWindows(w, scale, delta)
	}
}

func minimizeWindow(w fyne.Window) {
	if runtime.GOOS == "windows" {
		winMinimizeWindows(w)
	}
}

func toggleMaximizeWindow(w fyne.Window, maximized *bool) {
	if runtime.GOOS != "windows" {
		return
	}
	if *maximized {
		winRestoreWindows(w)
		*maximized = false
		return
	}
	winMaximizeWindows(w)
	*maximized = true
}

func closeWindow(w fyne.Window) {
	w.Close()
}
