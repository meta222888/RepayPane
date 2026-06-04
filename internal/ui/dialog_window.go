package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const popupShadowSize = float32(8)

var colorPopupShadow = color.NRGBA{R: 0, G: 0, B: 0, A: 120}

func newThemedWindow(a fyne.App, size fyne.Size) fyne.Window {
	var w fyne.Window
	if drv, ok := a.Driver().(desktop.Driver); ok {
		w = drv.CreateSplashWindow()
	} else {
		w = a.NewWindow("")
	}
	// Extra space for drop shadow on right/bottom.
	w.Resize(fyne.NewSize(size.Width+popupShadowSize, size.Height+popupShadowSize))
	w.CenterOnScreen()
	w.SetPadded(false)
	return w
}

func themedWindowChrome(w fyne.Window, title string, body fyne.CanvasObject) fyne.CanvasObject {
	titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() { w.Close() })
	closeBtn.Importance = widget.LowImportance

	titleInner := container.NewBorder(nil, nil, titleLbl, closeBtn, nil)
	dragLayer := newDragRegion(w, layout.NewSpacer())
	titleBar := withPanelHeader(container.NewStack(dragLayer, titleInner))

	inner := container.NewBorder(titleBar, nil, nil, nil, container.NewPadded(body))
	panel := withBackground(inner, colorBG)
	frame := withPopupFrame(panel)

	outer := canvas.NewRectangle(colorBG)
	return container.NewStack(outer, frame)
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
	bordered := container.NewBorder(top, bottom, left, right, nil, content)

	bottomShadow := canvas.NewRectangle(colorPopupShadow)
	bottomShadow.SetMinSize(fyne.NewSize(0, popupShadowSize))
	rightShadow := canvas.NewRectangle(colorPopupShadow)
	rightShadow.SetMinSize(fyne.NewSize(popupShadowSize, 0))

	return container.NewBorder(nil, bottomShadow, nil, rightShadow, bordered)
}

func showThemedWindow(a fyne.App, title string, size fyne.Size, body fyne.CanvasObject) fyne.Window {
	w := newThemedWindow(a, size)
	w.SetContent(themedWindowChrome(w, title, body))
	w.Show()
	return w
}
