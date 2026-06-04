package ui

import (
	"fyne.io/fyne/v2/dialog"
)

func dialogShow(a *App, title, msg string) {
	dialog.ShowInformation(title, msg, a.window)
}

func dialogShowError(a *App, err error) {
	dialog.ShowError(err, a.window)
}
