package walkui

import (
	"fmt"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/remote"
)

type tabState int

const (
	tabDisconnected tabState = iota
	tabConnecting
	tabConnected
)

type TabSession struct {
	server     config.Server
	client     *remote.Client
	state      tabState
	remotePath string
	localPath  string
}

func (s *TabSession) tabLabel() string {
	name := s.server.Name
	if name == "" {
		name = s.server.Host
	}
	if len(name) > 12 {
		name = name[:10] + "…"
	}
	host := s.server.Host
	if len(host) > 10 {
		host = host[:8] + "…"
	}
	return name + " @ " + host
}

func (a *App) activeSession() *TabSession {
	if a.activeTab < 0 || a.activeTab >= len(a.tabs) {
		return nil
	}
	return a.tabs[a.activeTab]
}

func (a *App) activeClient() *remote.Client {
	s := a.activeSession()
	if s == nil {
		return nil
	}
	return s.client
}

func (a *App) requireClient() (*remote.Client, bool) {
	c := a.activeClient()
	if c == nil {
		a.showMsg("RelayPane", "Not connected.")
		return nil, false
	}
	return c, true
}

func (a *App) openServerTab(s config.Server) {
	tab := &TabSession{
		server:     s,
		state:      tabDisconnected,
		remotePath: defaultRemoteRoot(&s),
		localPath:  a.localPath,
	}
	if s.LocalRoot != "" {
		tab.localPath = s.LocalRoot
	}
	a.tabs = append(a.tabs, tab)
	a.activeTab = len(a.tabs) - 1
	a.refreshTabBar()
	a.applySessionToUI()
	a.connectTab(tab)
}

func (a *App) activateTab(index int) {
	if index < 0 || index >= len(a.tabs) {
		return
	}
	if a.activeTab >= 0 && a.activeTab < len(a.tabs) && a.activeTab != index {
		a.tabs[a.activeTab].localPath = a.localPath
	}
	a.activeTab = index
	tab := a.tabs[index]
	if tab.localPath != "" {
		a.localPath = tab.localPath
	}
	a.refreshTabBar()
	a.applySessionToUI()
}

func (a *App) closeTab(index int) {
	if index < 0 || index >= len(a.tabs) {
		return
	}
	if index == a.activeTab {
		a.tabs[index].localPath = a.localPath
	}
	tab := a.tabs[index]
	if tab.client != nil {
		_ = tab.client.Close()
	}
	a.tabs = append(a.tabs[:index], a.tabs[index+1:]...)
	if len(a.tabs) == 0 {
		a.activeTab = -1
		a.client = nil
		a.connected = false
		a.remotePath = "/"
		a.refreshTabBar()
		a.refreshRemote()
		a.updateStatusBar()
		return
	}
	if a.activeTab >= len(a.tabs) {
		a.activeTab = len(a.tabs) - 1
	}
	if a.activeTab == index || index <= a.activeTab {
		a.activateTab(a.activeTab)
	} else {
		a.refreshTabBar()
	}
}

func (a *App) applySessionToUI() {
	tab := a.activeSession()
	if tab == nil {
		a.client = nil
		a.connected = false
		a.remotePath = "/"
		a.refreshLocal()
		a.refreshRemote()
		a.updateStatusBar()
		return
	}
	a.server = tab.server
	a.client = tab.client
	a.connected = tab.state == tabConnected && tab.client != nil
	a.remotePath = tab.remotePath
	a.refreshLocal()
	a.refreshRemote()
	a.updateStatusBar()
}

func (a *App) connectTab(tab *TabSession) {
	if tab.client != nil {
		_ = tab.client.Close()
		tab.client = nil
	}
	tab.state = tabConnecting
	if a.activeSession() == tab {
		a.connected = false
		a.client = nil
		a.updateStatusBar()
		a.refreshRemote()
	}

	go func() {
		s := tab.server
		client, err := a.dialServer(s)
		a.syncUI(func() {
			if a.activeSession() != tab && tab.client == nil {
				// tab still valid but not active
			}
			if a.tabs == nil {
				if client != nil {
					_ = client.Close()
				}
				return
			}
			found := false
			for _, t := range a.tabs {
				if t == tab {
					found = true
					break
				}
			}
			if !found {
				if client != nil {
					_ = client.Close()
				}
				return
			}
			if err != nil {
				tab.state = tabDisconnected
				if a.activeSession() == tab {
					a.applySessionToUI()
				}
				a.showError("Connection failed", err)
				a.refreshTabBar()
				return
			}
			tab.client = client
			tab.state = tabConnected
			tab.remotePath = defaultRemoteRoot(&s)
			if interval := s.HeartbeatInterval(); interval > 0 {
				client.StartHeartbeat(interval, func(err error) {
					a.syncUI(func() { a.handleHeartbeatFailure(tab, err) })
				})
			}
			if a.activeSession() == tab {
				a.applySessionToUI()
			}
			a.refreshTabBar()
		})
	}()
}

func (a *App) handleHeartbeatFailure(tab *TabSession, err error) {
	if tab.client != nil {
		_ = tab.client.Close()
		tab.client = nil
	}
	tab.state = tabDisconnected
	if a.activeSession() == tab {
		a.applySessionToUI()
		if a.reconnectBtn != nil {
			a.reconnectBtn.SetVisible(true)
		}
	}
	a.refreshTabBar()
	if err != nil {
		a.showError("Connection lost", fmt.Errorf("%s: %w", tab.server.Name, err))
	}
}

func (a *App) onNewTab() {
	if len(a.store.Servers) > 0 {
		a.showConnectDialog()
		return
	}
	a.showAddServer()
}

func (a *App) reconnectActiveTab() {
	tab := a.activeSession()
	if tab == nil {
		a.onNewTab()
		return
	}
	if tab.state == tabConnecting {
		return
	}
	a.connectTab(tab)
}

func (a *App) saveServers() error {
	return config.Save(a.store)
}
