package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// menuItemInnerPadding is the target horizontal inset per side; Fyne uses one
// InnerPadding for both axes, so we only widen via min width, not theme padding.
const menuItemInnerPadding float32 = 14

func newWideMenu(menu *fyne.Menu) (*widget.Menu, fyne.CanvasObject) {
	m := widget.NewMenu(menu)
	basePad := fyne.CurrentApp().Settings().Theme().Size(theme.SizeNameInnerPadding)
	extraW := (menuItemInnerPadding - basePad) * 2
	if extraW < 1 {
		return m, m
	}
	minW := m.MinSize().Width + extraW
	return m, minMenuWidth(minW, m)
}

func minMenuWidth(minW float32, obj fyne.CanvasObject) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(minW, 0))
	return container.NewStack(spacer, obj)
}

func showWidePopUpMenu(c fyne.Canvas, menu *fyne.Menu, at fyne.Position) *widget.PopUp {
	_, content := newWideMenu(menu)
	size := content.MinSize()
	pos := adjustMenuPosition(at, size, c)
	pop := widget.NewPopUp(content, c)
	pop.Resize(size)
	pop.ShowAtPosition(pos)
	return pop
}
