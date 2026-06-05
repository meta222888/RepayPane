package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func dialogShow(a *App, title, msg string) {
	w := a.fyneApp.NewWindow(title)
	w.Resize(fyne.NewSize(420, 180))
	lbl := widget.NewLabel(msg)
	lbl.Wrapping = fyne.TextWrapWord
	ok := widget.NewButton(i18n.T(i18n.KeyOK), func() { w.Close() })
	body := container.NewBorder(nil, ok, nil, nil, lbl)
	w.SetContent(container.NewPadded(withBackground(body, colorBG)))
	w.CenterOnScreen()
	w.Show()
}

func dialogShowError(a *App, err error) {
	if err == nil {
		return
	}
	w := a.fyneApp.NewWindow(i18n.T(i18n.KeyConnectionFailed))
	w.Resize(fyne.NewSize(480, 200))
	lbl := widget.NewLabel(err.Error())
	lbl.Wrapping = fyne.TextWrapWord
	ok := widget.NewButton(i18n.T(i18n.KeyOK), func() { w.Close() })
	body := container.NewBorder(nil, ok, nil, nil, lbl)
	w.SetContent(container.NewPadded(withBackground(body, colorBG)))
	w.CenterOnScreen()
	w.Show()
}

// dialogShowOn shows an informational dialog as an overlay on parent (stays above that window).
func dialogShowOn(parent fyne.Window, title, msg string) {
	if parent == nil {
		return
	}
	dialog.ShowInformation(title, msg, parent)
	raiseWindow(parent)
}

// dialogShowErrorOn shows an error dialog as an overlay on parent (stays above that window).
func dialogShowErrorOn(parent fyne.Window, err error) {
	if err == nil || parent == nil {
		return
	}
	dialog.ShowError(err, parent)
	raiseWindow(parent)
}
