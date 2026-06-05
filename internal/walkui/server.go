package walkui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (a *App) showAddServer() {
	a.showServerForm(config.Server{
		Port:         config.DefaultSFTPPort,
		HeartbeatSec: config.DefaultHeartbeatSec,
		AutoSSHKey:   true,
	}, false, -1)
}

func (a *App) showEditServer(idx int) {
	if idx < 0 || idx >= len(a.store.Servers) {
		return
	}
	s := a.store.Servers[idx]
	a.showServerForm(s, true, idx)
}

func (a *App) showServerForm(initial config.Server, editMode bool, editIdx int) {
	var dlg *walk.Dialog
	var nameEdit, hostEdit, portEdit, userEdit, passEdit, rootEdit, hbEdit *walk.LineEdit
	var autoKeyCheck *walk.CheckBox
	autoSSH := initial.AutoSSHKey
	keyPath := initial.PrivateKey
	if keyPath != "" {
		autoSSH = false
	}
	var keyLbl *walk.Label

	updateKeyLabel := func() {
		if keyLbl == nil {
			return
		}
		if autoSSH {
			keyLbl.SetText(i18n.T(i18n.KeyFormAutoSSHKey))
		} else if keyPath != "" {
			keyLbl.SetText(i18n.Tf(i18n.KeyFormKeySelected, filepath.Base(keyPath)))
		} else {
			keyLbl.SetText(i18n.T(i18n.KeyFormKeyNone))
		}
	}

	port := initial.Port
	if port == 0 {
		port = config.DefaultSFTPPort
	}
	hb := initial.HeartbeatSec
	if hb == 0 {
		hb = config.DefaultHeartbeatSec
	}

	buildServer := func() (config.Server, bool) {
		var p, heartbeat int
		fmt.Sscanf(portEdit.Text(), "%d", &p)
		fmt.Sscanf(strings.TrimSpace(hbEdit.Text()), "%d", &heartbeat)
		kp := ""
		if !autoSSH {
			kp = strings.TrimSpace(keyPath)
		}
		s := config.Server{
			ID:           initial.ID,
			Name:         strings.TrimSpace(nameEdit.Text()),
			Host:         strings.TrimSpace(hostEdit.Text()),
			Port:         p,
			Username:     strings.TrimSpace(userEdit.Text()),
			Password:     passEdit.Text(),
			AutoSSHKey:   autoSSH,
			PrivateKey:   kp,
			RemoteRoot:   strings.TrimSpace(rootEdit.Text()),
			LocalRoot:    initial.LocalRoot,
			HeartbeatSec: heartbeat,
		}
		if s.Host == "" || s.Username == "" {
			a.showMsg(i18n.T(i18n.KeyServerFormTitle), i18n.T(i18n.KeyFormRequired))
			return s, false
		}
		return s, true
	}

	children := []Widget{
		Label{Text: i18n.T(i18n.KeyFormName)},
		LineEdit{AssignTo: &nameEdit, Text: initial.Name},
		Label{Text: i18n.T(i18n.KeyFormHost)},
		LineEdit{AssignTo: &hostEdit, Text: initial.Host},
		Label{Text: i18n.T(i18n.KeyFormPort)},
		LineEdit{AssignTo: &portEdit, Text: fmt.Sprintf("%d", port)},
		Label{Text: i18n.T(i18n.KeyFormUsername)},
		LineEdit{AssignTo: &userEdit, Text: initial.Username},
		Label{Text: i18n.T(i18n.KeyFormPassword)},
		LineEdit{AssignTo: &passEdit, Text: initial.Password, PasswordMode: true},
		CheckBox{
			AssignTo: &autoKeyCheck,
			Text:     i18n.T(i18n.KeyFormAutoSSHKey),
			Checked:  autoSSH,
			OnCheckedChanged: func() {
				autoSSH = autoKeyCheck.Checked()
				if autoSSH {
					keyPath = ""
				}
				updateKeyLabel()
			},
		},
		Composite{
			Layout: HBox{},
			Children: []Widget{
				PushButton{
					Text: i18n.T(i18n.KeyFormSelectKey),
					OnClicked: func() {
						fd := walk.FileDialog{Title: i18n.T(i18n.KeyFormSelectKey), Filter: "Key files (*.pem;*.ppk;id_*;*.*)|*.pem;*.ppk;id_*|All (*.*)|*.*"}
						if ok, _ := fd.ShowOpen(a.mw); ok {
							keyPath = fd.FilePath
							autoSSH = false
							autoKeyCheck.SetChecked(false)
							updateKeyLabel()
						}
					},
				},
				Label{AssignTo: &keyLbl},
			},
		},
		Label{Text: i18n.T(i18n.KeyFormRemoteRoot)},
		LineEdit{AssignTo: &rootEdit, Text: initial.RemoteRoot},
		Label{Text: i18n.T(i18n.KeyFormHeartbeat)},
		LineEdit{AssignTo: &hbEdit, Text: strconv.Itoa(hb)},
	}

	var buttons []Widget
	buttons = append(buttons, PushButton{Text: i18n.T(i18n.KeyCancel), OnClicked: func() { dlg.Cancel() }})
	if editMode {
		buttons = append(buttons, PushButton{
			Text: i18n.T(i18n.KeySave),
			OnClicked: func() {
				s, ok := buildServer()
				if !ok {
					return
				}
				if editIdx >= 0 {
					s.ID = a.store.Servers[editIdx].ID
					a.store.Servers[editIdx] = s
				}
				_ = a.saveServers()
				dlg.Accept()
			},
		})
	} else {
		buttons = append(buttons, PushButton{
			Text: i18n.T(i18n.KeyFormConnectOnly),
			OnClicked: func() {
				s, ok := buildServer()
				if !ok {
					return
				}
				s.ID = fmt.Sprintf("tmp-%d", time.Now().UnixNano())
				dlg.Accept()
				a.openServerTab(s)
			},
		})
		buttons = append(buttons, PushButton{
			Text: i18n.T(i18n.KeyFormSaveConnect),
			OnClicked: func() {
				s, ok := buildServer()
				if !ok {
					return
				}
				s.ID = fmt.Sprintf("srv-%d", time.Now().UnixNano())
				a.store.Servers = append(a.store.Servers, s)
				_ = a.saveServers()
				dlg.Accept()
				a.openServerTab(s)
			},
		})
	}

	_, _ = Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyServerFormTitle),
		MinSize:  Size{480, 520},
		Layout:   VBox{},
		Children: append(children, Composite{
			Layout: HBox{},
			Children: append([]Widget{HSpacer{}}, buttons...),
		}),
	}.Run(a.mw)

	updateKeyLabel()
}

