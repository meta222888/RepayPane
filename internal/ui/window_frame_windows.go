//go:build windows

package ui

import (
	"image/color"
	"syscall"
	"time"
	"unsafe"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procFindWindowW              = user32.NewProc("FindWindowW")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetCurrentProcessId      = kernel32.NewProc("GetCurrentProcessId")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procSetCapture               = user32.NewProc("SetCapture")
	procReleaseCapture           = user32.NewProc("ReleaseCapture")
	procGetAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	procShowWindow               = user32.NewProc("ShowWindow")
	procIsZoomed                 = user32.NewProc("IsZoomed")
	procIsWindow                 = user32.NewProc("IsWindow")
)

const (
	swMinimize      = 6
	swMaximize      = 3
	swRestore       = 9
	swpNoSize       = 0x0001
	swpNoZOrder     = 0x0004
	vkLButton       = 0x01
	htLeft          = 10
	htRight         = 11
	htTop           = 12
	htTopLeft       = 13
	htTopRight      = 14
	htBottom        = 15
	htBottomLeft    = 16
	htBottomRight   = 17
	gripThickness   float32 = 8
	minWindowWidth  int32   = 640
	minWindowHeight int32   = 400
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
	hwnd  uintptr
	edge  uintptr
	start winRect
}

var (
	activeDrag     dragSession
	activeResize   resizeSession
	cachedMainHWND uintptr
)

func winHWND(w fyne.Window) uintptr {
	if cachedMainHWND != 0 {
		if ok, _, _ := procIsWindow.Call(cachedMainHWND); ok != 0 {
			return cachedMainHWND
		}
		cachedMainHWND = 0
	}
	for _, hwnd := range []uintptr{
		winHWNDFromNative(w),
		winHWNDByTitle(w.Title()),
		winHWNDByTitle(i18n.T(i18n.KeyAppTitle)),
		winHWNDForProcess(w.Title()),
		winHWNDForProcess(i18n.T(i18n.KeyAppTitle)),
	} {
		if hwnd != 0 {
			cachedMainHWND = hwnd
			return hwnd
		}
	}
	return 0
}

func winHWNDFromNative(w fyne.Window) uintptr {
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
	if ok, _, _ := procIsWindow.Call(hwnd); ok == 0 {
		return 0
	}
	return hwnd
}

func winHWNDByTitle(title string) uintptr {
	if title == "" {
		return 0
	}
	ptr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return 0
	}
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(ptr)))
	if hwnd == 0 {
		return 0
	}
	if ok, _, _ := procIsWindow.Call(hwnd); ok == 0 {
		return 0
	}
	return hwnd
}

func winHWNDForProcess(title string) uintptr {
	pid, _, _ := procGetCurrentProcessId.Call()
	var found uintptr
	cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		if found != 0 {
			return 1
		}
		var winPID uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&winPID)), 0)
		if uintptr(winPID) != pid {
			return 1
		}
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}
		if title != "" && !winTitleMatches(hwnd, title) {
			return 1
		}
		if ok, _, _ := procIsWindow.Call(hwnd); ok == 0 {
			return 1
		}
		found = hwnd
		return 0
	})
	procEnumWindows.Call(cb, 0)
	return found
}

func winTitleMatches(hwnd uintptr, want string) bool {
	buf := make([]uint16, 256)
	n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return false
	}
	return syscall.UTF16ToString(buf[:n]) == want
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
	winBeginResize(g.win, g.edge)
}

func (g *resizeGrip) MouseUp(*desktop.MouseEvent)   {}
func (g *resizeGrip) MouseIn(*desktop.MouseEvent)   {}
func (g *resizeGrip) MouseOut()                     {}

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

func winBeginDrag(w fyne.Window) bool {
	hwnd := winHWND(w)
	if hwnd == 0 {
		return false
	}
	if winIsMaximized(w) {
		winRestoreWindows(w)
		hwnd = winHWND(w)
		if hwnd == 0 {
			return false
		}
	}
	var pt winPoint
	var rc winRect
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	activeDrag = dragSession{
		hwnd:    hwnd,
		offsetX: pt.X - rc.Left,
		offsetY: pt.Y - rc.Top,
	}
	procSetCapture.Call(hwnd)
	go dragTrackLoop()
	return true
}

func dragTrackLoop() {
	ticker := time.NewTicker(8 * time.Millisecond)
	defer ticker.Stop()
	for winLeftButtonDown() {
		winContinueDrag()
		<-ticker.C
	}
	winEndDrag()
}

func winContinueDrag() {
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
}

func winEndDrag() {
	if activeDrag.hwnd != 0 {
		procReleaseCapture.Call()
		activeDrag = dragSession{}
	}
}

func winBeginResize(w fyne.Window, edge uintptr) bool {
	if winIsMaximized(w) {
		return false
	}
	hwnd := winHWND(w)
	if hwnd == 0 {
		return false
	}
	var rc winRect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	activeResize = resizeSession{hwnd: hwnd, edge: edge, start: rc}
	procSetCapture.Call(hwnd)
	go resizeTrackLoop()
	return true
}

func resizeTrackLoop() {
	ticker := time.NewTicker(8 * time.Millisecond)
	defer ticker.Stop()
	for winLeftButtonDown() {
		winContinueResize()
		<-ticker.C
	}
	winEndResize()
}

func winContinueResize() {
	if activeResize.hwnd == 0 {
		return
	}
	var pt winPoint
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	rc := activeResize.start
	x, y := rc.Left, rc.Top
	w, h := rc.Right-rc.Left, rc.Bottom-rc.Top
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
		x, y = pt.X, pt.Y
		w, h = rc.Right-x, rc.Bottom-y
	case htTopRight:
		y = pt.Y
		w, h = pt.X-rc.Left, rc.Bottom-y
	case htBottomLeft:
		x = pt.X
		w, h = rc.Right-x, pt.Y-rc.Top
	case htBottomRight:
		w, h = pt.X-rc.Left, pt.Y-rc.Top
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

func winLeftButtonDown() bool {
	r, _, _ := procGetAsyncKeyState.Call(vkLButton)
	return r&0x8000 != 0
}

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
