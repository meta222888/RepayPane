package ui

import (
	"image/color"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	tabBarHeight    float32 = 24
	tabChipHeight   float32 = 22
	tabChipMinWidth float32 = 148
	tabCloseWidth   float32 = 18
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
		layout.NewCustomPaddedLayout(1, 1, 4, 4),
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
	c.bg.SetMinSize(fyne.NewSize(64, tabChipHeight))
	c.label = canvas.NewText(i18n.T(i18n.KeyNewTabConnect), colorAccent)
	c.label.TextSize = 10
	content := container.NewStack(c.bg, container.NewCenter(compactTabText(c.label)))
	return widget.NewSimpleRenderer(content)
}

func (c *tabAddChip) MinSize() fyne.Size {
	return fyne.NewSize(64, tabChipHeight)
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
	if len(name) > 14 {
		name = name[:12] + "…"
	}

	nameColor := colorForeground
	if !active {
		nameColor = colorMuted
	}
	nameT := canvas.NewText(name, nameColor)
	nameT.TextSize = 10
	hostT := canvas.NewText(tab.addr(), colorMuted)
	hostT.TextSize = 8
	serverIcon := canvas.NewText("🖥", colorAccent)
	serverIcon.TextSize = 9
	selectArea := container.NewHBox(
		dotWidget(dot, 6),
		compactTabText(serverIcon),
		compactTabText(nameT),
		compactTabText(hostT),
	)
	selectTap := newTabTapLayer(func() { t.app.activateTab(idx) })
	selectRow := container.NewStack(selectTap, selectArea)
	closeChip := newTabCloseChip(func() { t.app.closeTab(idx) })
	tabRow := container.NewBorder(nil, nil, nil, closeChip, selectRow)

	bgColor := colorTabInactive
	if active {
		bgColor = colorTabActive
	}
	bg := canvas.NewRectangle(bgColor)
	bg.SetMinSize(fyne.NewSize(tabChipMinWidth, tabChipHeight))
	stack := container.NewStack(bg, container.New(
		layout.NewCustomPaddedLayout(2, 1, 4, 2),
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
	if sz.Height < t.TextSize {
		sz.Height = t.TextSize
	}
	if sz.Width < 1 {
		sz.Width = 1
	}
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(sz)
	return container.NewStack(spacer, t)
}

type tabTapLayer struct {
	widget.BaseWidget
	tap func()
}

func newTabTapLayer(tap func()) *tabTapLayer {
	l := &tabTapLayer{tap: tap}
	l.ExtendBaseWidget(l)
	return l
}

func (l *tabTapLayer) Tapped(*fyne.PointEvent) {
	if l.tap != nil {
		l.tap()
	}
}

func (l *tabTapLayer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (l *tabTapLayer) MinSize() fyne.Size { return fyne.NewSize(0, 0) }

var _ fyne.Tappable = (*tabTapLayer)(nil)

type tabCloseChip struct {
	widget.BaseWidget
	hovered bool
	tap     func()
	bg      *canvas.Rectangle
	label   *canvas.Text
}

func newTabCloseChip(tap func()) *tabCloseChip {
	c := &tabCloseChip{tap: tap}
	c.ExtendBaseWidget(c)
	return c
}

func (c *tabCloseChip) refreshStyle() {
	if c.bg == nil {
		return
	}
	if c.hovered {
		c.bg.FillColor = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x18}
		c.label.Color = colorForeground
	} else {
		c.bg.FillColor = color.Transparent
		c.label.Color = colorMuted
	}
	canvas.Refresh(c.bg)
	canvas.Refresh(c.label)
}

func (c *tabCloseChip) Tapped(*fyne.PointEvent) {
	if c.tap != nil {
		c.tap()
	}
}

func (c *tabCloseChip) MouseIn(_ *desktop.MouseEvent) {
	c.hovered = true
	c.refreshStyle()
}

func (c *tabCloseChip) MouseMoved(_ *desktop.MouseEvent) {}

func (c *tabCloseChip) MouseOut() {
	c.hovered = false
	c.refreshStyle()
}

func (c *tabCloseChip) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (c *tabCloseChip) CreateRenderer() fyne.WidgetRenderer {
	c.bg = canvas.NewRectangle(color.Transparent)
	c.label = canvas.NewText("×", colorMuted)
	c.label.TextSize = 12
	c.label.TextStyle = fyne.TextStyle{Bold: true}
	return widget.NewSimpleRenderer(container.NewStack(c.bg, container.NewCenter(compactTabText(c.label))))
}

func (c *tabCloseChip) MinSize() fyne.Size {
	return fyne.NewSize(tabCloseWidth, tabChipHeight)
}

var _ fyne.Tappable = (*tabCloseChip)(nil)
var _ desktop.Hoverable = (*tabCloseChip)(nil)
var _ desktop.Cursorable = (*tabCloseChip)(nil)

func separatorLine() fyne.CanvasObject {
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	return line
}
