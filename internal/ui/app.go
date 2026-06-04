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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window

	settings *config.Settings
	store    *config.Store

	tabs      []*TabSession
	activeTab int

	topBar    *TopBar
	tabBar    *TabBar
	statusBar *StatusBar
	localPane  *FilePane
	remotePane *FilePane

	selectedServerID int // for my servers dialog
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
	}

	appUI.topBar = NewTopBar(appUI)
	appUI.tabBar = NewTabBar(appUI)
	appUI.statusBar = NewStatusBar(appUI)
	appUI.localPane = NewLocalPane(appUI)
	appUI.remotePane = NewRemotePane(appUI)

	for i := range store.Servers {
		s := store.Servers[i]
		appUI.tabs = append(appUI.tabs, &TabSession{
			server:     s,
			state:      tabDisconnected,
			remotePath: defaultRemoteRoot(&s),
		})
	}
	if len(appUI.tabs) > 0 {
		appUI.activeTab = 0
	}

	panes := container.NewHSplit(appUI.localPane.Container(), appUI.remotePane.Container())
	panes.SetOffset(0.5)

	header := container.NewVBox(appUI.topBar.Container(), appUI.tabBar.Container())
	body := container.NewBorder(header, appUI.statusBar.Container(), nil, nil, panes)
	bg := canvas.NewRectangle(colorBG)
	w.SetPadded(false)
	w.SetContent(container.NewStack(bg, body))
	w.SetOnDropped(appUI.onWindowDropped)
	appUI.applyLanguage()
	appUI.tabBar.Refresh()
	if appUI.activeTab >= 0 {
		appUI.activateTab(appUI.activeTab)
	}
	return appUI
}

func defaultRemoteRoot(s *config.Server) string {
	if s.RemoteRoot != "" {
		return s.RemoteRoot
	}
	return "/"
}

func (a *App) applyLanguage() {
	a.window.SetTitle(i18n.T(i18n.KeyAppTitle))
	a.window.SetMainMenu(nil)
	a.topBar.ApplyLanguage()
	a.localPane.ApplyLanguage()
	a.remotePane.ApplyLanguage()
	a.statusBar.ApplyLanguage()
	a.tabBar.Refresh()
}

func (a *App) onWindowDropped(pos fyne.Position, uris []fyne.URI) {
	size := a.window.Content().Size()
	remoteArea := pos.X > size.Width/2
	a.handleDrop(pos, uris, remoteArea)
}

func (a *App) onNewTab() {
	a.showAddServer()
}

func (a *App) activateTab(index int) {
	if index < 0 || index >= len(a.tabs) {
		return
	}
	a.activeTab = index
	tab := a.tabs[index]
	a.tabBar.Refresh()
	a.statusBar.Refresh()

	if tab.state == tabConnected && tab.client != nil {
		a.remotePane.SetConnected(true)
		a.remotePane.Navigate(tab.remotePath)
		return
	}
	if tab.state == tabConnecting {
		return
	}
	a.connectTab(tab)
}

func (a *App) connectTab(tab *TabSession) {
	if tab.client != nil {
		_ = tab.client.Close()
		tab.client = nil
	}
	tab.state = tabConnecting
	a.tabBar.Refresh()
	a.statusBar.Refresh()

	a.statusBar.conn.SetText(i18n.Tf(i18n.KeyConnecting, tab.server.Name))
	go func() {
		s := tab.server
		client, err := remote.Connect(remote.ConnectOptions{
			Host:       s.Host,
			Port:       s.Port,
			Username:   s.Username,
			Password:   s.Password,
			AutoSSHKey: s.AutoSSHKey,
			PrivateKey: s.PrivateKey,
		})
		fyne.Do(func() {
			if a.activeSession() != tab {
				if client != nil {
					_ = client.Close()
				}
				tab.state = tabDisconnected
				a.tabBar.Refresh()
				return
			}
			if err != nil {
				tab.state = tabDisconnected
				a.statusBar.Refresh()
				dialog.ShowError(err, a.window)
				a.tabBar.Refresh()
				return
			}
			tab.client = client
			tab.state = tabConnected
			tab.remotePath = defaultRemoteRoot(&s)

			interval := s.HeartbeatInterval()
			if interval > 0 {
				client.StartHeartbeat(interval, func(err error) {
					fyne.Do(func() { a.handleHeartbeatFailure(tab, err) })
				})
			}

			a.remotePane.SetConnected(true)
			a.remotePane.Navigate(tab.remotePath)
			a.tabBar.Refresh()
			a.statusBar.Refresh()
		})
	}()
}

func (a *App) closeTab(index int) {
	if index < 0 || index >= len(a.tabs) {
		return
	}
	tab := a.tabs[index]
	if tab.client != nil {
		_ = tab.client.Close()
	}
	a.tabs = append(a.tabs[:index], a.tabs[index+1:]...)
	if len(a.tabs) == 0 {
		a.activeTab = -1
		a.remotePane.SetConnected(false)
		a.tabBar.Refresh()
		a.statusBar.Refresh()
		return
	}
	if a.activeTab >= len(a.tabs) {
		a.activeTab = len(a.tabs) - 1
	}
	if a.activeTab == index || index <= a.activeTab {
		a.activateTab(a.activeTab)
	} else {
		a.tabBar.Refresh()
	}
}

