package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const popupShadowSize = float32(8)

var colorPopupShadow = color.NRGBA{R: 0, G: 0, B: 0, A: 120}

// modalDialog is a themed modal overlay on the main window canvas (stable on Windows).
type modalDialog struct {
	popup *widget.PopUp
}

func (d *modalDialog) Close() {
	if d != nil && d.popup != nil {
		d.popup.Hide()
	}
}

func (d *modalDialog) Canvas() fyne.Canvas {
	if d == nil || d.popup == nil {
		return nil
	}
	return d.popup.Canvas
}

func newModalDialog(parent fyne.Window, title string, size fyne.Size, body fyne.CanvasObject) *modalDialog {
	md := &modalDialog{}
	onClose := func() { md.Close() }
	card := buildDialogCard(title, body, onClose)
	md.popup = widget.NewModalPopUp(withPopupFrame(card), parent.Canvas())
	md.popup.Resize(size)
	md.popup.Show()
	return md
}

func buildDialogCard(title string, body fyne.CanvasObject, onClose func()) fyne.CanvasObject {
	titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), onClose)
	closeBtn.Importance = widget.LowImportance
	titleBar := withPanelHeader(container.NewBorder(nil, nil, titleLbl, closeBtn, nil))
	inner := container.NewBorder(titleBar, nil, nil, nil, container.NewPadded(body))
	return withBackground(inner, colorBG)
}

func withPopupFrame(content fyne.CanvasObject) fyne.CanvasObject {
	top := canvas.NewRectangle(colorBorder)
	top.SetMinSize(fyne.NewSize(0, 1))
	bottom := canvas.NewRectangle(colorBorder)
	bottom.SetMinSize(fyne.NewSize(0, 1))
	left := canvas.NewRectangle(colorBorder)
	left.SetMinSize(fyne.NewSize(1, 0))
	right := canvas.NewRectangle(colorBorder)
	right.SetMinSize(fyne.NewSize(1, 0))
	bordered := container.NewBorder(top, bottom, left, right, content)

	bottomShadow := canvas.NewRectangle(colorPopupShadow)
	bottomShadow.SetMinSize(fyne.NewSize(0, popupShadowSize))
	rightShadow := canvas.NewRectangle(colorPopupShadow)
	rightShadow.SetMinSize(fyne.NewSize(popupShadowSize, 0))

	return container.NewBorder(nil, bottomShadow, nil, rightShadow, bordered)
}

// showThemedWindow opens a modal dialog on the main window.
func showThemedWindow(parent fyne.Window, title string, size fyne.Size, body fyne.CanvasObject) *modalDialog {
	return newModalDialog(parent, title, size, body)
}
