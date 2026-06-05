package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
)

func (a *App) saveLocalPathForTab(tabIndex int) {
	if tabIndex < 0 || tabIndex >= len(a.tabs) {
		return
	}
	a.persistServerLocalPath(a.tabs[tabIndex].server.ID, a.localPane.CurrentPath())
}

func (a *App) saveActiveServerLocalPath() {
	a.saveLocalPathForTab(a.activeTab)
}

func (a *App) persistServerLocalPath(serverID, path string) {
	if serverID == "" || strings.HasPrefix(serverID, "tmp-") {
		return
	}
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" {
		return
	}
	updated := false
	for i := range a.store.Servers {
		if a.store.Servers[i].ID == serverID {
			if a.store.Servers[i].LocalRoot == path {
				return
			}
			a.store.Servers[i].LocalRoot = path
			updated = true
			break
		}
	}
	if !updated {
		return
	}
	for _, tab := range a.tabs {
		if tab.server.ID == serverID {
			tab.server.LocalRoot = path
		}
	}
	a.saveServers()
}

func (a *App) restoreLocalPathForServer(s *config.Server) {
	path := resolveServerLocalRoot(s)
	if path == a.localPane.CurrentPath() {
		return
	}
	a.localPane.Navigate(path)
}

func resolveServerLocalRoot(s *config.Server) string {
	path := strings.TrimSpace(s.LocalRoot)
	if path != "" {
		path = filepath.Clean(path)
		if st, err := os.Stat(path); err == nil && st.IsDir() {
			return path
		}
	}
	return defaultLocalDir()
}

func (a *App) prepareQuit() {
	a.saveActiveServerLocalPath()
	a.fyneApp.Quit()
}
