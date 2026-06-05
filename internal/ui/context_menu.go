package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func dismissPopUpMenus(c fyne.Canvas) {
	for _, o := range c.Overlays().List() {
		o.Hide()
	}
}

func showPopUpContextMenu(w fyne.Window, at fyne.Position, menu *fyne.Menu, onDismiss func()) {
	c := w.Canvas()
	dismissPopUpMenus(c)
	fyne.Do(func() {
		pop := widget.NewPopUpMenu(menu, c)
		pop.ShowAtPosition(at)
	})
	if onDismiss != nil {
		watchPopUpMenuDismiss(c, onDismiss)
	}
}

func watchPopUpMenuDismiss(c fyne.Canvas, after func()) {
	go func() {
		for i := 0; i < 120; i++ {
			time.Sleep(50 * time.Millisecond)
			done := false
			fyne.Do(func() {
				if len(c.Overlays().List()) == 0 {
					after()
					done = true
				}
			})
			if done {
				return
			}
		}
	}()
}
