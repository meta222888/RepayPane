package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window

	settings     *config.Settings
	store        *config.Store
	client       *remote.Client
	activeServer *config.Server

	serverList *widget.List
	status     *widget.Label
	selectedServerID int

	sidebarTitle *widget.Label
	btnAddServer *widget.Button
	btnEdit      *widget.Button
	btnDelete    *widget.Button
	btnRefresh   *widget.Button
	btnDisconnect *widget.Button
	btnUpload    *widget.Button
	btnDownload  *widget.Button

	localPane  *FilePane
	remotePane *FilePane
}

func NewApp(a fyne.App, w fyne.Window) *App {
	store, err := config.Load()
	if err != nil {
		dialog.ShowError(err, w)
		store = &config.Store{}
	}

	settings, err := config.LoadSettings()
	if err != nil {
		dialog.ShowError(err, w)
		settings = &config.Settings{}
	}
	initLanguage(settings)
	if settings.Language != "" {
		_ = config.SaveSettings(settings)
	}

	appUI := &App{
		fyneApp:  a,
		window:   w,
		settings: settings,
		store:    store,
		status:   widget.NewLabel(i18n.T(i18n.KeyNotConnected)),
	}

	appUI.localPane = NewLocalPane(appUI)
	appUI.remotePane = NewRemotePane(appUI)

	appUI.sidebarTitle = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	appUI.btnAddServer = widget.NewButton("", appUI.showAddServer)
	appUI.btnEdit = widget.NewButton("", appUI.showEditServer)
	appUI.btnDelete = widget.NewButton("", appUI.showDeleteServer)
	appUI.btnRefresh = widget.NewButton("", appUI.refreshPanes)
	appUI.btnDisconnect = widget.NewButton("", appUI.disconnect)
	appUI.btnUpload = widget.NewButton("", appUI.uploadSelectedLocal)
	appUI.btnDownload = widget.NewButton("", appUI.downloadSelectedRemote)

	appUI.serverList = widget.NewList(
		func() int { return len(appUI.store.Servers) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(appUI.store.Servers[id].Name)
		},
	)
	appUI.serverList.OnSelected = func(id widget.ListItemID) {
		appUI.selectedServerID = int(id)
		appUI.connectServer(&appUI.store.Servers[id])
	}

	sidebar := container.NewBorder(
		container.NewVBox(
			appUI.sidebarTitle,
			appUI.btnAddServer,
			appUI.btnEdit,
			appUI.btnDelete,
		),
		nil, nil, nil,
		appUI.serverList,
	)

	panes := container.NewHSplit(appUI.localPane.Container(), appUI.remotePane.Container())
	panes.SetOffset(0.5)

	toolbar := container.NewHBox(
		appUI.btnRefresh,
		appUI.btnDisconnect,
		appUI.btnUpload,
		appUI.btnDownload,
	)

	content := container.NewBorder(
		container.NewVBox(toolbar, appUI.status),
		nil, sidebar, nil,
		panes,
	)

	w.SetContent(content)
	w.SetOnDropped(appUI.onWindowDropped)
	appUI.applyLanguage()
	return appUI
}

func (a *App) applyLanguage() {
	a.window.SetTitle(i18n.T(i18n.KeyAppTitle))
	a.buildMainMenu()

	a.sidebarTitle.SetText(i18n.T(i18n.KeyServers))
	a.btnAddServer.SetText(i18n.T(i18n.KeyAddServer))
	a.btnEdit.SetText(i18n.T(i18n.KeyEdit))
	a.btnDelete.SetText(i18n.T(i18n.KeyDelete))
	a.btnRefresh.SetText(i18n.T(i18n.KeyRefresh))
	a.btnDisconnect.SetText(i18n.T(i18n.KeyDisconnect))
	a.btnUpload.SetText(i18n.T(i18n.KeyUpload))
	a.btnDownload.SetText(i18n.T(i18n.KeyDownload))

	a.localPane.ApplyLanguage()
	a.remotePane.ApplyLanguage()

	if a.client == nil {
		a.status.SetText(i18n.T(i18n.KeyNotConnected))
	}
}

func (a *App) onWindowDropped(pos fyne.Position, uris []fyne.URI) {
	size := a.window.Content().Size()
	remoteArea := pos.X > size.Width/2
	a.handleDrop(pos, uris, remoteArea)
}

func (a *App) connectServer(s *config.Server) {
	if a.client != nil {
		_ = a.client.Close()
		a.client = nil
	}

	a.status.SetText(i18n.Tf(i18n.KeyConnecting, s.Name))
	go func() {
		client, err := remote.Connect(remote.ConnectOptions{
			Host:       s.Host,
			Port:       s.Port,
			Username:   s.Username,
			Password:   s.Password,
			PrivateKey: s.PrivateKey,
		})
		fyne.Do(func() {
			if err != nil {
				a.status.SetText(i18n.T(i18n.KeyConnectionFailed))
				dialog.ShowError(err, a.window)
				return
			}
			a.client = client
			a.activeServer = s
			root := s.RemoteRoot
			if root == "" {
				root = "/"
			}
			a.remotePane.SetConnected(true)
			a.remotePane.Navigate(root)

			interval := s.HeartbeatInterval()
			if interval > 0 {
				client.StartHeartbeat(interval, func(err error) {
					fyne.Do(func() {
						a.handleHeartbeatFailure(s, err)
					})
				})
			}

			status := i18n.Tf(i18n.KeyConnected, s.Name, s.Username, s.Host)
			if interval > 0 {
				status += i18n.Tf(i18n.KeyHeartbeatSuffix, int(interval.Seconds()))
			}
			a.status.SetText(status)
		})
	}()
}

