package main

import (
	"github.com/relaypane/relaypane/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("com.relaypane.app")
	ui.ApplyTheme(a)
	w := ui.NewMainWindow(a)

	ui.NewApp(a, w)
	w.ShowAndRun()
}