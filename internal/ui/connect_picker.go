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

var serverPickerDialogSize = fyne.NewSize(520, 420)

func (a *App) buildServerPickerList(selected *int, onActivate func(idx int)) *widget.List {
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
			name := serverDisplayName(s)
			row.onPrimary = func() { selectPickerRow(list, selected, idx) }
			row.onDouble = func() {
				selectPickerRow(list, selected, idx)
				if onActivate != nil {
					onActivate(idx)
				}
			}
			row.update(idx, serverPickerIcon, name, serverSubtitle(s), idx == *selected)
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		selectPickerRow(list, selected, int(id))
	}
	return list
}

func serverPickerBody(hint string, list fyne.CanvasObject, buttons fyne.CanvasObject) fyne.CanvasObject {
	hintLbl := widget.NewLabel(hint)
	hintLbl.Wrapping = fyne.TextWrapWord
	return container.NewBorder(hintLbl, buttons, nil, nil, list)
}

func (a *App) connectSelectedServer(dlg *modalDialog, selected int) {
	if selected < 0 || selected >= len(a.store.Servers) {
		return
	}
	dlg.Close()
	a.openServerTab(a.store.Servers[selected])
}

func (a *App) showConnectPicker() {
	selected := -1
	var dlg *modalDialog

	list := a.buildServerPickerList(&selected, func(idx int) {
		a.connectSelectedServer(dlg, idx)
	})

	connectBtn := newAccentButton(i18n.T(i18n.KeyConnect), func() {
		a.connectSelectedServer(dlg, selected)
	})
	newBtn := newAccentButton(i18n.T(i18n.KeyNewConnection), func() {
		dlg.Close()
		a.showAddServer()
	})
	cancelBtn := newAccentButton(i18n.T(i18n.KeyCancel), func() { dlg.Close() })

	buttons := container.NewHBox(cancelBtn, newBtn, connectBtn)
	body := serverPickerBody(i18n.T(i18n.KeyConnectPickerHint), list, buttons)
	dlg = newModalDialog(a, i18n.T(i18n.KeyConnectPickerTitle), serverPickerDialogSize, body)
}

func selectPickerRow(list *widget.List, selected *int, row int) {
	if row < 0 {
		return
	}
	prev := *selected
	if prev == row {
		return
	}
	*selected = row
	list.Select(widget.ListItemID(row))
	if prev >= 0 {
		list.RefreshItem(widget.ListItemID(prev))
	}
	list.RefreshItem(widget.ListItemID(row))
}

func serverDisplayName(s config.Server) string {
	if s.Name != "" {
		return s.Name
	}
	return s.Host
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
