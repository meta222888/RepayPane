package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type paneListUnderlay struct {
	widget.BaseWidget
	pane *FilePane
}

func newPaneListUnderlay(p *FilePane) *paneListUnderlay {
	b := &paneListUnderlay{pane: p}
	b.ExtendBaseWidget(b)
	return b
}

func (b *paneListUnderlay) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.NRGBA{A: 0})
	return widget.NewSimpleRenderer(bg)
}

func (b *paneListUnderlay) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (b *paneListUnderlay) Tapped(*fyne.PointEvent) {
	b.pane.noteActive()
	if c := b.pane.app.window.Canvas(); c != nil {
		c.Unfocus()
	}
	b.pane.clearSelectionQuiet()
}

func (b *paneListUnderlay) TappedSecondary(ev *fyne.PointEvent) {
	b.pane.noteActive()
	if c := b.pane.app.window.Canvas(); c != nil {
		c.Unfocus()
	}
	if b.pane.renamingRow >= 0 {
		b.pane.finishRenameBlurFromUI()
	}
	b.pane.showContextMenu(ev.AbsolutePosition, -1)
}

func (b *paneListUnderlay) Cursor() desktop.Cursor {
	if b.pane.dragReady {
		return fileDragCursor()
	}
	return desktop.DefaultCursor
}

func (b *paneListUnderlay) MouseIn(*desktop.MouseEvent)  {}
func (b *paneListUnderlay) MouseMoved(*desktop.MouseEvent) {}
func (b *paneListUnderlay) MouseOut()                      {}

var _ fyne.Tappable = (*paneListUnderlay)(nil)
var _ fyne.SecondaryTappable = (*paneListUnderlay)(nil)
var _ desktop.Cursorable = (*paneListUnderlay)(nil)
var _ desktop.Hoverable = (*paneListUnderlay)(nil)

type paneListStackLayout struct {
	pane *FilePane
}

func (l paneListStackLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	if len(objects) == 1 {
		return objects[0].MinSize()
	}
	return objects[1].MinSize()
}

func (l paneListStackLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	underlay, list := objects[0], objects[1]
	contentH := l.pane.listContentHeight()
	listH := contentH
	if listH > size.Height {
		listH = size.Height
	}
	// Shrink the list to visible rows so blank area below receives underlay taps.
	list.Resize(fyne.NewSize(size.Width, listH))
	list.Move(fyne.NewPos(0, 0))

	gap := size.Height - listH
	if gap < 1 {
		underlay.Hide()
		return
	}
	underlay.Show()
	underlay.Resize(fyne.NewSize(size.Width, gap))
	underlay.Move(fyne.NewPos(0, listH))
}

func newPaneListArea(p *FilePane, list *widget.List) fyne.CanvasObject {
	compactList := container.NewThemeOverride(list, newListCompactTheme())
	// Underlay first, list second — hit-testing prefers later children so list receives row clicks.
	return container.New(&paneListStackLayout{pane: p}, newPaneListUnderlay(p), compactList)
}

func relayoutPaneListArea(area fyne.CanvasObject) {
	if area == nil {
		return
	}
	size := area.Size()
	if size.Width > 0 && size.Height > 0 {
		area.Resize(fyne.NewSize(size.Width, size.Height+1))
		area.Resize(size)
	}
	if w, ok := area.(fyne.Widget); ok {
		w.Refresh()
	}
	canvas.Refresh(area)
}
