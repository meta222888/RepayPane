package ui

import (
	"fmt"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type EditorWindow struct {
	app       *App
	entry     remote.FileInfo
	content   *widget.Entry
	dirty     bool
	window    fyne.Window
	pathLabel *widget.Label
	hintLabel *widget.Label
}

func ShowEditor(app *App, entry remote.FileInfo, text string) {
	e := &EditorWindow{
		app:   app,
		entry: entry,
		dirty: false,
	}
	e.content = widget.NewMultiLineEntry()
	e.content.SetText(text)
	e.content.Wrapping = fyne.TextWrapWord
	e.content.OnChanged = func(string) { e.dirty = true }

	e.pathLabel = widget.NewLabel(entry.Path)
	e.hintLabel = widget.NewLabel(i18n.T(i18n.KeyCtrlSSave))

	e.window = app.fyneApp.NewWindow(i18n.Tf(i18n.KeyEditTitle, entry.Name))
	e.window.Resize(fyne.NewSize(900, 600))
	e.window.SetContent(container.NewBorder(
		e.pathLabel,
		e.hintLabel,
		nil, nil,
		container.NewScroll(e.content),
	))

	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	e.window.Canvas().AddShortcut(ctrlS, func(fyne.Shortcut) { e.save() })

	e.window.SetCloseIntercept(func() {
		if e.dirty {
			dialog.ShowConfirm(i18n.T(i18n.KeyUnsaved), i18n.T(i18n.KeyDiscard), func(ok bool) {
				if ok {
					e.window.Close()
				}
			}, e.window)
			return
		}
		e.window.Close()
	})

	e.window.Show()
}

func (e *EditorWindow) save() {
	if e.app.client == nil {
		dialog.ShowError(fmt.Errorf(i18n.T(i18n.KeyNotConnectedErr)), e.window)
		return
	}
	data := []byte(e.content.Text)
	go func() {
		err := e.app.client.WriteFile(e.entry.Path, data)
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(fmt.Errorf(i18n.Tf(i18n.KeySaveFailed, err.Error())), e.window)
				return
			}
			e.dirty = false
			dialog.ShowInformation(i18n.T(i18n.KeySaved), i18n.Tf(i18n.KeySavedMsg, e.entry.Path), e.window)
		})
	}()
}
