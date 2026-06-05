package ui

import (
	"fmt"
	"os"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type EditorWindow struct {
	app     *App
	path    string
	saveFn  func([]byte) error
	content *widget.Entry
	dirty   bool
	window  fyne.Window
}

func ShowEditor(app *App, entry remote.FileInfo, text string) {
	showTextEditor(app, i18n.Tf(i18n.KeyEditTitle, entry.Name), entry.Path, text, i18n.T(i18n.KeyCtrlSSave), func(data []byte) error {
		client := app.activeClient()
		if client == nil {
			return fmt.Errorf(i18n.T(i18n.KeyNotConnectedErr))
		}
		return client.WriteFile(entry.Path, data)
	})
}

func ShowLocalEditor(app *App, path, name, text string) {
	showTextEditor(app, i18n.Tf(i18n.KeyEditTitle, name), path, text, i18n.T(i18n.KeyCtrlSSaveLocal), func(data []byte) error {
		return os.WriteFile(path, data, 0o644)
	})
}

func showTextEditor(app *App, title, path, text, saveHint string, saveFn func([]byte) error) {
	e := &EditorWindow{app: app, path: path, saveFn: saveFn}
	e.content = widget.NewMultiLineEntry()
	e.content.SetText(text)
	e.content.Wrapping = fyne.TextWrapWord
	e.content.OnChanged = func(string) { e.dirty = true }

	e.window = app.fyneApp.NewWindow(title)
	e.window.Resize(fyne.NewSize(900, 600))
	e.window.SetContent(container.NewBorder(
		widget.NewLabel(path),
		widget.NewLabel(saveHint),
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
	data := []byte(e.content.Text)
	go func() {
		err := e.saveFn(data)
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(fmt.Errorf(i18n.Tf(i18n.KeySaveFailed, err.Error())), e.window)
				return
			}
			e.dirty = false
			dialog.ShowInformation(i18n.T(i18n.KeySaved), i18n.Tf(i18n.KeySavedMsg, e.path), e.window)
		})
	}()
}
