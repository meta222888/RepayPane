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
	title := i18n.T(i18n.KeyKeyPickerTitle)
	var dlg *modalDialog

	paths, _ := remote.ListSSHKeyFiles()
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder(`C:\Users\you\.ssh\id_rsa`)
	if current != "" {
		pathEntry.SetText(current)
	}

	selected := -1
	var list *widget.List
	list = widget.NewList(
		func() int { return len(paths) },
		func() fyne.CanvasObject { return newConnectPickerRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if int(id) >= len(paths) {
				return
			}
			row := obj.(*connectPickerRow)
			p := paths[id]
			idx := int(id)
			row.onPrimary = func() { selectPickerRow(list, &selected, idx) }
			row.update(idx, "🔑", filepath.Base(p), p, idx == selected)
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		selectPickerRow(list, &selected, int(id))
		if int(id) < len(paths) {
			pathEntry.SetText(paths[id])
		}
	}
	for i, p := range paths {
		if p == current {
			selected = i
			list.Select(widget.ListItemID(i))
			break
		}
	}

	hint := widget.NewLabel(i18n.T(i18n.KeyKeyPickerHint))
	okBtn := newAccentButton(i18n.T(i18n.KeyOK), func() {
		onPick(pathEntry.Text)
		dlg.Close()
	})
	cancelBtn := newAccentButton(i18n.T(i18n.KeyCancel), func() { dlg.Close() })

	buttons := container.NewHBox(cancelBtn, okBtn)
	body := container.NewBorder(
		container.NewVBox(hint, pathEntry),
		buttons,
		nil, nil,
		list,
	)
	dlg = newModalDialog(a.window, title, fyne.NewSize(620, 420), body)
}
