package main

import (
	"github.com/relaypane/relaypane/internal/assets"
	"github.com/relaypane/relaypane/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("com.relaypane.app")
	icon := assets.LogoResource()
	a.SetIcon(icon)
	w := ui.NewMainWindow(a)
	w.SetIcon(icon)
	ui.NewApp(a, w)
	w.ShowAndRun()
}