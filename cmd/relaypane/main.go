package main

import (
	"github.com/relaypane/relaypane/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("com.relaypane.app")
	a.SetIcon(nil)
	w := a.NewWindow("RelayPane")
	w.Resize(fyne.NewSize(1100, 700))
	w.SetMaster()

	ui.NewApp(a, w)
	w.ShowAndRun()
}
