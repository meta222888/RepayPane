package ui

import (
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showConnectPicker() {
	w := a.fyneApp.NewWindow(i18n.T(i18n.KeyConnectPickerTitle))
	w.Resize(fyne.NewSize(520, 400))
	w.CenterOnScreen()

	selected := -1
	var lastTap time.Time
	var lastID int

	list := widget.NewList(
		func() int { return len(a.store.Servers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			s := a.store.Servers[id]
			label := s.Username + "@" + s.Host
			if strings.TrimSpace(s.Username) == "" {
				label = s.Host + "  (" + i18n.T(i18n.KeyFormRequired) + ")"
			}
			if s.Name != "" {
				label = s.Name + "    " + label
			}
			obj.(*widget.Label).SetText(label)
		},
	)
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
		selected = idx
		sel := id
		time.AfterFunc(100*time.Millisecond, func() {
			fyne.Do(func() { list.Unselect(sel) })
		})
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

	buttons := container.NewHBox(cancelBtn, newBtn, connectBtn)
	hint := widget.NewLabel(i18n.T(i18n.KeyConnectPickerHint))
	content := container.NewBorder(hint, buttons, nil, nil, list)
	w.SetContent(container.NewPadded(content))
	w.Show()
}

func (a *App) openServerTab(s config.Server) {
	tab := &TabSession{server: s, state: tabDisconnected, remotePath: defaultRemoteRoot(&s)}
	a.tabs = append(a.tabs, tab)
	a.activeTab = len(a.tabs) - 1
	a.tabBar.Refresh()
	a.activateTab(a.activeTab)
}
