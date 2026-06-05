package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// modalDialog is a separate native window (movable and resizable on Windows).
type modalDialog struct {
	window  fyne.Window
	onClose func()
}

func (d *modalDialog) Close() {
	if d == nil || d.window == nil {
		return
	}
	d.window.Close()
}

func (d *modalDialog) Window() fyne.Window {
	if d == nil {
		return nil
	}
	return d.window
}

func (d *modalDialog) Canvas() fyne.Canvas {
	if d == nil || d.window == nil {
		return nil
	}
	return d.window.Canvas()
}

func (d *modalDialog) SetOnClose(fn func()) {
	if d != nil {
		d.onClose = fn
	}
}

func newModalDialog(a *App, title string, size fyne.Size, body fyne.CanvasObject) *modalDialog {
	w := a.fyneApp.NewWindow(title)
	w.Resize(size)
	w.SetPadded(true)
	w.SetContent(container.NewPadded(withBackground(body, colorBG)))
	w.CenterOnScreen()

	md := &modalDialog{window: w}
	w.SetCloseIntercept(func() {
		if md.onClose != nil {
			md.onClose()
		}
		w.SetCloseIntercept(nil)
		w.Close()
	})
	w.Show()
	setWindowOwner(w, a.window)
	raiseWindow(w)
	return md
}

func newModalDialogAuto(a *App, title string, minWidth float32, body fyne.CanvasObject) *modalDialog {
	w := a.fyneApp.NewWindow(title)
	w.SetPadded(true)
	content := container.NewPadded(withBackground(body, colorBG))
	w.SetContent(content)
	min := content.MinSize()
	w.Resize(fyne.NewSize(max(minWidth, min.Width), min.Height))
	w.CenterOnScreen()

	md := &modalDialog{window: w}
	w.SetCloseIntercept(func() {
		if md.onClose != nil {
			md.onClose()
		}
		w.SetCloseIntercept(nil)
		w.Close()
	})
	w.Show()
	setWindowOwner(w, a.window)
	raiseWindow(w)
	return md
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func showThemedWindow(a *App, title string, size fyne.Size, body fyne.CanvasObject) *modalDialog {
	return newModalDialog(a, title, size, body)
}
