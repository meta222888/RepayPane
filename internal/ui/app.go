package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	transfers *TransferQueue

	clipboard *PaneClipboard
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
	appUI.transfers = NewTransferQueue(appUI)
	appUI.localPane = NewLocalPane(appUI)
	appUI.remotePane = NewRemotePane(appUI)
	appUI.activeTab = -1

	panes := container.NewHSplit(
		appUI.localPane.Container(),
		splitBorder(appUI.remotePane.Container()),
	)
	panes.SetOffset(0.5)

	header := container.NewVBox(
		appUI.topBar.Container(),
		appUI.tabBar.Container(),
	)
	body := container.NewBorder(header, appUI.statusBar.Container(), nil, nil, panes)
	bg := canvas.NewRectangle(colorBG)
	w.SetPadded(false)
	w.SetContent(container.NewStack(bg, body))
	w.SetOnDropped(appUI.onWindowDropped)
	w.SetCloseIntercept(func() {
		appUI.saveActiveServerLocalPath()
		w.SetCloseIntercept(nil)
		w.Close()
	})
	appUI.applyLanguage()
	appUI.tabBar.Refresh()
	appUI.statusBar.Refresh()
	appUI.registerShellShortcut()
	go appUI.installWindowFrame()
	return appUI
}

func (a *App) installWindowFrame() {
	for i := 0; i < 40; i++ {
		time.Sleep(50 * time.Millisecond)
		var ready bool
		fyne.Do(func() {
			ready = winInstallResizeHook(a.window)
		})
		if ready {
			return
		}
	}
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
	if len(a.store.Servers) > 0 {
		a.showConnectPicker()
		return
	}
	a.showAddServer()
}

func (a *App) activateTab(index int) {
	if index < 0 || index >= len(a.tabs) {
		return
	}
	if a.activeTab >= 0 && a.activeTab < len(a.tabs) && a.activeTab != index {
		a.saveLocalPathForTab(a.activeTab)
	}
	a.activeTab = index
	tab := a.tabs[index]
	a.restoreLocalPathForServer(&tab.server)
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
	if strings.TrimSpace(tab.server.Host) == "" || strings.TrimSpace(tab.server.Username) == "" {
		tab.state = tabDisconnected
		a.tabBar.Refresh()
		a.statusBar.Refresh()
		dialog.ShowInformation(i18n.T(i18n.KeyServerFormTitle), i18n.T(i18n.KeyFormRequired), a.window)
		return
	}
	tab.state = tabConnecting
	a.tabBar.Refresh()
	a.statusBar.Refresh()

	a.statusBar.conn.SetText(i18n.Tf(i18n.KeyConnecting, tab.server.Name))
	go func() {
		s := tab.server
		client, err := a.dialServer(s)
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
	if index == a.activeTab {
		a.saveLocalPathForTab(index)
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

func (a *App) dialServer(s config.Server) (*remote.Client, error) {
	var pass []byte
	if remote.NeedsPassphrase(s.AutoSSHKey, s.PrivateKey) {
		pass = a.askPassphraseBlocking()
		if len(pass) == 0 {
			return nil, remote.ErrPassphraseRequired
		}
	}
	client, err := a.tryConnect(s, pass)
	if errors.Is(err, remote.ErrPassphraseRequired) {
		pass = a.askPassphraseBlocking()
		if len(pass) == 0 {
			return nil, err
		}
		return a.tryConnect(s, pass)
	}
	return client, err
}

func (a *App) tryConnect(s config.Server, passphrase []byte) (*remote.Client, error) {
	return remote.Connect(remote.ConnectOptions{
		Host:          s.Host,
		Port:          s.Port,
		Username:      s.Username,
		Password:      s.Password,
		AutoSSHKey:    s.AutoSSHKey,
		PrivateKey:    s.PrivateKey,
		KeyPassphrase: passphrase,
	})
}

func (a *App) showAddServer() {
	showServerForm(a, config.Server{
		Port:         config.DefaultSFTPPort,
		HeartbeatSec: config.DefaultHeartbeatSec,
		AutoSSHKey:   true,
	}, false, func(s config.Server, save bool) {
		if save {
			s.ID = fmt.Sprintf("srv-%d", time.Now().UnixNano())
			a.store.Servers = append(a.store.Servers, s)
			a.saveServers()
		} else {
			s.ID = fmt.Sprintf("tmp-%d", time.Now().UnixNano())
		}
		a.openServerTab(s)
	})
}

func (a *App) showEditServer() {
	id := a.selectedServerID
	if id < 0 || id >= len(a.store.Servers) {
		dialogShow(a, i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseEdit))
		return
	}
	s := a.store.Servers[id]
	showServerForm(a, s, true, func(updated config.Server, save bool) {
		if !save {
			return
		}
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
