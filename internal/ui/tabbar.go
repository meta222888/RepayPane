package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
	scroll.SetMinSize(fyne.NewSize(0, 36))
	dragLayer := newDragRegion(t.app.window, layout.NewSpacer())
	return container.NewStack(dragLayer, withBorderBottom(scroll))
}

func (t *TabBar) Refresh() {
	t.inner.Objects = nil
	for i, tab := range t.app.tabs {
		idx := i
		active := i == t.app.activeTab
		t.inner.Add(t.buildTab(idx, tab, active))
	}
	addBtn := widget.NewButton(i18n.T(i18n.KeyNewTabConnect), t.app.onNewTab)
	addBtn.Importance = widget.LowImportance
	t.inner.Add(addBtn)
	t.inner.Refresh()
}

func (t *TabBar) buildTab(idx int, tab *TabSession, active bool) fyne.CanvasObject {
	dotColor := colorDisconnected
	if tab.state == tabConnected {
		dotColor = colorConnected
	}
	dot := canvas.NewText("●", dotColor)
	dot.TextSize = 10

	name := tab.server.Name
	if name == "" {
		name = tab.server.Host
	}
	if len(name) > 20 {
		name = name[:18] + "…"
	}

	nameLbl := widget.NewLabel(name)

	selectArea := container.NewHBox(dot, widget.NewLabel("⬡"), nameLbl)
	selectBtn := widget.NewButton("", func() { t.app.activateTab(idx) })
	selectBtn.Importance = widget.LowImportance
	if active {
		selectBtn.Importance = widget.MediumImportance
	}

	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		t.app.closeTab(idx)
	})
	closeBtn.Importance = widget.LowImportance

	tabRow := container.NewBorder(nil, nil, selectArea, closeBtn, selectBtn)

	bgColor := colorTabInactive
	if active {
		bgColor = colorTabActive
	}
	bg := canvas.NewRectangle(bgColor)
	stack := container.NewStack(bg, tabRow)
	if active {
		accent := canvas.NewRectangle(colorAccent)
		accent.SetMinSize(fyne.NewSize(0, 2))
		return container.NewBorder(accent, nil, nil, nil, stack)
	}
	return stack
}
