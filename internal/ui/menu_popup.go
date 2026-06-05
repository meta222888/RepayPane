package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func newWideMenu(menu *fyne.Menu) (*widget.Menu, fyne.CanvasObject) {
	m := widget.NewMenu(menu)
	content := container.NewThemeOverride(m, newMenuWideTheme())
	return m, content
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
