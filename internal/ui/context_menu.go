package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ctxMenuState struct {
	popup     *widget.PopUp
	onDismiss func()
}

var activeCtxMenu *ctxMenuState

// ctxMenuLayer catches clicks outside the menu. Secondary tap dismisses and reopens at the new point.
type ctxMenuLayer struct {
	widget.BaseWidget
	onPrimary   func()
	onSecondary func(*fyne.PointEvent)
}

func newCtxMenuLayer(onPrimary func(), onSecondary func(*fyne.PointEvent)) *ctxMenuLayer {
	l := &ctxMenuLayer{onPrimary: onPrimary, onSecondary: onSecondary}
	l.ExtendBaseWidget(l)
	return l
}

func (l *ctxMenuLayer) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.NRGBA{A: 0})
	return widget.NewSimpleRenderer(bg)
}

func (l *ctxMenuLayer) Tapped(*fyne.PointEvent) {
	if l.onPrimary != nil {
		l.onPrimary()
	}
}

func (l *ctxMenuLayer) TappedSecondary(ev *fyne.PointEvent) {
	if l.onSecondary != nil {
		l.onSecondary(ev)
	}
}

var _ fyne.Tappable = (*ctxMenuLayer)(nil)
var _ fyne.SecondaryTappable = (*ctxMenuLayer)(nil)

type ctxMenuOverlayLayout struct {
	menuPos fyne.Position
}

func (l ctxMenuOverlayLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	return objects[0].MinSize()
}

func (l ctxMenuOverlayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	layer, menu := objects[0], objects[1]
	layer.Resize(size)
	layer.Move(fyne.NewPos(0, 0))
	ms := menu.MinSize()
	menu.Resize(ms)
	menu.Move(l.menuPos)
}

func hideActiveContextMenu() {
	if activeCtxMenu == nil {
		return
	}
	st := activeCtxMenu
	activeCtxMenu = nil
	st.popup.Hide()
	if st.onDismiss != nil {
		st.onDismiss()
	}
}

func dismissPopUpMenus(c fyne.Canvas) {
	hideActiveContextMenu()
	for _, o := range c.Overlays().List() {
		o.Hide()
	}
}

func adjustMenuPosition(at fyne.Position, menuSize fyne.Size, c fyne.Canvas) fyne.Position {
	_, areaSize := c.InteractiveArea()
	x, y := at.X, at.Y
	if x+menuSize.Width > areaSize.Width {
		x = areaSize.Width - menuSize.Width
		if x < 0 {
			x = 0
		}
	}
	if y+menuSize.Height > areaSize.Height {
		y = areaSize.Height - menuSize.Height
		if y < 0 {
			y = 0
		}
	}
	return fyne.NewPos(x, y)
}

func showPopUpContextMenu(w fyne.Window, at fyne.Position, menu *fyne.Menu, onDismiss func(), onRetap func(fyne.Position)) {
	c := w.Canvas()
	dismissPopUpMenus(c)

	menuW := widget.NewMenu(menu)
	menuW.OnDismiss = func() { hideActiveContextMenu() }
	menuSize := menuW.MinSize()
	menuPos := adjustMenuPosition(at, menuSize, c)

	layer := newCtxMenuLayer(
		func() { hideActiveContextMenu() },
		func(ev *fyne.PointEvent) {
			hideActiveContextMenu()
			if onRetap != nil {
				onRetap(ev.AbsolutePosition)
			}
		},
	)

	content := container.New(&ctxMenuOverlayLayout{menuPos: menuPos}, layer, menuW)
	pop := widget.NewPopUp(content, c)
	pop.Resize(c.Size())

	activeCtxMenu = &ctxMenuState{popup: pop, onDismiss: onDismiss}
	pop.Show()
}

func (a *App) routeContextMenu(at fyne.Position) {
	if a.localPane.openContextMenuIfHit(at) {
		return
	}
	a.remotePane.openContextMenuIfHit(at)
}

func (p *FilePane) openContextMenuIfHit(at fyne.Position) bool {
	if p.listArea == nil {
		return false
	}
	drv := fyne.CurrentApp().Driver()
	areaPos := drv.AbsolutePositionForObject(p.listArea)
	rel := at.Subtract(areaPos)
	size := p.listArea.Size()
	if rel.X < 0 || rel.Y < 0 || rel.X >= size.Width || rel.Y >= size.Height {
		return false
	}
	if rel.Y >= p.listContentHeight() {
		p.showContextMenu(at, -1)
		return true
	}
	row := p.rowAtPosition(rel.Y)
	if row < 0 || row >= p.rowCount() {
		return false
	}
	p.showContextMenu(at, row)
	return true
}

func (p *FilePane) rowAtPosition(relY float32) int {
	if p.list == nil {
		return -1
	}
	pad := fyne.CurrentApp().Settings().Theme().Size(theme.SizeNamePadding)
	rowH := paneRowMinHeight + pad
	y := relY + p.list.GetScrollOffset()
	row := int(y / rowH)
	if row < 0 {
		return 0
	}
	return row
}
