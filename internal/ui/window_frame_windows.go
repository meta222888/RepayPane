//go:build windows

package ui

import (
	"image/color"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	procReleaseCapture = user32.NewProc("ReleaseCapture")
	procSendMessage    = user32.NewProc("SendMessageW")
	procShowWindow     = user32.NewProc("ShowWindow")
	procIsZoomed       = user32.NewProc("IsZoomed")
	procIsWindow       = user32.NewProc("IsWindow")
)

const (
	swMinimize      = 6
	swMaximize      = 3
	swRestore       = 9
	wmNcLButtonDown = 0x00A1
	htCaption       = 2
	htLeft          = 10
	htRight         = 11
	htTop           = 12
	htTopLeft       = 13
	htTopRight      = 14
	htBottom        = 15
	htBottomLeft    = 16
	htBottomRight   = 17
	gripThickness   float32 = 8
)

func winHWND(w fyne.Window) uintptr {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return 0
	}
	var hwnd uintptr
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		hwnd = c.HWND
	})
	if hwnd == 0 {
		return 0
	}
	okWin, _, _ := procIsWindow.Call(hwnd)
	if okWin == 0 {
		return 0
	}
	return hwnd
}

func winSendNcLButtonDown(w fyne.Window, hit uintptr) bool {
	if hit != htCaption && winIsMaximized(w) {
		return false
	}
	hwnd := winHWND(w)
	if hwnd == 0 {
		return false
	}
	procReleaseCapture.Call()
	procSendMessage.Call(hwnd, wmNcLButtonDown, hit, 0)
	return true
}

func winInstallResizeHook(w fyne.Window) bool {
	return winHWND(w) != 0
}

func wrapWindowResizePlatform(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	return winWrapResizeGrips(w, content)
}

func winWrapResizeGrips(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	mk := func(edge uintptr) fyne.CanvasObject {
		return newResizeGrip(w, edge)
	}
	top := container.NewGridWithColumns(3, mk(htTopLeft), mk(htTop), mk(htTopRight))
	bottom := container.NewGridWithColumns(3, mk(htBottomLeft), mk(htBottom), mk(htBottomRight))
	mid := container.NewBorder(nil, nil, mk(htLeft), mk(htRight), content)
	return container.NewBorder(top, bottom, nil, nil, mid)
}

type resizeGrip struct {
	widget.BaseWidget
	win  fyne.Window
	edge uintptr
}

func newResizeGrip(w fyne.Window, edge uintptr) *resizeGrip {
	g := &resizeGrip{win: w, edge: edge}
	g.ExtendBaseWidget(g)
	return g
}

func (g *resizeGrip) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(bg)
}

func (g *resizeGrip) MinSize() fyne.Size {
	if g.edge == htTop || g.edge == htBottom {
		return fyne.NewSize(0, gripThickness)
	}
	if g.edge == htLeft || g.edge == htRight {
		return fyne.NewSize(gripThickness, 0)
	}
	return fyne.NewSize(gripThickness, gripThickness)
}

func (g *resizeGrip) MouseDown(e *desktop.MouseEvent) {
	if e.Button != desktop.MouseButtonPrimary {
		return
	}
	winSendNcLButtonDown(g.win, g.edge)
}

func (g *resizeGrip) MouseUp(*desktop.MouseEvent) {}
func (g *resizeGrip) MouseIn(*desktop.MouseEvent) {}
func (g *resizeGrip) MouseOut()                   {}

var _ desktop.Mouseable = (*resizeGrip)(nil)

func winIsMaximized(w fyne.Window) bool {
	hwnd := winHWND(w)
	if hwnd == 0 {
		return false
	}
	r, _, _ := procIsZoomed.Call(hwnd)
	return r != 0
}

func winToggleMaximize(w fyne.Window) {
	if winIsMaximized(w) {
		winRestoreWindows(w)
		return
	}
	winMaximizeWindows(w)
}

func winBeginDrag(d *dragRegion) bool {
	return winSendNcLButtonDown(d.win, htCaption)
}

func winEndDrag() {}

func winLeftButtonDown() bool { return false }

func winMinimizeWindows(w fyne.Window) {
	hwnd := winHWND(w)
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swMinimize)
}

func winMaximizeWindows(w fyne.Window) {
	hwnd := winHWND(w)
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swMaximize)
}

func winRestoreWindows(w fyne.Window) {
	hwnd := winHWND(w)
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swRestore)
}
