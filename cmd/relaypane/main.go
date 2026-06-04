package main

import (
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("com.relaypane.app")
	ui.ApplyTheme(a)
	w := a.NewWindow(i18n.T(i18n.KeyAppTitle))
	w.Resize(fyne.NewSize(1280, 760))
	w.SetPadded(false)
	w.SetMaster()

	ui.NewApp(a, w)
	w.ShowAndRun()
}
