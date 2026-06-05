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
	tabBarHeight    float32 = 20
	tabChipHeight   float32 = 20
	tabChipMinWidth float32 = 148
	tabCloseWidth   float32 = 18
	tabAccentLineH  float32 = 2
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
		layout.NewCustomPaddedLayout(4, 0, 4, 0),
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
	hostT := canvas.NewText(tab.tabAddrShort(), colorMuted)
	hostT.TextSize = 9
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
	if active {
		return newTabActiveChip(bgColor, tabRow)
	}
	return newTabInactiveChip(bgColor, tabRow)
}

type tabInactiveChip struct {
	widget.BaseWidget
	bgColor color.Color
	content fyne.CanvasObject
}

func newTabInactiveChip(bgColor color.Color, content fyne.CanvasObject) *tabInactiveChip {
	c := &tabInactiveChip{bgColor: bgColor, content: content}
	c.ExtendBaseWidget(c)
	return c
}

func (c *tabInactiveChip) MinSize() fyne.Size {
	return fyne.NewSize(tabChipMinWidth, tabChipHeight)
}

type tabInactiveChipRenderer struct {
	chip *tabInactiveChip
	bg   *canvas.Rectangle
	body fyne.CanvasObject
}

func (r *tabInactiveChipRenderer) Layout(size fyne.Size) {
	size.Height = tabChipHeight
	r.chip.Resize(size)
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))

	padH := float32(4)
	innerW := size.Width - padH*2
	if innerW < 0 {
		innerW = 0
	}
	r.body.Resize(fyne.NewSize(innerW, size.Height))
	contentH := r.body.MinSize().Height
	if contentH > size.Height {
		contentH = size.Height
	}
	y := (size.Height - contentH) / 2
	if y < 0 {
		y = 0
	}
	r.body.Move(fyne.NewPos(padH, y))
}

func (r *tabInactiveChipRenderer) MinSize() fyne.Size {
	return r.chip.MinSize()
}

func (r *tabInactiveChipRenderer) Refresh() {
	r.bg.FillColor = r.chip.bgColor
	canvas.Refresh(r.bg)
}

func (r *tabInactiveChipRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.body}
}

func (r *tabInactiveChipRenderer) Destroy() {}

func (c *tabInactiveChip) CreateRenderer() fyne.WidgetRenderer {
	return &tabInactiveChipRenderer{
		chip: c,
		bg:   canvas.NewRectangle(c.bgColor),
		body: c.content,
	}
}

type tabActiveChip struct {
	widget.BaseWidget
	bgColor color.Color
	content fyne.CanvasObject
}

func newTabActiveChip(bgColor color.Color, content fyne.CanvasObject) *tabActiveChip {
	c := &tabActiveChip{bgColor: bgColor, content: content}
	c.ExtendBaseWidget(c)
	return c
}

func (c *tabActiveChip) MinSize() fyne.Size {
	return fyne.NewSize(tabChipMinWidth, tabChipHeight)
}

type tabActiveChipRenderer struct {
	chip    *tabActiveChip
	bg      *canvas.Rectangle
	line    *canvas.Rectangle
	padded  fyne.CanvasObject
}

func (r *tabActiveChipRenderer) Layout(size fyne.Size) {
	size.Height = tabChipHeight
	r.chip.Resize(size)

	r.line.Resize(fyne.NewSize(size.Width, tabAccentLineH))
	r.line.Move(fyne.NewPos(0, 0))

	bodyH := size.Height - tabAccentLineH
	r.bg.Resize(fyne.NewSize(size.Width, bodyH))
	r.bg.Move(fyne.NewPos(0, tabAccentLineH))

	padH := float32(4)
	innerW := size.Width - padH*2
	innerH := bodyH
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}
	r.padded.Resize(fyne.NewSize(innerW, innerH))
	contentH := r.padded.MinSize().Height
	if contentH > innerH {
		contentH = innerH
	}
	y := tabAccentLineH + (bodyH-contentH)/2
	if y < tabAccentLineH {
		y = tabAccentLineH
	}
	r.padded.Move(fyne.NewPos(padH, y))
}

func (r *tabActiveChipRenderer) MinSize() fyne.Size {
	return r.chip.MinSize()
}

func (r *tabActiveChipRenderer) Refresh() {
	r.bg.FillColor = r.chip.bgColor
	canvas.Refresh(r.bg)
	canvas.Refresh(r.line)
}

func (r *tabActiveChipRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.line, r.padded}
}

func (r *tabActiveChipRenderer) Destroy() {}

func (c *tabActiveChip) CreateRenderer() fyne.WidgetRenderer {
	return &tabActiveChipRenderer{
		chip:   c,
		bg:     canvas.NewRectangle(c.bgColor),
		line:   canvas.NewRectangle(colorAccent),
		padded: c.content,
	}
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
