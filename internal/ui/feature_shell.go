package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const maxShellHistory = 100

func (a *App) showRemoteShell() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	title := i18n.T(i18n.KeyFeatShell)
	history := append([]string(nil), a.settings.ShellHistory...)
	histIndex := len(history)

	entry := widget.NewEntry()
	entry.SetPlaceHolder(i18n.T(i18n.KeyFeatShellHint))

	setOutput, outputScroll := scrollSelectableText()

	histSelect := widget.NewSelect(history, func(s string) {
		if s != "" {
			entry.SetText(s)
		}
	})
	if len(history) == 0 {
		histSelect.PlaceHolder = i18n.T(i18n.KeyFeatShellNoHistory)
	}

	runCmd := func() {
		cmd := strings.TrimSpace(entry.Text)
		if cmd == "" {
			return
		}
		if remote.IsInteractiveCommand(cmd) {
			setOutput(formatShellOutput(cmd, "", i18n.T(i18n.KeyFeatShellInteractive)))
			return
		}
		a.pushShellHistory(cmd)
		history = append([]string(nil), a.settings.ShellHistory...)
		histSelect.Options = history
		histSelect.Refresh()
		histIndex = len(history)

		setOutput(i18n.T(i18n.KeyFeatRunning) + "\n$ " + cmd + "\n")
		go func() {
			out, err := client.RunCombined(cmd)
			fyne.Do(func() {
				setOutput(formatShellOutput(cmd, out, formatShellError(err)))
			})
		}()
	}

	entry.OnSubmitted = func(string) { runCmd() }

	delOne := newAccentButton(i18n.T(i18n.KeyFeatShellDelOne), func() {
		sel := histSelect.Selected
		if sel == "" {
			return
		}
		a.removeShellHistory(sel)
		history = append([]string(nil), a.settings.ShellHistory...)
		histSelect.Options = history
		histSelect.Selected = ""
		histSelect.Refresh()
	})
	delAll := newAccentButton(i18n.T(i18n.KeyFeatShellDelAll), func() {
		a.settings.ShellHistory = nil
		_ = config.SaveSettings(a.settings)
		history = nil
		histSelect.Options = history
		histSelect.Selected = ""
		histSelect.Refresh()
	})
	runBtn := newAccentButton(i18n.T(i18n.KeyFeatShellRun), runCmd)

	inputRow := container.NewBorder(nil, nil, nil, runBtn, entry)
	histRow := container.NewHBox(
		widget.NewLabel(i18n.T(i18n.KeyFeatShellHistory)),
		histSelect,
		delOne,
		delAll,
	)
	copyHint := widget.NewLabel(i18n.T(i18n.KeyFeatShellCopyHint))

	showThemedFeature(a, title, fyne.NewSize(760, 560), container.NewBorder(
		container.NewVBox(inputRow, histRow),
		copyHint,
		nil, nil,
		outputScroll,
	))

	navHistory := func(delta int) {
		if len(history) == 0 {
			return
		}
		histIndex += delta
		if histIndex < 0 {
			histIndex = 0
		}
		if histIndex >= len(history) {
			histIndex = len(history) - 1
		}
		entry.SetText(history[histIndex])
	}

	upKey := &desktop.CustomShortcut{KeyName: fyne.KeyUp, Modifier: 0}
	downKey := &desktop.CustomShortcut{KeyName: fyne.KeyDown, Modifier: 0}
	a.window.Canvas().AddShortcut(upKey, func(fyne.Shortcut) { navHistory(-1) })
	a.window.Canvas().AddShortcut(downKey, func(fyne.Shortcut) { navHistory(1) })
}

func formatShellOutput(cmd, out, errLine string) string {
	var b strings.Builder
	b.WriteString("$ ")
	b.WriteString(cmd)
	b.WriteString("\n\n")
	trimmed := strings.TrimSpace(out)
	if trimmed != "" {
		b.WriteString(trimmed)
	}
	if errLine != "" {
		if trimmed != "" {
			b.WriteString("\n\n")
		}
		b.WriteString(errLine)
	}
	return b.String()
}

func formatShellError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, remote.ErrCommandTimeout) {
		return "[" + i18n.T(i18n.KeyFeatShellTimeout) + "]"
	}
	if code, ok := remote.ExitStatus(err); ok {
		return fmt.Sprintf("[%s: %d]", i18n.T(i18n.KeyFeatShellExitCode), code)
	}
	return "[" + err.Error() + "]"
}

func (a *App) pushShellHistory(cmd string) {
	for i, h := range a.settings.ShellHistory {
		if h == cmd {
			a.settings.ShellHistory = append(a.settings.ShellHistory[:i], a.settings.ShellHistory[i+1:]...)
			break
		}
	}
	a.settings.ShellHistory = append(a.settings.ShellHistory, cmd)
	if len(a.settings.ShellHistory) > maxShellHistory {
		a.settings.ShellHistory = a.settings.ShellHistory[len(a.settings.ShellHistory)-maxShellHistory:]
	}
	_ = config.SaveSettings(a.settings)
}

func (a *App) removeShellHistory(cmd string) {
	out := a.settings.ShellHistory[:0]
	for _, h := range a.settings.ShellHistory {
		if h != cmd {
			out = append(out, h)
		}
	}
	a.settings.ShellHistory = out
	_ = config.SaveSettings(a.settings)
}

func (a *App) registerShellShortcut() {
	if _, ok := a.window.Canvas().(desktop.Canvas); !ok {
		return
	}
	ctrlE := &desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: desktop.ControlModifier}
	a.window.Canvas().AddShortcut(ctrlE, func(fyne.Shortcut) {
		a.showRemoteShell()
	})
}
