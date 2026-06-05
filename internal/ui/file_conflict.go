package ui

import (
	"strconv"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// resolveFileConflict asks the user to cancel, overwrite, or rename when dst already exists.
// onProceed receives the destination file name; empty string means cancelled.
func (a *App) resolveFileConflict(fileName string, exists func(name string) bool, onProceed func(name string)) {
	msg := widget.NewLabel(i18n.Tf(i18n.KeyFileExistsConflict, fileName))
	msg.Wrapping = fyne.TextWrapWord

	var dlg *modalDialog
	closeDlg := func() { dlg.Close() }

	cancelBtn := widget.NewButton(i18n.T(i18n.KeyCancel), func() {
		closeDlg()
		onProceed("")
	})
	overwriteBtn := widget.NewButton(i18n.T(i18n.KeyOverwrite), func() {
		closeDlg()
		onProceed(fileName)
	})
	renameBtn := widget.NewButton(i18n.T(i18n.KeyRename), func() {
		closeDlg()
		a.promptConflictRename(fileName, exists, onProceed)
	})
	overwriteBtn.Importance = widget.HighImportance
	buttons := container.NewHBox(cancelBtn, layout.NewSpacer(), renameBtn, overwriteBtn)
	body := container.NewBorder(nil, buttons, nil, nil, msg)
	dlg = newModalDialog(a, i18n.T(i18n.KeyFileExistsTitle), fyne.NewSize(420, 180), body)
}

func (a *App) promptConflictRename(original string, exists func(name string) bool, onProceed func(name string)) {
	entry := widget.NewEntry()
	entry.SetText(suggestCopyName(original, exists))

	var dlg *modalDialog
	cancelBtn := widget.NewButton(i18n.T(i18n.KeyCancel), func() {
		dlg.Close()
		onProceed("")
	})
	okBtn := widget.NewButton(i18n.T(i18n.KeyOK), func() {
		name := strings.TrimSpace(entry.Text)
		if name == "" {
			dlg.Close()
			onProceed("")
			return
		}
		if exists(name) {
			dlg.Close()
			a.resolveFileConflict(name, exists, onProceed)
			return
		}
		dlg.Close()
		onProceed(name)
	})
	okBtn.Importance = widget.HighImportance
	buttons := container.NewHBox(cancelBtn, layout.NewSpacer(), okBtn)
	form := container.NewVBox(widget.NewLabel(i18n.T(i18n.KeyRenamePrompt)), entry)
	body := container.NewBorder(nil, buttons, nil, nil, form)
	dlg = newModalDialog(a, i18n.T(i18n.KeyRename), fyne.NewSize(400, 150), body)
}

func suggestCopyName(original string, exists func(name string) bool) string {
	for i := 1; i < 1000; i++ {
		candidate := appendCopySuffix(original, i)
		if !exists(candidate) {
			return candidate
		}
	}
	return original + " copy"
}

func appendCopySuffix(name string, n int) string {
	dot := strings.LastIndex(name, ".")
	if dot > 0 {
		return name[:dot] + " (" + strconv.Itoa(n) + ")" + name[dot:]
	}
	return name + " (" + strconv.Itoa(n) + ")"
}
