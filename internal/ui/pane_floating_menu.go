package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var colorCtxMenuBG = color.NRGBA{R: 0x28, G: 0x2c, B: 0x36, A: 255}

type paneMenuEntry struct {
	label     string
	action    func()
	disabled  bool
	separator bool
}

// paneFloatingMenu is an in-tree context menu (not a canvas overlay).
// Avoids PopUpMenu focus steal → List.FocusLost → full list refresh → window shrink.
type paneFloatingMenu struct {
	widget.BaseWidget

	parent    fyne.CanvasObject
	visible   bool
	localPos  fyne.Position
	body      fyne.CanvasObject
	onDismiss func()
}

func newPaneFloatingMenu(parent fyne.CanvasObject) *paneFloatingMenu {
	m := &paneFloatingMenu{parent: parent}
	m.ExtendBaseWidget(m)
	return m
}

func (m *paneFloatingMenu) MinSize() fyne.Size {
	if !m.visible || m.body == nil {
		return fyne.NewSize(0, 0)
	}
	return m.body.MinSize()
}

func (m *paneFloatingMenu) CreateRenderer() fyne.WidgetRenderer {
	return &paneFloatingMenuRenderer{menu: m}
}

func (m *paneFloatingMenu) IsVisible() bool {
	return m.visible
}

func (m *paneFloatingMenu) Dismiss() {
	if !m.visible {
		return
	}
	m.visible = false
	m.body = nil
	cb := m.onDismiss
	m.onDismiss = nil
	m.Hide()
	m.Refresh()
	if cb != nil {
		cb()
	}
}

func (m *paneFloatingMenu) ShowAtCanvas(at fyne.Position, entries []paneMenuEntry, onDismiss func()) {
	m.Dismiss()

	rows := make([]fyne.CanvasObject, 0, len(entries))
	for _, e := range entries {
		if e.separator {
			rows = append(rows, widget.NewSeparator())
			continue
		}
		entry := e
		btn := widget.NewButton(entry.label, func() {
			if entry.disabled || entry.action == nil {
				return
			}
			entry.action()
			m.Dismiss()
		})
		btn.Importance = widget.LowImportance
		if entry.disabled {
			btn.Disable()
		}
		rows = append(rows, btn)
	}

	inner := container.NewVBox(rows...)
	bg := canvas.NewRectangle(colorCtxMenuBG)
	m.body = container.NewStack(bg, container.NewPadded(inner))
	m.onDismiss = onDismiss

	drv := fyne.CurrentApp().Driver()
	m.localPos = at.Subtract(drv.AbsolutePositionForObject(m.parent))

	m.visible = true
	m.Show()
	m.Refresh()
}

func (m *paneFloatingMenu) layoutIn(bounds fyne.Size) {
	if !m.visible || m.body == nil {
		m.Hide()
		return
	}
	m.Show()
	size := m.body.MinSize()
	pos := m.localPos
	if pos.X+size.Width > bounds.Width {
		pos.X = bounds.Width - size.Width
	}
	if pos.Y+size.Height > bounds.Height {
		pos.Y = bounds.Height - size.Height
	}
	if pos.X < 0 {
		pos.X = 0
	}
	if pos.Y < 0 {
		pos.Y = 0
	}
	m.Resize(size)
	m.Move(pos)
}

type paneFloatingMenuRenderer struct {
	menu *paneFloatingMenu
}

func (r *paneFloatingMenuRenderer) Destroy() {}

func (r *paneFloatingMenuRenderer) Layout(size fyne.Size) {
	if r.menu.body != nil {
		r.menu.body.Resize(size)
		r.menu.body.Move(fyne.NewPos(0, 0))
	}
}

func (r *paneFloatingMenuRenderer) MinSize() fyne.Size {
	return r.menu.MinSize()
}

func (r *paneFloatingMenuRenderer) Objects() []fyne.CanvasObject {
	if r.menu.body == nil {
		return nil
	}
	return []fyne.CanvasObject{r.menu.body}
}

func (r *paneFloatingMenuRenderer) Refresh() {
	if r.menu.body != nil {
		r.menu.body.Refresh()
	}
}
