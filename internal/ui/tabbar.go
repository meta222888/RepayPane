package ui

import (
	"image/color"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	tabBarHeight   float32 = 30
	tabChipHeight  float32 = 28
	tabChipMinWidth float32 = 168
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
	scroll.SetMinSize(fyne.NewSize(0, tabBarHeight))
	tabBg := canvas.NewRectangle(colorTabInactive)
	stack := container.NewStack(tabBg, container.New(
		layout.NewCustomPaddedLayout(2, 2, 6, 6),
		scroll,
	))
	return container.NewVBox(stack, separatorLine())
}

func (t *TabBar) Refresh() {
	t.inner.Objects = nil
	for i, tab := range t.app.tabs {
		idx := i
		active := i == t.app.activeTab
		t.inner.Add(t.buildTab(idx, tab, active))
	}
	t.inner.Add(t.buildAddTab())
	t.inner.Refresh()
}

func (t *TabBar) buildAddTab() fyne.CanvasObject {
	return newTabAddChip(t.app.onNewTab)
}

type tabAddChip struct {
	widget.BaseWidget
	hovered bool
	bg      *canvas.Rectangle
	label   *canvas.Text
	tap     func()
}

func newTabAddChip(tap func()) *tabAddChip {
	c := &tabAddChip{tap: tap}
	c.ExtendBaseWidget(c)
	return c
}

func (c *tabAddChip) bgColor() color.Color {
	if c.hovered {
		return colorRowHover
	}
	return colorTabInactive
}

func (c *tabAddChip) refreshStyle() {
	if c.bg == nil {
		return
	}
	c.bg.FillColor = c.bgColor()
	canvas.Refresh(c.bg)
}

func (c *tabAddChip) Tapped(*fyne.PointEvent) {
	if c.tap != nil {
		c.tap()
	}
}

func (c *tabAddChip) MouseIn(_ *desktop.MouseEvent) {
	c.hovered = true
	c.refreshStyle()
}

func (c *tabAddChip) MouseMoved(_ *desktop.MouseEvent) {}

func (c *tabAddChip) MouseOut() {
	c.hovered = false
	c.refreshStyle()
}

func (c *tabAddChip) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (c *tabAddChip) CreateRenderer() fyne.WidgetRenderer {
	c.bg = canvas.NewRectangle(c.bgColor())
	c.bg.SetMinSize(fyne.NewSize(72, tabChipHeight))
	c.label = canvas.NewText(i18n.T(i18n.KeyNewTabConnect), colorAccent)
	c.label.TextSize = 11
	content := container.NewStack(c.bg, container.NewCenter(compactTabText(c.label)))
	return widget.NewSimpleRenderer(content)
}

func (c *tabAddChip) MinSize() fyne.Size {
	return fyne.NewSize(72, tabChipHeight)
}

var _ fyne.Tappable = (*tabAddChip)(nil)
var _ desktop.Hoverable = (*tabAddChip)(nil)
var _ desktop.Cursorable = (*tabAddChip)(nil)

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
	nameT.TextSize = 11
	hostT := canvas.NewText(tab.addr(), colorDisconnected)
	hostT.TextSize = 9
	serverIcon := canvas.NewText("🖥", colorAccent)
	serverIcon.TextSize = 10
	selectArea := container.NewHBox(
		dotWidget(dot, 7),
		compactTabText(serverIcon),
		compactTabText(nameT),
		compactTabText(hostT),
	)
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
	bg.SetMinSize(fyne.NewSize(tabChipMinWidth, tabChipHeight))
	stack := container.NewStack(bg, container.New(
		layout.NewCustomPaddedLayout(3, 3, 6, 4),
		tabRow,
	))
	if active {
		topLine := canvas.NewRectangle(colorAccent)
		topLine.SetMinSize(fyne.NewSize(0, 2))
		return container.NewBorder(topLine, nil, nil, nil, stack)
	}
	return stack
}

func compactTabText(t *canvas.Text) fyne.CanvasObject {
	sz, _ := fyne.CurrentApp().Driver().RenderedTextSize(t.Text, t.TextSize, t.TextStyle, t.FontSource)
	if sz.Height < t.TextSize+1 {
		sz.Height = t.TextSize + 1
	}
	if sz.Width < 1 {
		sz.Width = 1
	}
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(sz)
	return container.NewStack(spacer, t)
}

func separatorLine() fyne.CanvasObject {
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	return line
}
