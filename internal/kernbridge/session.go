package kernbridge

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/remote"
)

func resolveTabLocalPath(s config.Server) string {
	if p := strings.TrimSpace(s.LocalRoot); p != "" {
		return p
	}
	return defaultLocalDir()
}

func (a *App) OpenServerTab(serverIndex int) {
	if serverIndex < 0 || serverIndex >= len(a.store.Servers) {
		return
	}
	s := a.store.Servers[serverIndex]
	tab := &TabSession{
		server:     s,
		state:      tabDisconnected,
		remotePath: defaultRemoteRoot(&s),
		localPath:  resolveTabLocalPath(s),
	}
	a.localPath = tab.localPath
	a.tabs = append(a.tabs, tab)
	a.activeTab = len(a.tabs) - 1
	a.host.refreshTabs()
	a.applySessionToUI()
	a.connectTab(tab)
}

func (a *App) OpenServerByJSON(serverJSON string) error {
	var s config.Server
	if err := jsonUnmarshal(serverJSON, &s); err != nil {
		return err
	}
	if s.ID == "" {
		s.ID = fmt.Sprintf("srv-%d", time.Now().UnixNano())
	}
	tab := &TabSession{
		server:     s,
		state:      tabDisconnected,
		remotePath: defaultRemoteRoot(&s),
		localPath:  resolveTabLocalPath(s),
	}
	a.localPath = tab.localPath
	a.tabs = append(a.tabs, tab)
	a.activeTab = len(a.tabs) - 1
	a.host.refreshTabs()
	a.applySessionToUI()
	a.connectTab(tab)
	return nil
}

func (a *App) ActivateTab(index int) {
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
	a.host.refreshTabs()
	a.applySessionToUI()
}

func (a *App) CloseTab(index int) {
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
		a.host.refreshTabs()
		a.host.refreshRemote()
		a.host.updateStatusBar()
		return
	}
	if a.activeTab >= len(a.tabs) {
		a.activeTab = len(a.tabs) - 1
	}
	if a.activeTab == index || index <= a.activeTab {
		a.ActivateTab(a.activeTab)
	} else {
		a.host.refreshTabs()
	}
}

func (a *App) applySessionToUI() {
	tab := a.activeSession()
	if tab == nil {
		a.client = nil
		a.connected = false
		a.remotePath = "/"
		a.host.refreshLocal()
		a.host.refreshRemote()
		a.host.updateStatusBar()
		return
	}
	a.server = tab.server
	a.client = tab.client
	a.connected = tab.state == tabConnected && tab.client != nil
	a.remotePath = tab.remotePath
	a.host.refreshLocal()
	a.host.refreshRemote()
	a.host.updateStatusBar()
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
		a.host.updateStatusBar()
		a.host.refreshRemote()
	}

	go func() {
		s := tab.server
		client, err := a.dialServer(s)
		a.host.sync(func() {
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
				a.host.showError("Connection failed", err)
				a.host.refreshTabs()
				return
			}
			tab.client = client
			tab.state = tabConnected
			tab.remotePath = defaultRemoteRoot(&s)
			if interval := s.HeartbeatInterval(); interval > 0 {
				client.StartHeartbeat(interval, func(err error) {
					a.host.sync(func() { a.handleHeartbeatFailure(tab, err) })
				})
			}
			if a.activeSession() == tab {
				a.applySessionToUI()
			}
			a.host.refreshTabs()
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
	}
	a.host.refreshTabs()
	if err != nil {
		a.host.showError("Connection lost", fmt.Errorf("%s: %w", tab.server.Name, err))
	}
}

func (a *App) ReconnectActiveTab() {
	tab := a.activeSession()
	if tab == nil {
		return
	}
	if tab.state == tabConnecting {
		return
	}
	a.connectTab(tab)
}

func (a *App) NavigateLocal(p string) {
	if p == "" {
		return
	}
	a.localPath = ensureTrailingSlash(p)
	if tab := a.activeSession(); tab != nil {
		tab.localPath = a.localPath
		tab.server.LocalRoot = a.localPath
		a.saveLocalPathForActiveServer()
	}
	a.host.refreshLocal()
}

