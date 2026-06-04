package ui

import (
	"path/filepath"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func showSSHKeyPicker(a *App, current string, onPick func(path string)) {
	w := a.fyneApp.NewWindow(i18n.T(i18n.KeyKeyPickerTitle))
	w.Resize(fyne.NewSize(620, 420))
	w.CenterOnScreen()

	paths, _ := remote.ListSSHKeyFiles()
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder(`C:\Users\you\.ssh\id_rsa`)
	if current != "" {
		pathEntry.SetText(current)
	}

	list := widget.NewList(
		func() int { return len(paths) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if int(id) >= len(paths) {
				return
			}
			obj.(*widget.Label).SetText(filepath.Base(paths[id]) + "    " + paths[id])
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		if int(id) < len(paths) {
			pathEntry.SetText(paths[id])
		}
	}
	for i, p := range paths {
		if p == current {
			list.Select(widget.ListItemID(i))
			break
		}
	}

	hint := widget.NewLabel(i18n.T(i18n.KeyKeyPickerHint))
	okBtn := widget.NewButton(i18n.T(i18n.KeyOK), func() {
		onPick(pathEntry.Text)
		w.Close()
	})
	okBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButton(i18n.T(i18n.KeyCancel), func() { w.Close() })

	buttons := container.NewHBox(cancelBtn, okBtn)
	content := container.NewBorder(
		container.NewVBox(hint, pathEntry),
		buttons,
		nil, nil,
		list,
	)
	w.SetContent(container.NewPadded(content))
	w.Show()
}
