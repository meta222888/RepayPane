package ui

import (
	"fmt"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showConnectPicker() {
	selected := -1
	var dlg *modalDialog

	var list *widget.List
	list = widget.NewList(
		func() int { return len(a.store.Servers) },
		func() fyne.CanvasObject { return newConnectPickerRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			idx := int(id)
			if idx >= len(a.store.Servers) {
				return
			}
			row := obj.(*connectPickerRow)
			s := a.store.Servers[idx]
			name := s.Name
			if name == "" {
				name = s.Host
			}
			row.onPrimary = func() { selectPickerRow(list, &selected, idx) }
			row.onDouble = func() {
				selectPickerRow(list, &selected, idx)
				dlg.Close()
				a.openServerTab(a.store.Servers[idx])
			}
			row.update(name, serverSubtitle(s), idx == selected)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectPickerRow(list, &selected, int(id))
	}

	connectBtn := newAccentButton(i18n.T(i18n.KeyConnect), func() {
		if selected < 0 || selected >= len(a.store.Servers) {
			return
		}
		dlg.Close()
		a.openServerTab(a.store.Servers[selected])
	})

	newBtn := newAccentButton(i18n.T(i18n.KeyNewConnection), func() {
		dlg.Close()
		a.showAddServer()
	})
	cancelBtn := newAccentButton(i18n.T(i18n.KeyCancel), func() { dlg.Close() })

	hint := widget.NewLabel(i18n.T(i18n.KeyConnectPickerHint))
	buttons := container.NewHBox(cancelBtn, newBtn, connectBtn)
	body := container.NewBorder(hint, buttons, nil, nil, list)

	title := i18n.T(i18n.KeyConnectPickerTitle)
	dlg = newModalDialog(a.window, title, fyne.NewSize(520, 420), body)
}

func selectPickerRow(list *widget.List, selected *int, row int) {
	if row < 0 {
		return
	}
	prev := *selected
	*selected = row
	list.Select(widget.ListItemID(row))
	if prev >= 0 && prev != row {
		list.RefreshItem(widget.ListItemID(prev))
	}
	list.RefreshItem(widget.ListItemID(row))
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