func (a *App) NavigateRemote(p string) {
	if !a.connected || a.client == nil {
		return
	}
	if p == "" {
		p = "/"
	}
	a.remotePath = path.Clean(strings.ReplaceAll(p, `\`, `/`))
	if a.remotePath == "." {
		a.remotePath = "/"
	}
	if tab := a.activeSession(); tab != nil {
		tab.remotePath = a.remotePath
	}
	a.host.refreshRemote()
}

func (a *App) LocalUp() {
	a.NavigateLocal(filepath.Dir(strings.TrimSuffix(a.localPath, `\`)))
}

func (a *App) RemoteUp() {
	if a.remotePath == "/" {
		return
	}
	a.NavigateRemote(path.Dir(a.remotePath))
}

func (a *App) saveLocalPathForActiveServer() {
	tab := a.activeSession()
	if tab == nil || strings.HasPrefix(tab.server.ID, "tmp-") {
		return
	}
	for i := range a.store.Servers {
		if a.store.Servers[i].ID == tab.server.ID {
			a.store.Servers[i].LocalRoot = a.localPath
			_ = a.saveServers()
			return
		}
	}
}

func listRemoteDirJSON(client *remote.Client, dir string) ([]DirEntryJSON, error) {
	files, err := client.ListDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]DirEntryJSON, 0, len(files)+1)
	if dir != "/" {
		out = append(out, DirEntryJSON{Name: "..", Path: path.Dir(dir), IsDir: true})
	}
	for _, f := range files {
		out = append(out, DirEntryJSON{
			Name:    f.Name,
			Path:    f.Path,
			Size:    f.Size,
			IsDir:   f.IsDir,
			ModTime: formatModTime(f.ModTime),
			Mode:    formatMode(f.Mode),
			Owner:   "root",
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].nameIsParent() {
			return true
		}
		if out[j].nameIsParent() {
			return false
		}
		if out[i].IsDir != out[j].IsDir {
			return out[i].IsDir
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (a *App) AddServerJSON(serverJSON string) error {
	var s config.Server
	if err := jsonUnmarshal(serverJSON, &s); err != nil {
		return err
	}
	if s.ID == "" {
		s.ID = fmt.Sprintf("srv-%d", time.Now().UnixNano())
	}
	if s.Port == 0 {
		s.Port = config.DefaultSFTPPort
	}
	if s.HeartbeatSec == 0 {
		s.HeartbeatSec = config.DefaultHeartbeatSec
	}
	a.store.Servers = append(a.store.Servers, s)
	return a.saveServers()
}

func (a *App) UpdateServerJSON(index int, serverJSON string) error {
	if index < 0 || index >= len(a.store.Servers) {
		return fmt.Errorf("invalid server index")
	}
	var s config.Server
	if err := jsonUnmarshal(serverJSON, &s); err != nil {
		return err
	}
	oldID := a.store.Servers[index].ID
	a.store.Servers[index] = s
	if s.ID == "" {
		a.store.Servers[index].ID = oldID
	}
	for _, tab := range a.tabs {
		if tab.server.ID == oldID {
			tab.server = a.store.Servers[index]
			if tab.client != nil {
				_ = tab.client.Close()
				tab.client = nil
			}
			tab.state = tabDisconnected
		}
	}
	return a.saveServers()
}

func (a *App) DeleteServer(index int) error {
	if index < 0 || index >= len(a.store.Servers) {
		return fmt.Errorf("invalid server index")
	}
	id := a.store.Servers[index].ID
	a.store.Servers = append(a.store.Servers[:index], a.store.Servers[index+1:]...)
	for i := len(a.tabs) - 1; i >= 0; i-- {
		if a.tabs[i].server.ID == id {
			a.CloseTab(i)
		}
	}
	return a.saveServers()
}

func jsonUnmarshal(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}
