package walkui

import (
	"fmt"
	"os"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func initLanguage(settings *config.Settings) {
	if settings.Language == "zh" {
		i18n.SetLanguage(i18n.ZH)
		return
	}
	if settings.Language == "en" {
		i18n.SetLanguage(i18n.EN)
		return
	}
	if lang := os.Getenv("LANG"); strings.HasPrefix(strings.ToLower(lang), "en") {
		i18n.SetLanguage(i18n.EN)
		settings.Language = "en"
		return
	}
	i18n.SetLanguage(i18n.ZH)
	settings.Language = "zh"
}

func (a *App) showConnectDialog() {
	if len(a.store.Servers) == 0 {
		a.showAddServer()
		return
	}

	var dlg *walk.Dialog
	var lb *walk.ListBox
	selected := -1

	names := make([]string, len(a.store.Servers))
	for i, s := range a.store.Servers {
		names[i] = serverDisplayName(s) + "  (" + serverSubtitle(s) + ")"
	}

	_, _ = Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyConnectPickerTitle),
		MinSize:  Size{480, 320},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: i18n.T(i18n.KeyConnectPickerHint)},
			ListBox{
				AssignTo: &lb,
				Model:    names,
				OnCurrentIndexChanged: func() {
					selected = lb.CurrentIndex()
				},
				OnItemActivated: func() {
					if selected >= 0 && selected < len(a.store.Servers) {
						dlg.Cancel()
						a.openServerTab(a.store.Servers[selected])
					}
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: i18n.T(i18n.KeyCancel), OnClicked: func() { dlg.Cancel() }},
					PushButton{Text: i18n.T(i18n.KeyNewConnection), OnClicked: func() {
						dlg.Cancel()
						a.showAddServer()
					}},
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyConnect), OnClicked: func() {
						if selected < 0 {
							a.showMsg(i18n.T(i18n.KeyConnectPickerTitle), i18n.T(i18n.KeySelectServer))
							return
						}
						dlg.Cancel()
						a.openServerTab(a.store.Servers[selected])
					}},
				},
			},
		},
	}.Run(a.mw)
}

func (a *App) askPassphrase() []byte {
	var dlg *walk.Dialog
	var edit *walk.LineEdit
	var acceptBtn *walk.PushButton
	accepted := false

	_, _ = Dialog{
		AssignTo:      &dlg,
		Title:         "SSH Passphrase",
		DefaultButton: &acceptBtn,
		MinSize:       Size{360, 120},
		Layout:        VBox{},
		Children: []Widget{
			Label{Text: "Enter passphrase for private key:"},
			LineEdit{AssignTo: &edit, PasswordMode: true},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptBtn,
						Text:     "OK",
						OnClicked: func() {
							accepted = true
							dlg.Accept()
						},
					},
					PushButton{Text: "Cancel", OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}.Run(a.mw)
	if !accepted {
		return nil
	}
	return []byte(edit.Text())
}

func serverSubtitle(s config.Server) string {
	port := s.Port
	if port == 0 {
		port = 22
	}
	if s.Username == "" {
		return fmt.Sprintf("%s:%d", s.Host, port)
	}
	return fmt.Sprintf("%s@%s:%d", s.Username, s.Host, port)
}
