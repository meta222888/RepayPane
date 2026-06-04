package ui

import (
	"fmt"

	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type EditorWindow struct {
	app     *App
	entry   remote.FileInfo
	content *widget.Entry
	dirty   bool
	window  fyne.Window
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

	title := fmt.Sprintf("Edit: %s", entry.Name)
	e.window = app.fyneApp.NewWindow(title)
	e.window.Resize(fyne.NewSize(900, 600))
	e.window.SetContent(container.NewBorder(
		widget.NewLabel(entry.Path),
		widget.NewLabel("Ctrl+S to save to server"),
		nil, nil,
		container.NewScroll(e.content),
	))

	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	e.window.Canvas().AddShortcut(ctrlS, func(fyne.Shortcut) { e.save() })

	e.window.SetCloseIntercept(func() {
		if e.dirty {
			dialog.ShowConfirm("Unsaved changes", "Discard changes?", func(ok bool) {
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
		dialog.ShowError(fmt.Errorf("not connected"), e.window)
		return
	}
	data := []byte(e.content.Text)
	go func() {
		err := e.app.client.WriteFile(e.entry.Path, data)
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(fmt.Errorf("save failed: %w", err), e.window)
				return
			}
			e.dirty = false
			dialog.ShowInformation("Saved", fmt.Sprintf("Saved to %s", e.entry.Path), e.window)
		})
	}()
}