func (a *App) handleHeartbeatFailure(tab *TabSession, err error) {
	if tab.client == nil {
		return
	}
	_ = tab.client.Close()
	tab.client = nil
	tab.state = tabDisconnected
	if a.activeSession() == tab {
		a.remotePane.SetConnected(false)
	}
	a.tabBar.Refresh()
	a.statusBar.Refresh()
	dialog.ShowError(fmt.Errorf("connection to %s lost: %w", tab.server.Name, err), a.window)
}

func (a *App) openRemoteEditor(entry remote.FileInfo) {
	client := a.activeClient()
	if client == nil {
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
	client := a.activeClient()
	if client == nil {
		return
	}
	go func() {
		data, err := client.ReadFile(entry.Path)
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
	a.tabBar.Refresh()
}

func (a *App) showAddServer() {
	showServerForm(a, config.Server{
		Port:         config.DefaultSFTPPort,
		HeartbeatSec: config.DefaultHeartbeatSec,
	}, func(s config.Server) {
		s.ID = fmt.Sprintf("srv-%d", time.Now().UnixNano())
		a.store.Servers = append(a.store.Servers, s)
		a.saveServers()
		tab := &TabSession{server: s, state: tabDisconnected, remotePath: defaultRemoteRoot(&s)}
		a.tabs = append(a.tabs, tab)
		a.activeTab = len(a.tabs) - 1
		a.tabBar.Refresh()
		a.activateTab(a.activeTab)
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
		for _, tab := range a.tabs {
			if tab.server.ID == s.ID {
				tab.server = updated
				if tab.client != nil {
					_ = tab.client.Close()
					tab.client = nil
					tab.state = tabDisconnected
					if a.activeSession() == tab {
						a.activateTab(a.activeTab)
					}
				}
			}
		}
		a.tabBar.Refresh()
	})
}

func (a *App) showDeleteServer() {
	id := a.selectedServerID
	if id < 0 || id >= len(a.store.Servers) {
		dialogShow(a, i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseDelete))
		return
	}
	name := a.store.Servers[id]
	dialog.ShowConfirm(i18n.T(i18n.KeyDelete), i18n.Tf(i18n.KeyDeleteConfirm, name.Name), func(ok bool) {
		if !ok {
			return
		}
		srvID := a.store.Servers[id].ID
		a.store.Servers = append(a.store.Servers[:id], a.store.Servers[id+1:]...)
		a.saveServers()
		for i := len(a.tabs) - 1; i >= 0; i-- {
			if a.tabs[i].server.ID == srvID {
				a.closeTab(i)
			}
		}
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

	var selectedKeyPath string
	autoSSH := initial.AutoSSHKey
	if initial.PrivateKey != "" {
		selectedKeyPath = initial.PrivateKey
		autoSSH = false
	} else if !initial.AutoSSHKey && initial.ID == "" {
		autoSSH = true
	}

	keyStatus := widget.NewLabel(i18n.T(i18n.KeyFormKeyNone))
	if selectedKeyPath != "" {
		keyStatus.SetText(i18n.Tf(i18n.KeyFormKeySelected, filepath.Base(selectedKeyPath)))
	}

	updateKeyUI := func() {}
	autoCheck := widget.NewCheck(i18n.T(i18n.KeyFormAutoSSHKey), func(checked bool) {
		autoSSH = checked
		updateKeyUI()
	})
	autoCheck.SetChecked(autoSSH)

	selectBtn := widget.NewButton(i18n.T(i18n.KeyFormSelectKey), func() {
		d := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			defer rc.Close()
			selectedKeyPath = filepath.FromSlash(rc.URI().Path())
			autoSSH = false
			autoCheck.SetChecked(false)
			updateKeyUI()
		}, a.window)
		d.Show()
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

	form := dialog.NewForm(i18n.T(i18n.KeyServerFormTitle), i18n.T(i18n.KeySave), i18n.T(i18n.KeyCancel), []*widget.FormItem{
		{Text: i18n.T(i18n.KeyFormName), Widget: name},
		{Text: i18n.T(i18n.KeyFormHost), Widget: host},
		{Text: i18n.T(i18n.KeyFormPort), Widget: port},
		{Text: i18n.T(i18n.KeyFormUsername), Widget: user},
		{Text: i18n.T(i18n.KeyFormPassword), Widget: pass},
		{Text: i18n.T(i18n.KeyFormPrivateKey), Widget: keyRow},
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
		keyPath := ""
		if !autoSSH {
			keyPath = strings.TrimSpace(selectedKeyPath)
		}
		onSave(config.Server{
			Name:         strings.TrimSpace(name.Text),
			Host:         strings.TrimSpace(host.Text),
			Port:         p,
			Username:     strings.TrimSpace(user.Text),
			Password:     pass.Text,
			AutoSSHKey:   autoSSH,
			PrivateKey:   keyPath,
			RemoteRoot:   strings.TrimSpace(root.Text),
			HeartbeatSec: hb,
		})
	}, a.window)
	form.Resize(fyne.NewSize(520, 480))
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
