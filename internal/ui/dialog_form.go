package ui

import (
	"image/color"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	nameFormDialogWidth float32 = 480
	nameFormEntryWidth  float32 = 320
)

// dialogNameFormOn shows a wide name-entry form overlay on parent.
func dialogNameFormOn(parent fyne.Window, title string, submit func(name string)) {
	if parent == nil {
		return
	}
	c := parent.Canvas()

	entry := widget.NewEntry()
	entrySpacer := canvas.NewRectangle(color.Transparent)
	entrySpacer.SetMinSize(fyne.NewSize(nameFormEntryWidth, 0))
	entryField := container.NewStack(entrySpacer, entry)

	var pop *widget.PopUp
	hide := func() {
		if pop != nil {
			pop.Hide()
		}
	}
	doOK := func() {
		name := strings.TrimSpace(entry.Text)
		hide()
		if name != "" && submit != nil {
			submit(name)
		}
	}
	entry.OnSubmitted = func(string) { doOK() }

	form := widget.NewForm(widget.NewFormItem(i18n.T(i18n.KeyFormName), entryField))
	titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	okBtn := newAccentButton(i18n.T(i18n.KeyOK), doOK)
	cancelBtn := widget.NewButton(i18n.T(i18n.KeyCancel), hide)
	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, okBtn)
	card := withBackground(container.NewPadded(container.NewVBox(titleLbl, form, buttons)), colorPanel)

	pop = widget.NewModalPopUp(card, c)
	size := card.MinSize().Add(fyne.NewSize(32, 24))
	pop.Resize(fyne.NewSize(max(nameFormDialogWidth, size.Width), max(160, size.Height)))
	pop.Show()
	raiseWindow(parent)
	c.Focus(entry)
}