func (a *App) showMyServers() {
	var dlg *walk.Dialog
	var lb *walk.ListBox
	selected := -1
	names := make([]string, len(a.store.Servers))
	for i, s := range a.store.Servers {
		names[i] = serverDisplayName(s) + "  (" + serverSubtitle(s) + ")"
	}

	_, _ = Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyMyServersTitle),
		MinSize:  Size{520, 420},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: i18n.T(i18n.KeyMyServersHint)},
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
					PushButton{Text: i18n.T(i18n.KeyAddServer), OnClicked: func() {
						dlg.Cancel()
						a.showAddServer()
					}},
					PushButton{Text: i18n.T(i18n.KeyEdit), OnClicked: func() {
						if selected < 0 {
							a.showMsg(i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseEdit))
							return
						}
						dlg.Cancel()
						a.showEditServer(selected)
					}},
					PushButton{Text: i18n.T(i18n.KeyDelete), OnClicked: func() {
						if selected < 0 {
							a.showMsg(i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseDelete))
							return
						}
						dlg.Cancel()
						a.deleteServer(selected)
					}},
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyConnect), OnClicked: func() {
						if selected < 0 {
							return
						}
						dlg.Cancel()
						a.openServerTab(a.store.Servers[selected])
					}},
					PushButton{Text: i18n.T(i18n.KeyCancel), OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}.Run(a.mw)
}

func (a *App) deleteServer(idx int) {
	if idx < 0 || idx >= len(a.store.Servers) {
		return
	}
	name := serverDisplayName(a.store.Servers[idx])
	a.showConfirm(i18n.T(i18n.KeyDelete), i18n.Tf(i18n.KeyDeleteConfirm, name), func() {
		a.store.Servers = append(a.store.Servers[:idx], a.store.Servers[idx+1:]...)
		_ = a.saveServers()
	})
}