func (a *App) handleHeartbeatFailure(s *config.Server, err error) {
	if a.client == nil || a.activeServer == nil || a.activeServer.ID != s.ID {
		return
	}
	_ = a.client.Close()
	a.client = nil
	a.activeServer = nil
	a.remotePane.SetConnected(false)
	a.status.SetText(i18n.T(i18n.KeyConnectionLost))
	dialog.ShowError(fmt.Errorf("connection to %s lost: %w", s.Name, err), a.window)
}

func (a *App) disconnect() {
	if a.client != nil {
		_ = a.client.Close()
		a.client = nil
	}
	a.activeServer = nil
	a.remotePane.SetConnected(false)
	a.status.SetText(i18n.T(i18n.KeyDisconnected))
}

func (a *App) refreshPanes() {
	a.localPane.RefreshListing()
	a.remotePane.RefreshListing()
}

func (a *App) openRemoteEditor(entry remote.FileInfo) {
	if a.client == nil {
		return
	}
	if entry.Size > config.MaxEditBytes {
		dialog.ShowConfirm(
			i18n.T(i18n.KeyFileTooLarge),
			i18n.Tf(i18n.KeyFileTooLargeMsg, entry.Name, float64(entry.Size)/(1024*1024)),
			func(ok bool) {
				if ok {
					a.loadEditor(entry)
				}
			},
			a.window,
		)
		return
	}
	a.loadEditor(entry)
}

func (a *App) loadEditor(entry remote.FileInfo) {
	go func() {
		data, err := a.client.ReadFile(entry.Path)
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			ShowEditor(a, entry, string(data))
		})
	}()
}

func (a *App) saveServers() {
	if err := config.Save(a.store); err != nil {
		dialog.ShowError(err, a.window)
	}
	a.serverList.Refresh()
}

func (a *App) showAddServer() {
	showServerForm(a, config.Server{
		Port:         config.DefaultSFTPPort,
		HeartbeatSec: config.DefaultHeartbeatSec,
	}, func(s config.Server) {
		s.ID = fmt.Sprintf("srv-%d", time.Now().UnixNano())
		a.store.Servers = append(a.store.Servers, s)
		a.saveServers()
	})
}

func (a *App) showEditServer() {
	id := a.selectedServerID
	if id < 0 || id >= len(a.store.Servers) {
		dialogShow(a, i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseEdit))
		return
	}
	s := a.store.Servers[id]
	showServerForm(a, s, func(updated config.Server) {
		updated.ID = s.ID
		a.store.Servers[id] = updated
		a.saveServers()
		if a.activeServer != nil && a.activeServer.ID == s.ID {
			a.connectServer(&a.store.Servers[id])
		}
	})
}

func (a *App) showDeleteServer() {
	id := a.selectedServerID
	if id < 0 || id >= len(a.store.Servers) {
		dialogShow(a, i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseDelete))
		return
	}
	name := a.store.Servers[id].Name
	dialog.ShowConfirm(i18n.T(i18n.KeyDelete), i18n.Tf(i18n.KeyDeleteConfirm, name), func(ok bool) {
		if !ok {
			return
		}
		a.store.Servers = append(a.store.Servers[:id], a.store.Servers[id+1:]...)
		a.saveServers()
	}, a.window)
}

func showServerForm(a *App, initial config.Server, onSave func(config.Server)) {
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
	keyPath := widget.NewEntry()
	keyPath.SetText(initial.PrivateKey)
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

	form := dialog.NewForm(i18n.T(i18n.KeyServerFormTitle), i18n.T(i18n.KeySave), i18n.T(i18n.KeyCancel), []*widget.FormItem{
		{Text: i18n.T(i18n.KeyFormName), Widget: name},
		{Text: i18n.T(i18n.KeyFormHost), Widget: host},
		{Text: i18n.T(i18n.KeyFormPort), Widget: port},
		{Text: i18n.T(i18n.KeyFormUsername), Widget: user},
		{Text: i18n.T(i18n.KeyFormPassword), Widget: pass},
		{Text: i18n.T(i18n.KeyFormPrivateKey), Widget: keyPath},
		{Text: i18n.T(i18n.KeyFormRemoteRoot), Widget: root},
		{Text: i18n.T(i18n.KeyFormHeartbeat), Widget: heartbeat},
	}, func(ok bool) {
		if !ok {
			return
		}
		var p int
		fmt.Sscanf(port.Text, "%d", &p)
		var hb int
		fmt.Sscanf(strings.TrimSpace(heartbeat.Text), "%d", &hb)
		onSave(config.Server{
			Name:         strings.TrimSpace(name.Text),
			Host:         strings.TrimSpace(host.Text),
			Port:         p,
			Username:     strings.TrimSpace(user.Text),
			Password:     pass.Text,
			PrivateKey:   strings.TrimSpace(keyPath.Text),
			RemoteRoot:   strings.TrimSpace(root.Text),
			HeartbeatSec: hb,
		})
	}, a.window)
	form.Resize(fyne.NewSize(480, 460))
	form.Show()
}

type localEntry struct {
	name  string
	path  string
	size  int64
	isDir bool
	mod   time.Time
}

func listLocal(dir string) ([]localEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]localEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, localEntry{
			name:  e.Name(),
			path:  filepath.Join(dir, e.Name()),
			size:  info.Size(),
			isDir: e.IsDir(),
			mod:   info.ModTime(),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].isDir != out[j].isDir {
			return out[i].isDir
		}
		return strings.ToLower(out[i].name) < strings.ToLower(out[j].name)
	})
	return out, nil
}

func defaultLocalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "C:\\"
	}
	return home
}
