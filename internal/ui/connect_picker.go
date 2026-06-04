package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showConnectPicker() {
	selected := -1
	prevSelected := -1
	var lastTap time.Time
	var lastID int

	list := widget.NewList(
		func() int { return len(a.store.Servers) },
		func() fyne.CanvasObject { return newConnectPickerRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row := obj.(*connectPickerRow)
			s := a.store.Servers[id]
			name := s.Name
			if name == "" {
				name = s.Host
			}
			row.update(name, serverSubtitle(s), int(id) == selected)
		},
	)

	title := i18n.T(i18n.KeyConnectPickerTitle)
	w := newThemedWindow(a.fyneApp, fyne.NewSize(520, 420))

	list.OnSelected = func(id widget.ListItemID) {
		idx := int(id)
		now := time.Now()
		if idx == lastID && now.Sub(lastTap) < 500*time.Millisecond {
			w.Close()
			a.openServerTab(a.store.Servers[idx])
			return
		}
		lastTap = now
		lastID = idx
		prevSelected = selected
		selected = idx
		if prevSelected >= 0 {
			list.RefreshItem(widget.ListItemID(prevSelected))
		}
		list.RefreshItem(id)
	}

	connectBtn := widget.NewButton(i18n.T(i18n.KeyConnect), func() {
		if selected < 0 {
			return
		}
		w.Close()
		a.openServerTab(a.store.Servers[selected])
	})
	connectBtn.Importance = widget.HighImportance

	newBtn := widget.NewButton(i18n.T(i18n.KeyNewConnection), func() {
		w.Close()
		a.showAddServer()
	})
	cancelBtn := widget.NewButton(i18n.T(i18n.KeyCancel), func() { w.Close() })

	hint := widget.NewLabel(i18n.T(i18n.KeyConnectPickerHint))
	buttons := container.NewHBox(cancelBtn, newBtn, connectBtn)
	body := container.NewBorder(hint, buttons, nil, nil, list)
	w.SetContent(themedWindowChrome(w, title, body))
	w.Show()
}

func serverSubtitle(s config.Server) string {
	port := s.Port
	if port == 0 {
		port = 22
	}
	if strings.TrimSpace(s.Username) == "" {
		return fmt.Sprintf("%s:%d", s.Host, port)
	}
	return fmt.Sprintf("%s@%s:%d", s.Username, s.Host, port)
}

func (a *App) openServerTab(s config.Server) {
	tab := &TabSession{server: s, state: tabDisconnected, remotePath: defaultRemoteRoot(&s)}
	a.tabs = append(a.tabs, tab)
	a.activeTab = len(a.tabs) - 1
	a.tabBar.Refresh()
	a.activateTab(a.activeTab)
}
