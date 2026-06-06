package walkui

import (
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const maxShellHistory = 100

func (a *App) showRemoteShell() {
	client, ok := a.requireClient()
	if !ok {
		return
	}

	var dlg *walk.Dialog
	var cmdEdit *walk.LineEdit
	var outEdit *walk.TextEdit
	var histBox *walk.ComboBox
	history := mergedShellHistory(a.settings)

	title := i18n.T(i18n.KeyFeatShell)
	if tab := a.activeSession(); tab != nil {
		title += " — " + serverDisplayName(tab.server)
	}

	setOutput := func(text string) {
		setMultilineText(outEdit, text)
	}

	applyHistorySelection := func() {
		if histBox == nil || cmdEdit == nil {
			return
		}
		idx := histBox.CurrentIndex()
		if idx < 0 || idx >= len(history) {
			return
		}
		cmd := history[idx]
		cmdEdit.SetText(cmd)
		cmdEdit.SetFocus()
		cmdEdit.SetTextSelection(0, len(cmd))
	}

	refreshHistory := func() {
		history = mergedShellHistory(a.settings)
		if histBox != nil {
			histBox.SetModel(history)
		}
	}

	runCmd := func() {
		cmd := strings.TrimSpace(cmdEdit.Text())
		if cmd == "" {
			return
		}
		if remote.IsInteractiveCommand(cmd) {
			setOutput(formatShellOutput(cmd, "", i18n.T(i18n.KeyFeatShellInteractive)))
			return
		}
		a.pushShellHistory(cmd)
		refreshHistory()
		setOutput(i18n.T(i18n.KeyFeatRunning) + "\n$ " + cmd + "\n")
		go func() {
			out, err := client.RunCombined(cmd)
			a.syncUI(func() {
				setOutput(formatShellOutput(cmd, out, formatShellError(err)))
			})
		}()
	}

	_, _ = Dialog{
		AssignTo: &dlg,
		Title:    title,
		MinSize:  Size{760, 560},
		Font:     uiFont(),
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			dlgBody(
				Composite{
					Layout: HBox{Spacing: 6},
					Children: []Widget{
						LineEdit{
							AssignTo: &cmdEdit,
							Font:     monoFont(),
							OnKeyDown: func(key walk.Key) {
								if key == walk.KeyReturn {
									runCmd()
								}
							},
						},
						PushButton{Text: i18n.T(i18n.KeyFeatShellRun), OnClicked: runCmd},
					},
				},
				Composite{
					Layout: HBox{Spacing: 6},
					Children: []Widget{
						Label{Text: i18n.T(i18n.KeyFeatShellHistory)},
						ComboBox{
							AssignTo:              &histBox,
							Model:                 history,
							Editable:              true,
							OnCurrentIndexChanged: func() { applyHistorySelection() },
						},
						PushButton{Text: i18n.T(i18n.KeyFeatShellPin), OnClicked: func() {
							cmd := strings.TrimSpace(cmdEdit.Text())
							if cmd == "" && histBox != nil {
								cmd = strings.TrimSpace(histBox.Text())
							}
							if cmd != "" {
								a.pinShellHistory(cmd)
								refreshHistory()
							}
						}},
						PushButton{Text: i18n.T(i18n.KeyFeatShellDelOne), OnClicked: func() {
							sel := histBox.Text()
							if sel != "" {
								a.removeShellHistory(sel)
								refreshHistory()
							}
						}},
						PushButton{Text: i18n.T(i18n.KeyFeatShellDelAll), OnClicked: func() {
							a.clearShellHistory()
							refreshHistory()
						}},
					},
				},
				Label{Text: i18n.T(i18n.KeyFeatShellCopyHint), TextColor: colorTextMuted},
				TextEdit{AssignTo: &outEdit, ReadOnly: true, VScroll: true, Font: monoFont()},
			),
			dlgFooter(
				PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
			),
		},
	}.Run(a.mw)
}

func mergedShellHistory(s *config.Settings) []string {
	seen := make(map[string]bool)
	var out []string
	for _, c := range s.ShellPinned {
		c = strings.TrimSpace(c)
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	for _, c := range s.ShellHistory {
		c = strings.TrimSpace(c)
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	return out
}

func (a *App) pushShellHistory(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	h := a.settings.ShellHistory
	for i, c := range h {
		if c == cmd {
			h = append(h[:i], h[i+1:]...)
			break
		}
	}
	h = append(h, cmd)
	if len(h) > maxShellHistory {
		h = h[len(h)-maxShellHistory:]
	}
	a.settings.ShellHistory = h
	_ = config.SaveSettings(a.settings)
}

func (a *App) pinShellHistory(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	for _, c := range a.settings.ShellPinned {
		if c == cmd {
			return
		}
	}
	a.settings.ShellPinned = append(a.settings.ShellPinned, cmd)
	a.removeShellHistoryOnly(cmd)
	_ = config.SaveSettings(a.settings)
}

func (a *App) removeShellHistoryOnly(cmd string) {
	var out []string
	for _, c := range a.settings.ShellHistory {
		if c != cmd {
			out = append(out, c)
		}
	}
	a.settings.ShellHistory = out
}

func (a *App) removeShellHistory(cmd string) {
	pinned := false
	for _, c := range a.settings.ShellPinned {
		if c == cmd {
			pinned = true
			break
		}
	}
	if pinned {
		var out []string
		for _, c := range a.settings.ShellPinned {
			if c != cmd {
				out = append(out, c)
			}
		}
		a.settings.ShellPinned = out
	} else {
		a.removeShellHistoryOnly(cmd)
	}
	_ = config.SaveSettings(a.settings)
}

func (a *App) clearShellHistory() {
	a.settings.ShellHistory = nil
	_ = config.SaveSettings(a.settings)
}

func formatShellOutput(cmd, out, errMsg string) string {
	var b strings.Builder
	b.WriteString("$ ")
	b.WriteString(cmd)
	b.WriteString("\n\n")
	if strings.TrimSpace(out) != "" {
		b.WriteString(out)
		if !strings.HasSuffix(out, "\n") {
			b.WriteString("\n")
		}
	}
	if errMsg != "" {
		b.WriteString("\n")
		b.WriteString(errMsg)
	}
	return b.String()
}

func formatShellError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
