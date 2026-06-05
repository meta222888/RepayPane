package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"
	"github.com/relaypane/relaypane/internal/textencoding"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type EditorWindow struct {
	app       *App
	path      string
	saveFn    func([]byte) error
	loadRawFn func() ([]byte, error)
	encInfo   textencoding.Info
	content   *widget.Entry
	statusLbl *widget.Label
	hintBase  string
	dirty     bool
	saving    bool
	window    fyne.Window
}

func ShowEditor(app *App, entry remote.FileInfo, text string, enc textencoding.Info) {
	path := entry.Path
	showTextEditor(app, i18n.Tf(i18n.KeyEditTitle, entry.Name), path, text, enc, i18n.T(i18n.KeyCtrlSSave),
		func(data []byte) error {
			client := app.activeClient()
			if client == nil {
				return fmt.Errorf(i18n.T(i18n.KeyNotConnectedErr))
			}
			return client.WriteFile(path, data)
		},
		func() ([]byte, error) {
			client := app.activeClient()
			if client == nil {
				return nil, fmt.Errorf(i18n.T(i18n.KeyNotConnectedErr))
			}
			return client.ReadFile(path)
		},
	)
}

func ShowLocalEditor(app *App, path, name, text string, enc textencoding.Info) {
	showTextEditor(app, i18n.Tf(i18n.KeyEditTitle, name), path, text, enc, i18n.T(i18n.KeyCtrlSSaveLocal),
		func(data []byte) error {
			return os.WriteFile(path, data, 0o644)
		},
		func() ([]byte, error) {
			return os.ReadFile(path)
		},
	)
}

func showTextEditor(app *App, title, path, text string, enc textencoding.Info, hint string, saveFn func([]byte) error, loadRawFn func() ([]byte, error)) {
	e := &EditorWindow{
		app:       app,
		path:      path,
		saveFn:    saveFn,
		loadRawFn: loadRawFn,
		encInfo:   enc,
		hintBase:  hint,
	}
	e.content = widget.NewMultiLineEntry()
	e.content.SetText(text)
	e.content.Wrapping = fyne.TextWrapWord
	e.content.OnChanged = func(string) { e.dirty = true }

	e.statusLbl = widget.NewLabel(e.statusHint())

	revertBtn := widget.NewButton(i18n.T(i18n.KeyEditorRevert), func() { e.revert() })
	saveBtn := newAccentButton(i18n.T(i18n.KeySave), func() { e.save() })
	bottomRight := container.NewHBox(revertBtn, saveBtn)
	bottom := container.NewBorder(nil, nil, e.statusLbl, bottomRight, nil)

	e.window = app.fyneApp.NewWindow(title)
	e.window.Resize(fyne.NewSize(900, 600))
	// Entry wraps internally; an outer Scroll uses unwrapped MinSize and breaks word wrap until focus.
	e.window.SetContent(container.NewBorder(
		widget.NewLabel(path),
		bottom,
		nil, nil,
		e.content,
	))

	e.registerSaveShortcut()

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

func (e *EditorWindow) statusHint() string {
	return fmt.Sprintf("%s · %s", e.hintBase, i18n.Tf(i18n.KeyEditorEncoding, e.encInfo.Label()))
}

func (e *EditorWindow) registerSaveShortcut() {
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	save := func() { e.save() }
	e.window.Canvas().AddShortcut(ctrlS, func(fyne.Shortcut) { save() })
	saveItem := fyne.NewMenuItem(i18n.T(i18n.KeySave), save)
	saveItem.Shortcut = ctrlS
	e.window.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File", saveItem)))
}

func (e *EditorWindow) setStatus(text string) {
	e.statusLbl.SetText(text)
}

func (e *EditorWindow) save() {
	if e.saving {
		return
	}
	e.saving = true
	e.setStatus(i18n.T(i18n.KeyEditorSaving))
	go func() {
		data, encErr := textencoding.Encode(e.content.Text, e.encInfo)
		if encErr != nil {
			fyne.Do(func() {
				e.saving = false
				e.setStatus(i18n.Tf(i18n.KeySaveFailedAt, encErr.Error(), time.Now().Format("2006-01-02 15:04:05")))
			})
			return
		}
		err := e.saveFn(data)
		fyne.Do(func() {
			e.saving = false
			ts := time.Now().Format("2006-01-02 15:04:05")
			if err != nil {
				e.setStatus(i18n.Tf(i18n.KeySaveFailedAt, err.Error(), ts))
				return
			}
			e.dirty = false
			e.setStatus(i18n.Tf(i18n.KeySaveSuccessAt, ts))
		})
	}()
}

func (e *EditorWindow) revert() {
	doReload := func() {
		e.setStatus(i18n.T(i18n.KeyEditorSaving))
		go func() {
			raw, err := e.loadRawFn()
			if err != nil {
				fyne.Do(func() {
					e.setStatus(i18n.Tf(i18n.KeyEditorRevertFailed, err.Error()))
				})
				return
			}
			text, enc, err := textencoding.Decode(raw)
			fyne.Do(func() {
				if err != nil {
					e.setStatus(i18n.Tf(i18n.KeyEditorRevertFailed, err.Error()))
					return
				}
				e.encInfo = enc
				e.content.SetText(text)
				e.dirty = false
				e.setStatus(e.statusHint())
			})
		}()
	}
	if !e.dirty {
		doReload()
		return
	}
	dialog.ShowConfirm(i18n.T(i18n.KeyEditorRevert), i18n.T(i18n.KeyEditorRevertConfirm), func(ok bool) {
		if ok {
			doReload()
		}
	}, e.window)
}
