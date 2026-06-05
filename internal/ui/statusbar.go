package ui

import (
	"fmt"
	"image/color"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	slimProgressWidth  float32 = 72
	slimProgressHeight float32 = 5
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
	barH := slimProgressHeight
	if size.Height < barH {
		barH = size.Height
	}
	y := (size.Height - barH) / 2
	radius := barH / 2
	r.bg.CornerRadius = radius
	r.fill.CornerRadius = radius

	r.bg.Resize(fyne.NewSize(size.Width, barH))
	r.bg.Move(fyne.NewPos(0, y))

	fillW := size.Width * float32(r.bar.value)
	if r.bar.value > 0 && fillW < barH {
		fillW = barH
	}
	r.fill.Resize(fyne.NewSize(fillW, barH))
	r.fill.Move(fyne.NewPos(0, y))
	if fillW <= 0 {
		r.fill.Hide()
	} else {
		r.fill.Show()
	}
}

func (r *slimProgressRenderer) MinSize() fyne.Size {
	return fyne.NewSize(slimProgressWidth, slimProgressHeight)
}

func (r *slimProgressRenderer) Refresh() {
	r.bg.FillColor = colorProgressTrack
	r.fill.FillColor = colorAccent
	size := r.bar.Size()
	if size.Width <= 0 || size.Height <= 0 {
		size = r.MinSize()
	}
	r.Layout(size)
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
	app          *App
	connDot      *canvas.Circle
	connText     *canvas.Text
	reconnectBtn *widget.Button
	syncText     *canvas.Text
	speedText    *canvas.Text
	progress     *slimProgressBar
	percentText  *canvas.Text
	sep          *canvas.Rectangle
	queueText    *canvas.Text
	root         fyne.CanvasObject
}

func newStatusText(text string, muted bool) *canvas.Text {
	c := colorForeground
	if muted {
		c = colorMuted
	}
	t := canvas.NewText(text, c)
	t.TextSize = AppTextSize
	return t
}

func bandSlimProgress(p *slimProgressBar) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, AppTextSize))
	return container.NewStack(spacer, container.NewCenter(p))
}

func NewStatusBar(app *App) *StatusBar {
	s := &StatusBar{app: app}
	s.connDot = canvas.NewCircle(colorDisconnected)
	s.connText = newStatusText(i18n.T(i18n.KeyDisconnected), false)
	s.reconnectBtn = widget.NewButton(i18n.T(i18n.KeyReconnect), func() {
		s.app.reconnectActiveTab()
	})
	s.reconnectBtn.Importance = widget.LowImportance
	s.reconnectBtn.Hide()
	s.syncText = newStatusText("", false)
	s.syncText.Hide()
	s.speedText = newStatusText(i18n.T(i18n.KeyTransferIdle), true)
	s.progress = newSlimProgressBar()
	s.percentText = newStatusText("0%", true)
	s.sep = canvas.NewRectangle(colorBorder)
	s.sep.SetMinSize(fyne.NewSize(1, 8))
	s.queueText = newStatusText(i18n.Tf(i18n.KeyStatusQueue, 0), true)

	left := container.NewHBox(
		dotWidget(s.connDot, 6),
		bandCanvasText(s.connText),
		bandCanvasText(s.syncText),
		wrapCompactToolbar(s.reconnectBtn),
	)
	right := container.NewHBox(
		bandCanvasText(s.speedText),
		container.New(layout.NewCustomPaddedLayout(0, 0, 6, 6), bandSlimProgress(s.progress)),
		bandCanvasText(s.percentText),
		s.sep,
		bandCanvasText(s.queueText),
	)
	inner := container.NewBorder(nil, nil, left, right, nil)
	s.root = withStatusBar(inner)
	return s
}

func (s *StatusBar) Container() fyne.CanvasObject { return s.root }

func (s *StatusBar) Refresh() {
	sess := s.app.activeSession()
	if sess == nil || sess.state != tabConnected {
		s.connDot.FillColor = colorDisconnected
		s.connText.Text = i18n.T(i18n.KeyDisconnected)
		s.connText.Color = colorForeground
		if sess != nil && sess.state == tabConnecting {
			s.reconnectBtn.Hide()
		} else {
			s.reconnectBtn.Show()
		}
	} else {
		s.connDot.FillColor = colorConnected
		s.connText.Text = i18n.T(i18n.KeyStatusConnected) + " " + sess.addr()
		s.connText.Color = colorForeground
		s.reconnectBtn.Hide()
	}
	canvas.Refresh(s.connDot)
	canvas.Refresh(s.connText)
	s.RefreshTransfer()
}

func (s *StatusBar) RefreshTransfer() {
	active, pct, speed, queue := s.app.transfers.Snapshot()
	s.speedText.Text = speed
	s.progress.SetValue(pct / 100)
	s.percentText.Text = fmt.Sprintf("%.0f%%", pct)
	s.queueText.Text = i18n.Tf(i18n.KeyStatusQueue, queue)
	canvas.Refresh(s.speedText)
	canvas.Refresh(s.percentText)
	canvas.Refresh(s.queueText)
	if active {
		s.syncText.Text = "  ⟳ " + i18n.T(i18n.KeyStatusSyncing)
		s.syncText.Show()
		canvas.Refresh(s.syncText)
	} else {
		s.syncText.Hide()
	}
}

func (s *StatusBar) ApplyLanguage() {
	s.reconnectBtn.SetText(i18n.T(i18n.KeyReconnect))
	s.Refresh()
}
