package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type StatusBar struct {
	app      *App
	conn     *widget.Label
	speed    *widget.Label
	progress *widget.ProgressBar
	queue    *widget.Label
	root     *fyne.Container
}

func NewStatusBar(app *App) *StatusBar {
	s := &StatusBar{app: app}
	s.conn = widget.NewLabel(i18n.T(i18n.KeyNotConnected))
	s.speed = widget.NewLabel(i18n.T(i18n.KeyTransferIdle))
	s.progress = widget.NewProgressBar()
	s.progress.SetValue(0)
	s.progress.Hide()
	s.queue = widget.NewLabel(i18n.Tf(i18n.KeyQueue, 0))
	s.queue.Hide()

	right := container.NewHBox(s.speed, s.progress, s.queue)
	s.root = container.NewBorder(nil, nil, s.conn, right, nil)
	return s
}

func (s *StatusBar) Container() fyne.CanvasObject { return s.root }

func (s *StatusBar) Refresh() {
	sess := s.app.activeSession()
	if sess == nil || sess.state != tabConnected {
		s.conn.SetText(i18n.T(i18n.KeyNotConnected))
		s.speed.SetText(i18n.T(i18n.KeyTransferIdle))
		s.progress.Hide()
		s.queue.Hide()
		return
	}
	s.conn.SetText(i18n.Tf(i18n.KeyStatusBarConnected, sess.addr()))
	s.speed.SetText(i18n.T(i18n.KeyTransferIdle))
	s.progress.Hide()
	s.queue.Hide()
}

func (s *StatusBar) ApplyLanguage() {
	s.Refresh()
}
