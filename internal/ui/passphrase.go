package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) promptPassphrase(onOK func(string), onCancel func()) {
	title := i18n.T(i18n.KeyPassphraseTitle)
	var dlg *modalDialog

	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder(i18n.T(i18n.KeyPassphrasePrompt))
	hint := widget.NewLabel(i18n.T(i18n.KeyPassphraseHint))

	submit := func() {
		dlg.Close()
		onOK(entry.Text)
	}
	entry.OnSubmitted = func(string) { submit() }

	okBtn := newAccentButton(i18n.T(i18n.KeyOK), submit)
	cancelBtn := newAccentButton(i18n.T(i18n.KeyCancel), func() {
		dlg.Close()
		onCancel()
	})

	buttons := container.NewHBox(cancelBtn, okBtn)
	body := container.NewBorder(hint, buttons, nil, nil, entry)
	dlg = newModalDialog(a, title, fyne.NewSize(420, 180), body)
}

func (a *App) askPassphraseBlocking() []byte {
	done := make(chan []byte, 1)
	fyne.Do(func() {
		a.promptPassphrase(
			func(p string) { done <- []byte(p) },
			func() { done <- nil },
		)
	})
	pass := <-done
	return pass
}
