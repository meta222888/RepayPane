//go:build windows

package ui

import (
	"image/color"
	"syscall"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procGetWindowRect    = user32.NewProc("GetWindowRect")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
	procSetCapture       = user32.NewProc("SetCapture")
	procReleaseCapture   = user32.NewProc("ReleaseCapture")
	procShowWindow       = user32.NewProc("ShowWindow")
	procIsZoomed         = user32.NewProc("IsZoomed")
	procGetWindowLongPtr = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProc   = user32.NewProc("CallWindowProcW")
)

const (
	swMinimize   = 6
	swMaximize   = 3
	swRestore    = 9
	swpNoSize    = 0x0001
	swpNoZOrder  = 0x0004
	gwlpWndProc  = uintptr(0xfffffffffffffffc) // GWLP_WNDPROC = -4
	wmNcHitTest  = 0x0084
	htLeft       = 10
	htRight      = 11
	htTop        = 12
	htTopLeft    = 13
	htTopRight   = 14
	htBottom     = 15
	htBottomLeft = 16
	htBottomRight = 17
	resizeBorder   = 8
	minWindowWidth = 640
	minWindowHeight = 400
	gripThickness  float32 = 8
)

type winPoint struct {
	X, Y int32
}

type winRect struct {
	Left, Top, Right, Bottom int32
}

type dragSession struct {
	hwnd    uintptr
	offsetX int32
	offsetY int32
}

type resizeSession struct {
	hwnd     uintptr
	edge     uintptr
	start    winRect
}

var (
	activeDrag          dragSession
	activeResize        resizeSession
	originalWndProc     uintptr
	resizeHookInstalled bool
	mainWndProc         = syscall.NewCallback(mainWindowProc)
)

func mainWindowProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if msg == wmNcHitTest {
		x := int32(int16(uint16(lParam & 0xffff)))
		y := int32(int16(uint16(lParam >> 16)))
		if hit := hitTestResize(hwnd, x, y); hit != 0 {
			return hit
		}
	}
	r, _, _ := procCallWindowProc.Call(originalWndProc, hwnd, msg, wParam, lParam)
	return r
}

func hitTestResize(hwnd uintptr, x, y int32) uintptr {
	zoomed, _, _ := procIsZoomed.Call(hwnd)
	if zoomed != 0 {
		return 0
	}
	var rc winRect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	atLeft := x-rc.Left < resizeBorder
	atRight := rc.Right-x <= resizeBorder
	atTop := y-rc.Top < resizeBorder
	atBottom := rc.Bottom-y <= resizeBorder
	switch {
	case atLeft && atTop:
		return htTopLeft
	case atRight && atTop:
		return htTopRight
	case atLeft && atBottom:
		return htBottomLeft
	case atRight && atBottom:
		return htBottomRight
	case atLeft:
		return htLeft
	case atRight:
		return htRight
	case atTop:
		return htTop
	case atBottom:
		return htBottom
	}
	return 0
}

func winInstallResizeHook(w fyne.Window) bool {
	if resizeHookInstalled {
		return true
	}
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return false
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 || resizeHookInstalled {
			return
		}
		prev, _, _ := procGetWindowLongPtr.Call(c.HWND, gwlpWndProc)
		if prev == 0 {
			return
		}
		if prev == mainWndProc {
			resizeHookInstalled = true
			return
		}
		ret, _, _ := procSetWindowLongPtr.Call(c.HWND, gwlpWndProc, mainWndProc)
		if ret == 0 {
			return
		}
		originalWndProc = prev
		resizeHookInstalled = true
	})
	return resizeHookInstalled
}

func wrapWindowResizePlatform(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	return winWrapResizeGrips(w, content)
}

func winWrapResizeGrips(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	if winIsMaximized(w) {
		return content
	}
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
	if e.Button != desktop.MouseButtonPrimary || winIsMaximized(g.win) {
		return
	}
	winBeginResize(g.win, g.edge)
}

func (g *resizeGrip) MouseUp(*desktop.MouseEvent)   {}
func (g *resizeGrip) MouseIn(*desktop.MouseEvent)   {}
func (g *resizeGrip) MouseOut()                     {}

var _ desktop.Mouseable = (*resizeGrip)(nil)

func winBeginResize(w fyne.Window, edge uintptr) bool {
	winInstallResizeHook(w)
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return false
	}
	var started bool
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		var rc winRect
		procGetWindowRect.Call(c.HWND, uintptr(unsafe.Pointer(&rc)))
		activeResize = resizeSession{hwnd: c.HWND, edge: edge, start: rc}
		procSetCapture.Call(c.HWND)
		started = true
		go resizeTrackLoop()
	})
	return started
}

