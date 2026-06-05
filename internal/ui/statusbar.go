package ui

import (
	"fmt"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type slimProgressBar struct {
	widget.BaseWidget
	value float64 // 0..1
}

func newSlimProgressBar() *slimProgressBar {
	p := &slimProgressBar{}
	p.ExtendBaseWidget(p)
	return p
}

func (p *slimProgressBar) SetValue(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.value = v
	p.Refresh()
}

type slimProgressRenderer struct {
	bar  *slimProgressBar
	bg   *canvas.Rectangle
	fill *canvas.Rectangle
}

func (r *slimProgressRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	fillW := size.Width * float32(r.bar.value)
	r.fill.Resize(fyne.NewSize(fillW, size.Height))
	r.fill.Move(fyne.NewPos(0, 0))
}

func (r *slimProgressRenderer) MinSize() fyne.Size {
	return fyne.NewSize(160, 6)
}

func (r *slimProgressRenderer) Refresh() {
	r.bg.FillColor = colorInput
	r.fill.FillColor = colorAccent
	canvas.Refresh(r.bg)
	canvas.Refresh(r.fill)
}

func (r *slimProgressRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.fill}
}

func (r *slimProgressRenderer) Destroy() {}

func (p *slimProgressBar) CreateRenderer() fyne.WidgetRenderer {
	r := &slimProgressRenderer{bar: p}
	r.bg = canvas.NewRectangle(colorInput)
	r.fill = canvas.NewRectangle(colorAccent)
	return r
}

type StatusBar struct {
	app      *App
	connDot  *canvas.Circle
	conn     *widget.Label
	syncLabel  *widget.Label
	speed      *widget.Label
	progress   *slimProgressBar
	percent    *widget.Label
	sep        *canvas.Rectangle
	queue      *widget.Label
	root       fyne.CanvasObject
}

func NewStatusBar(app *App) *StatusBar {
	s := &StatusBar{app: app}
	s.connDot = canvas.NewCircle(colorDisconnected)
	s.conn = widget.NewLabel(i18n.T(i18n.KeyDisconnected))
	s.syncLabel = widget.NewLabel("")
	s.syncLabel.Hide()
	s.speed = widget.NewLabel(i18n.T(i18n.KeyTransferIdle))
	s.progress = newSlimProgressBar()
	s.percent = widget.NewLabel("0%")
	s.sep = canvas.NewRectangle(colorBorder)
	s.sep.SetMinSize(fyne.NewSize(1, 12))
	s.queue = widget.NewLabel(i18n.Tf(i18n.KeyStatusQueue, 0))

	left := container.NewHBox(dotWidget(s.connDot, 8), s.conn, s.syncLabel)
	right := container.NewHBox(s.speed, s.progress, s.percent, s.sep, s.queue)
	inner := container.NewBorder(nil, nil, left, right, nil)
	s.root = withStatusBar(inner)
	return s
}

func (s *StatusBar) Container() fyne.CanvasObject { return s.root }

func (s *StatusBar) Refresh() {
	sess := s.app.activeSession()
	if sess == nil || sess.state != tabConnected {
		s.connDot.FillColor = colorDisconnected
		s.conn.SetText(i18n.T(i18n.KeyDisconnected))
	} else {
		s.connDot.FillColor = colorConnected
		s.conn.SetText(i18n.T(i18n.KeyStatusConnected) + " " + sess.addr())
	}
	canvas.Refresh(s.connDot)
	s.RefreshTransfer()
}

func (s *StatusBar) RefreshTransfer() {
	active, pct, speed, queue := s.app.transfers.Snapshot()
	s.speed.SetText(speed)
	s.progress.SetValue(pct / 100)
	s.percent.SetText(fmt.Sprintf("%.0f%%", pct))
	s.queue.SetText(i18n.Tf(i18n.KeyStatusQueue, queue))
	if active {
		s.syncLabel.SetText("  ⟳ " + i18n.T(i18n.KeyStatusSyncing))
		s.syncLabel.Show()
	} else {
		s.syncLabel.Hide()
	}
}

func (s *StatusBar) ApplyLanguage() {
	s.Refresh()
}
