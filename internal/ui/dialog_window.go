package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newThemedWindow(a fyne.App, size fyne.Size) fyne.Window {
	var w fyne.Window
	if drv, ok := a.Driver().(desktop.Driver); ok {
		w = drv.CreateSplashWindow()
	} else {
		w = a.NewWindow("")
	}
	w.Resize(size)
	w.CenterOnScreen()
	w.SetPadded(false)
	return w
}

func themedWindowChrome(w fyne.Window, title string, body fyne.CanvasObject) fyne.CanvasObject {
	titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() { w.Close() })
	closeBtn.Importance = widget.LowImportance
	titleBar := withPanelHeader(container.NewBorder(nil, nil, titleLbl, closeBtn))
	inner := container.NewBorder(titleBar, nil, nil, nil, container.NewPadded(body))
	bg := canvas.NewRectangle(colorBG)
	return container.NewStack(bg, inner)
}

func showThemedWindow(a fyne.App, title string, size fyne.Size, body fyne.CanvasObject) fyne.Window {
	w := newThemedWindow(a, size)
	w.SetContent(themedWindowChrome(w, title, body))
	w.Show()
	return w
}
