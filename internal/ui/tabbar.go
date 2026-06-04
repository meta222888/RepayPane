package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TabBar struct {
	app   *App
	inner *fyne.Container
}

func NewTabBar(app *App) *TabBar {
	return &TabBar{app: app, inner: container.NewHBox()}
}

func (t *TabBar) Container() fyne.CanvasObject {
	scroll := container.NewHScroll(t.inner)
	scroll.SetMinSize(fyne.NewSize(0, 34))
	return scroll
}

func (t *TabBar) Refresh() {
	t.inner.Objects = nil
	for i, tab := range t.app.tabs {
		idx := i
		prefix := "○ "
		if tab.state == tabConnected {
			prefix = "● "
		}
		label := prefix + tab.tabLabel()
		btn := widget.NewButton(label, func() { t.app.activateTab(idx) })
		btn.Importance = widget.LowImportance
		if i == t.app.activeTab {
			btn.Importance = widget.MediumImportance
		}
		closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
			t.app.closeTab(idx)
		})
		closeBtn.Importance = widget.LowImportance
		row := container.NewBorder(nil, nil, nil, closeBtn, btn)
		t.inner.Add(row)
	}
	addBtn := widget.NewButtonWithIcon(i18n.T(i18n.KeyNewTab), theme.ContentAddIcon(), t.app.onNewTab)
	addBtn.Importance = widget.LowImportance
	t.inner.Add(addBtn)
	t.inner.Refresh()
}
