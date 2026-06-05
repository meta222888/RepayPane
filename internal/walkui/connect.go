package walkui

import (
	"fmt"

	"github.com/relaypane/relaypane/internal/config"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (a *App) askPassphrase() []byte {
	var dlg *walk.Dialog
	var edit *walk.LineEdit
	var acceptBtn *walk.PushButton
	accepted := false

	if _, err := (Dialog{
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
					PushButton{
						Text: "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}).Run(a.mw); err != nil {
		return nil
	}
	if !accepted {
		return nil
	}
	return []byte(edit.Text())
}

func (a *App) showConnectDialog() {
	if len(a.store.Servers) == 0 {
		walk.MsgBox(a.mw, "RelayPane", "No saved servers. Add one in the Fyne edition or edit ~/.relaypane/servers.json.", walk.MsgBoxIconInformation)
		return
	}

	var dlg *walk.Dialog
	var lb *walk.ListBox
	selected := -1

	names := make([]string, len(a.store.Servers))
	for i, s := range a.store.Servers {
		names[i] = serverDisplayName(s) + "  (" + serverSubtitle(s) + ")"
	}

	if _, err := (Dialog{
		AssignTo: &dlg,
		Title:    "Connect to Server",
		MinSize:  Size{480, 320},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "Select a server and click Connect:"},
			ListBox{
				AssignTo: &lb,
				Model:    names,
				OnCurrentIndexChanged: func() {
					selected = lb.CurrentIndex()
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Connect",
						OnClicked: func() {
							if selected < 0 || selected >= len(a.store.Servers) {
								walk.MsgBox(dlg, "Connect", "Please select a server.", walk.MsgBoxIconWarning)
								return
							}
							s := a.store.Servers[selected]
							dlg.Accept()
							a.connectServer(s)
						},
					},
					PushButton{
						Text: "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}).Run(a.mw); err != nil {
		return
	}
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
