package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type ctxMenuState struct {
	popup     *widget.PopUp
	onDismiss func()
}

var activeCtxMenu *ctxMenuState

func hideActiveContextMenu() {
	if activeCtxMenu == nil {
		return
	}
	st := activeCtxMenu
	activeCtxMenu = nil
	st.popup.Hide()
	if st.onDismiss != nil {
		st.onDismiss()
	}
}

func dismissPopUpMenus(c fyne.Canvas) {
	hideActiveContextMenu()
	for _, o := range c.Overlays().List() {
		o.Hide()
	}
}

func adjustMenuPosition(at fyne.Position, menuSize fyne.Size, c fyne.Canvas) fyne.Position {
	_, areaSize := c.InteractiveArea()
	x, y := at.X, at.Y
	if x+menuSize.Width > areaSize.Width {
		x = areaSize.Width - menuSize.Width
		if x < 0 {
			x = 0
		}
	}
	if y+menuSize.Height > areaSize.Height {
		y = areaSize.Height - menuSize.Height
		if y < 0 {
			y = 0
		}
	}
	return fyne.NewPos(x, y)
}

// showPopUpContextMenu shows a menu-sized overlay only (not full-window).
// Main UI stays visible; right-click elsewhere reaches the file list and replaces the menu.
func showPopUpContextMenu(w fyne.Window, at fyne.Position, menu *fyne.Menu, onDismiss func()) {
	c := w.Canvas()
	dismissPopUpMenus(c)

	menuWidget, menuContent := newWideMenu(menu, nil)
	menuWidget.OnDismiss = func() { hideActiveContextMenu() }
	menuSize := menuContent.MinSize()
	menuPos := adjustMenuPosition(at, menuSize, c)

	pop := widget.NewPopUp(menuContent, c)
	pop.Resize(menuSize)

	activeCtxMenu = &ctxMenuState{popup: pop, onDismiss: onDismiss}
	pop.ShowAtPosition(menuPos)
}
