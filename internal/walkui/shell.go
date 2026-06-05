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
	history := append([]string(nil), a.settings.ShellHistory...)

	setOutput := func(text string) {
		if outEdit != nil {
			outEdit.SetText(text)
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
		history = append([]string(nil), a.settings.ShellHistory...)
		if histBox != nil {
			histBox.SetModel(history)
		}
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
		Title:    i18n.T(i18n.KeyFeatShell),
		MinSize:  Size{760, 560},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{AssignTo: &cmdEdit, OnKeyDown: func(key walk.Key) {
						if key == walk.KeyReturn {
							runCmd()
						}
					}},
					PushButton{Text: i18n.T(i18n.KeyFeatShellRun), OnClicked: runCmd},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: i18n.T(i18n.KeyFeatShellHistory)},
					ComboBox{AssignTo: &histBox, Model: history, Editable: true},
					PushButton{Text: i18n.T(i18n.KeyFeatShellDelOne), OnClicked: func() {
						sel := histBox.Text()
						if sel != "" {
							a.removeShellHistory(sel)
							history = append([]string(nil), a.settings.ShellHistory...)
							histBox.SetModel(history)
						}
					}},
					PushButton{Text: i18n.T(i18n.KeyFeatShellDelAll), OnClicked: func() {
						a.settings.ShellHistory = nil
						_ = config.SaveSettings(a.settings)
						history = nil
						histBox.SetModel(history)
					}},
				},
			},
			Label{Text: i18n.T(i18n.KeyFeatShellCopyHint)},
			TextEdit{AssignTo: &outEdit, ReadOnly: true, VScroll: true, Font: Font{Family: "Consolas", PointSize: 9}},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}.Run(a.mw)
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

func (a *App) removeShellHistory(cmd string) {
	var out []string
	for _, c := range a.settings.ShellHistory {
		if c != cmd {
			out = append(out, c)
		}
	}
	a.settings.ShellHistory = out
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
