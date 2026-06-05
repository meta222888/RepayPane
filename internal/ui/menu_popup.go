package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// menuItemHorizontalPad is the target horizontal inset per side. Row spacing uses
// menuItemInnerPadding via theme; extra width is added without outer margin wrappers.
const menuItemHorizontalPad float32 = 14

func newWideMenu(menu *fyne.Menu, overlay fyne.Theme) (*widget.Menu, fyne.CanvasObject) {
	m := widget.NewMenu(menu)
	content := fyne.CanvasObject(m)
	if overlay != nil {
		content = container.NewThemeOverride(m, overlay)
	}
	extraW := (menuItemHorizontalPad - menuItemInnerPadding) * 2
	if extraW < 1 {
		return m, content
	}
	minW := m.MinSize().Width + extraW
	return m, minMenuWidth(minW, content)
}

func minMenuWidth(minW float32, obj fyne.CanvasObject) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(minW, 0))
	return container.NewStack(spacer, obj)
}

func showWidePopUpMenu(c fyne.Canvas, menu *fyne.Menu, at fyne.Position) *widget.PopUp {
	_, content := newWideMenu(menu, newMenuPopupTheme())
	size := content.MinSize()
	pos := adjustMenuPosition(at, size, c)
	pop := widget.NewPopUp(content, c)
	pop.Resize(size)
	pop.ShowAtPosition(pos)
	return pop
}
