package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	menuSidePadding float32 = 18
	menuVertPadding float32 = 8
)

func newPaddedMenu(menu *fyne.Menu) (*widget.Menu, fyne.CanvasObject) {
	m := widget.NewMenu(menu)
	padded := container.New(
		layout.NewCustomPaddedLayout(menuVertPadding, menuVertPadding, menuSidePadding, menuSidePadding),
		m,
	)
	return m, padded
}

func showPaddedPopUpMenu(c fyne.Canvas, menu *fyne.Menu, at fyne.Position) *widget.PopUp {
	_, content := newPaddedMenu(menu)
	size := content.MinSize()
	pos := adjustMenuPosition(at, size, c)
	pop := widget.NewPopUp(content, c)
	pop.Resize(size)
	pop.ShowAtPosition(pos)
	return pop
}
