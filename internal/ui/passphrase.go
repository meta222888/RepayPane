package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) promptPassphrase(onOK func(string), onCancel func()) {
	w := a.fyneApp.NewWindow(i18n.T(i18n.KeyPassphraseTitle))
	w.Resize(fyne.NewSize(420, 160))
	w.CenterOnScreen()

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
	content := container.NewBorder(hint, buttons, nil, nil, entry)
	w.SetContent(container.NewPadded(content))
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
