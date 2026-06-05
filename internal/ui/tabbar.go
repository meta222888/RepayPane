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
	scroll.SetMinSize(fyne.NewSize(0, 38))
	dragLayer := newDragRegion(t.app.window, layout.NewSpacer())
	tabBg := canvas.NewRectangle(colorTabInactive)
	stack := container.NewStack(tabBg, container.NewStack(dragLayer, container.NewPadded(scroll)))
	return container.NewVBox(stack, separatorLine())
}

func (t *TabBar) Refresh() {
	t.inner.Objects = nil
	for i, tab := range t.app.tabs {
		idx := i
		active := i == t.app.activeTab
		t.inner.Add(t.buildTab(idx, tab, active))
	}
	addBtn := widget.NewButton("+ "+i18n.T(i18n.KeyNewTabConnect), t.app.onNewTab)
	addBtn.Importance = widget.LowImportance
	t.inner.Add(addBtn)
	t.inner.Refresh()
}

func (t *TabBar) buildTab(idx int, tab *TabSession, active bool) fyne.CanvasObject {
	dotColor := colorDisconnected
	switch tab.state {
	case tabConnected:
		dotColor = colorConnected
	case tabConnecting:
		dotColor = colorWarning
	}
	dot := canvas.NewCircle(dotColor)

	name := tab.server.Name
	if name == "" {
		name = tab.server.Host
	}
	if len(name) > 16 {
		name = name[:14] + "…"
	}

	nameColor := colorForeground
	if !active {
		nameColor = colorMuted
	}
	nameT := canvas.NewText(name, nameColor)
	nameT.TextSize = 12

	hostT := canvas.NewText(tab.addr(), colorDisconnected)
	hostT.TextSize = 10

	serverIcon := canvas.NewText("🖥", colorAccent)
	serverIcon.TextSize = 11

	selectArea := container.NewHBox(dotWidget(dot, 8), serverIcon, nameT, hostT)
	selectBtn := widget.NewButton("", func() { t.app.activateTab(idx) })
	selectBtn.Importance = widget.LowImportance

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
	bg.SetMinSize(fyne.NewSize(180, 34))
	stack := container.NewStack(bg, container.NewPadded(tabRow))
	if active {
		topLine := canvas.NewRectangle(colorAccent)
		topLine.SetMinSize(fyne.NewSize(0, 2))
		return container.NewBorder(topLine, nil, nil, nil, stack)
	}
	return stack
}

func separatorLine() fyne.CanvasObject {
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	return line
}