func resizeTrackLoop() {
	ticker := time.NewTicker(time.Millisecond * 8)
	defer ticker.Stop()
	for winLeftButtonDown() {
		fyne.Do(winContinueResize)
		<-ticker.C
	}
	fyne.Do(winEndResize)
}

func winContinueResize() {
	if activeResize.hwnd == 0 {
		return
	}
	var pt winPoint
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	rc := activeResize.start
	x, y, w, h := rc.Left, rc.Top, rc.Right-rc.Left, rc.Bottom-rc.Top
	switch activeResize.edge {
	case htLeft:
		x = pt.X
		w = rc.Right - x
	case htRight:
		w = pt.X - rc.Left
	case htTop:
		y = pt.Y
		h = rc.Bottom - y
	case htBottom:
		h = pt.Y - rc.Top
	case htTopLeft:
		x = pt.X
		y = pt.Y
		w = rc.Right - x
		h = rc.Bottom - y
	case htTopRight:
		y = pt.Y
		w = pt.X - rc.Left
		h = rc.Bottom - y
	case htBottomLeft:
		x = pt.X
		w = rc.Right - x
		h = pt.Y - rc.Top
	case htBottomRight:
		w = pt.X - rc.Left
		h = pt.Y - rc.Top
	}
	if w < minWindowWidth {
		if activeResize.edge == htLeft || activeResize.edge == htTopLeft || activeResize.edge == htBottomLeft {
			x = rc.Right - minWindowWidth
		}
		w = minWindowWidth
	}
	if h < minWindowHeight {
		if activeResize.edge == htTop || activeResize.edge == htTopLeft || activeResize.edge == htTopRight {
			y = rc.Bottom - minWindowHeight
		}
		h = minWindowHeight
	}
	procSetWindowPos.Call(
		activeResize.hwnd, 0,
		uintptr(x), uintptr(y),
		uintptr(w), uintptr(h),
		0,
	)
}

func winEndResize() {
	if activeResize.hwnd != 0 {
		procReleaseCapture.Call()
		activeResize = resizeSession{}
	}
}

func winIsMaximized(w fyne.Window) bool {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return false
	}
	var maximized bool
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		r, _, _ := procIsZoomed.Call(c.HWND)
		maximized = r != 0
	})
	return maximized
}

func winToggleMaximize(w fyne.Window) {
	if winIsMaximized(w) {
		winRestoreWindows(w)
		return
	}
	winMaximizeWindows(w)
}

func winBeginDrag(d *dragRegion) bool {
	if winIsMaximized(d.win) {
		winRestoreWindows(d.win)
	}
	winInstallResizeHook(d.win)
	nw, ok := d.win.(driver.NativeWindow)
	if !ok {
		return false
	}
	var started bool
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		var pt winPoint
		var r winRect
		procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
		procGetWindowRect.Call(c.HWND, uintptr(unsafe.Pointer(&r)))
		activeDrag = dragSession{
			hwnd:    c.HWND,
			offsetX: pt.X - r.Left,
			offsetY: pt.Y - r.Top,
		}
		procSetCapture.Call(c.HWND)
		started = true
		go dragTrackLoop(d)
	})
	return started
}

func dragTrackLoop(d *dragRegion) {
	ticker := time.NewTicker(time.Millisecond * 8)
	defer ticker.Stop()
	for winLeftButtonDown() {
		fyne.Do(func() { winContinueDrag(d) })
		<-ticker.C
	}
	fyne.Do(func() {
		winEndDrag()
		d.dragging = false
	})
}

func winContinueDrag(d *dragRegion) {
	if activeDrag.hwnd == 0 {
		return
	}
	nw, ok := d.win.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		if activeDrag.hwnd == 0 {
			return
		}
		var pt winPoint
		procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
		procSetWindowPos.Call(
			activeDrag.hwnd, 0,
			uintptr(int64(pt.X)-int64(activeDrag.offsetX)),
			uintptr(int64(pt.Y)-int64(activeDrag.offsetY)),
			0, 0,
			swpNoSize|swpNoZOrder,
		)
	})
}

func winEndDrag() {
	if activeDrag.hwnd != 0 {
		procReleaseCapture.Call()
		activeDrag = dragSession{}
	}
}

func winLeftButtonDown() bool {
	const vkLButton = 0x01
	r, _, _ := user32.NewProc("GetAsyncKeyState").Call(vkLButton)
	return r&0x8000 != 0
}

func winMinimizeWindows(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		procShowWindow.Call(c.HWND, swMinimize)
	})
}

func winMaximizeWindows(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		procShowWindow.Call(c.HWND, swMaximize)
	})
}

func winRestoreWindows(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		procShowWindow.Call(c.HWND, swRestore)
	})
}
