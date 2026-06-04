package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) promptPassphrase(onOK func(string), onCancel func()) {
	title := i18n.T(i18n.KeyPassphraseTitle)
	w := newThemedWindow(a.fyneApp, fyne.NewSize(420, 180))

	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder(i18n.T(i18n.KeyPassphrasePrompt))
	hint := widget.NewLabel(i18n.T(i18n.KeyPassphraseHint))

	submit := func() {
		w.Close()
		onOK(entry.Text)
	}
	entry.OnSubmitted = func(string) { submit() }

	okBtn := widget.NewButton(i18n.T(i18n.KeyOK), submit)
	okBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButton(i18n.T(i18n.KeyCancel), func() {
		w.Close()
		onCancel()
	})

	buttons := container.NewHBox(cancelBtn, okBtn)
	body := container.NewBorder(hint, buttons, nil, nil, entry)
	w.SetContent(themedWindowChrome(w, title, body))
	w.Show()
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
