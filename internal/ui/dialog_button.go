package ui

import (
	"fyne.io/fyne/v2/widget"
)

func newAccentButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButton(label, tapped)
	btn.Importance = widget.HighImportance
	return btn
}
