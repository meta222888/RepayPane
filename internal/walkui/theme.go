package walkui

import (
	"github.com/lxn/walk"
)

var (
	colorConnected    = walk.RGB(46, 204, 113)
	colorDisconnected = walk.RGB(231, 76, 60)
	colorConnecting   = walk.RGB(241, 196, 15)
)

func (a *App) updateConnDot() {
	if a.connDot == nil {
		return
	}
	tab := a.activeSession()
	switch {
	case tab == nil:
		a.connDot.SetText("●")
		a.connDot.SetTextColor(colorDisconnected)
	case tab.state == tabConnecting:
		a.connDot.SetText("●")
		a.connDot.SetTextColor(colorConnecting)
	case tab.state == tabConnected:
		a.connDot.SetText("●")
		a.connDot.SetTextColor(colorConnected)
	default:
		a.connDot.SetText("●")
		a.connDot.SetTextColor(colorDisconnected)
	}
}
