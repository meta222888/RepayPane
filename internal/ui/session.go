package ui

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
}

func (s *TabSession) tabLabel() string {
	host := s.server.Host
	if len(host) > 12 {
		host = host[:10] + "…"
	}
	name := s.server.Name
	if len(name) > 10 {
		name = name[:8] + "…"
	}
	port := s.server.Port
	if port == 0 {
		port = 22
	}
	return fmt.Sprintf("%s  %s:%d", name, host, port)
}

func (s *TabSession) addr() string {
	port := s.server.Port
	if port == 0 {
		port = 22
	}
	return fmt.Sprintf("%s:%d", s.server.Host, port)
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
