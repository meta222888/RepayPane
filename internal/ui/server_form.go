package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showServerForm(a *App, initial config.Server, editMode bool, onDone func(config.Server, bool)) {
	name := widget.NewEntry()
	name.SetText(initial.Name)
	host := widget.NewEntry()
	host.SetText(initial.Host)
	port := widget.NewEntry()
	if initial.Port == 0 {
		initial.Port = config.DefaultSFTPPort
	}
	port.SetText(fmt.Sprintf("%d", initial.Port))
	user := widget.NewEntry()
	user.SetText(initial.Username)
	pass := widget.NewPasswordEntry()
	pass.SetText(initial.Password)

	var selectedKeyPath string
	autoSSH := initial.AutoSSHKey
	if initial.PrivateKey != "" {
		selectedKeyPath = initial.PrivateKey
		autoSSH = false
	} else if initial.ID == "" {
		autoSSH = true
	}

	keyStatus := widget.NewLabel(i18n.T(i18n.KeyFormKeyNone))
	if selectedKeyPath != "" {
		keyStatus.SetText(i18n.Tf(i18n.KeyFormKeySelected, filepath.Base(selectedKeyPath)))
	}

	updateKeyUI := func() {}
	autoCheck := widget.NewCheck(i18n.T(i18n.KeyFormAutoSSHKey), func(checked bool) {
		autoSSH = checked
		if checked {
			selectedKeyPath = ""
		}
		updateKeyUI()
	})
	autoCheck.SetChecked(autoSSH)

	selectBtn := widget.NewButton(i18n.T(i18n.KeyFormSelectKey), func() {
		showSSHKeyPicker(a, selectedKeyPath, func(path string) {
			selectedKeyPath = strings.TrimSpace(path)
			autoSSH = false
			autoCheck.SetChecked(false)
			updateKeyUI()
		})
	})

	updateKeyUI = func() {
		if autoSSH {
			keyStatus.SetText(i18n.T(i18n.KeyFormAutoSSHKey))
			selectBtn.Disable()
			return
		}
		selectBtn.Enable()
		if selectedKeyPath != "" {
			keyStatus.SetText(i18n.Tf(i18n.KeyFormKeySelected, filepath.Base(selectedKeyPath)))
		} else {
			keyStatus.SetText(i18n.T(i18n.KeyFormKeyNone))
		}
	}
	updateKeyUI()

	keyRow := container.NewVBox(autoCheck, container.NewHBox(selectBtn, keyStatus))

	root := widget.NewEntry()
	root.SetText(initial.RemoteRoot)
	root.SetPlaceHolder("/")

	heartbeat := widget.NewEntry()
	hbSec := initial.HeartbeatSec
	if hbSec == 0 {
		hbSec = config.DefaultHeartbeatSec
	}
	heartbeat.SetText(strconv.Itoa(hbSec))
	heartbeat.SetPlaceHolder("30")

	w := newThemedWindow(a.fyneApp, fyne.NewSize(540, 520))
	title := i18n.T(i18n.KeyServerFormTitle)

	buildServer := func() (config.Server, bool) {
		var p int
		fmt.Sscanf(port.Text, "%d", &p)
		var hb int
		fmt.Sscanf(strings.TrimSpace(heartbeat.Text), "%d", &hb)
		keyPath := ""
		if !autoSSH {
			keyPath = strings.TrimSpace(selectedKeyPath)
		}
		s := config.Server{
			Name:         strings.TrimSpace(name.Text),
			Host:         strings.TrimSpace(host.Text),
			Port:         p,
			Username:     strings.TrimSpace(user.Text),
			Password:     pass.Text,
			AutoSSHKey:   autoSSH,
			PrivateKey:   keyPath,
			RemoteRoot:   strings.TrimSpace(root.Text),
			HeartbeatSec: hb,
		}
		if s.Host == "" || s.Username == "" {
			dialog.ShowInformation(i18n.T(i18n.KeyServerFormTitle), i18n.T(i18n.KeyFormRequired), w)
			return s, false
		}
		return s, true
	}

	form := widget.NewForm(
		widget.NewFormItem(i18n.T(i18n.KeyFormName), name),
		widget.NewFormItem(i18n.T(i18n.KeyFormHost), host),
		widget.NewFormItem(i18n.T(i18n.KeyFormPort), port),
		widget.NewFormItem(i18n.T(i18n.KeyFormUsername), user),
		widget.NewFormItem(i18n.T(i18n.KeyFormPassword), pass),
		widget.NewFormItem(i18n.T(i18n.KeyFormPrivateKey), keyRow),
		widget.NewFormItem(i18n.T(i18n.KeyFormRemoteRoot), root),
		widget.NewFormItem(i18n.T(i18n.KeyFormHeartbeat), heartbeat),
	)

	cancelBtn := widget.NewButton(i18n.T(i18n.KeyCancel), func() { w.Close() })
	var buttons *fyne.Container
	if editMode {
		saveBtn := widget.NewButton(i18n.T(i18n.KeySave), func() {
			s, ok := buildServer()
			if !ok {
				return
			}
			w.Close()
			onDone(s, true)
		})
		saveBtn.Importance = widget.HighImportance
		buttons = container.NewHBox(cancelBtn, saveBtn)
	} else {
		connectOnlyBtn := widget.NewButton(i18n.T(i18n.KeyFormConnectOnly), func() {
			s, ok := buildServer()
			if !ok {
				return
			}
			w.Close()
			onDone(s, false)
		})
		saveBtn := widget.NewButton(i18n.T(i18n.KeyFormSaveConnect), func() {
			s, ok := buildServer()
			if !ok {
				return
			}
			w.Close()
			onDone(s, true)
		})
		saveBtn.Importance = widget.HighImportance
		buttons = container.NewHBox(cancelBtn, connectOnlyBtn, saveBtn)
	}

	body := container.NewBorder(nil, buttons, nil, nil, form)
	w.SetContent(themedWindowChrome(w, title, body))
	w.Show()
}
