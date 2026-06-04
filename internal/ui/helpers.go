package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)

func fyneCurrent() fyne.App {
	return fyne.CurrentApp()
}

func dialogShow(a *App, title, msg string) {
	dialog.ShowInformation(title, msg, a.window)
}

func dialogShowError(a *App, err error) {
	dialog.ShowError(err, a.window)
}

func withBackground(obj fyne.CanvasObject, bg color.Color) fyne.CanvasObject {
	rect := canvas.NewRectangle(bg)
	return container.NewStack(rect, obj)
}

func withPanelHeader(obj fyne.CanvasObject) fyne.CanvasObject {
	return withBackground(container.NewPadded(obj), colorPanelHeader)
}

func withStatusBar(obj fyne.CanvasObject) fyne.CanvasObject {
	return withBackground(container.NewPadded(obj), colorPanelHeader)
}
