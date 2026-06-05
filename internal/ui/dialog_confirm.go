package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// dialogConfirmOn shows a confirmation overlay on parent without lowering the parent window.
func dialogConfirmOn(parent fyne.Window, title, msg string, callback func(bool)) {
	if parent == nil {
		return
	}
	c := parent.Canvas()

	msgLbl := widget.NewLabel(msg)
	msgLbl.Wrapping = fyne.TextWrapWord
	titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	var pop *widget.PopUp
	hide := func() {
		if pop != nil {
			pop.Hide()
		}
	}

	okBtn := newAccentButton(i18n.T(i18n.KeyOK), func() {
		hide()
		if callback != nil {
			callback(true)
		}
	})
	cancelBtn := widget.NewButton(i18n.T(i18n.KeyCancel), func() {
		hide()
		if callback != nil {
			callback(false)
		}
	})
	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, okBtn)
	card := withBackground(container.NewPadded(container.NewVBox(titleLbl, msgLbl, buttons)), colorPanel)

	pop = widget.NewModalPopUp(card, c)
	size := card.MinSize().Add(fyne.NewSize(32, 24))
	pop.Resize(fyne.NewSize(max(400, size.Width), max(180, size.Height)))
	pop.Show()
	raiseWindow(parent)
}
