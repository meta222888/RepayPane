//go:build windows

package ui

import (
	"image/color"
	"syscall"
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
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetCurrentProcessId      = kernel32.NewProc("GetCurrentProcessId")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW          = user32.NewProc("CallWindowProcW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procIsZoomed                 = user32.NewProc("IsZoomed")
	procIsWindow                 = user32.NewProc("IsWindow")
)

const (
	gwlpWndProc     = ^uintptr(3) // GWLP_WNDPROC (-4)
	wmNcHitTest     = 0x0084
	swMinimize      = 6
	swMaximize      = 3
	swRestore       = 9
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
	resizeBorderPx  int32   = 8
	titleBarHeight  int32   = 88
	titleRightSkip  int32   = 420
	minWindowWidth  int32   = 640
	minWindowHeight int32   = 400
	glfwClassName   = "GLFW30"
)

type winRect struct {
	Left, Top, Right, Bottom int32
}

var (
	cachedMainHWND      uintptr
	originalWndProc     uintptr
	wndProcHookedHWND   uintptr
	mainWndProcCallback = syscall.NewCallback(mainWindowWndProc)
)

func mainWindowWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if msg == wmNcHitTest {
		x := int32(int16(uint16(lParam & 0xffff)))
		y := int32(int16(uint16(lParam >> 16)))
		if hit := hitTestWindow(hwnd, x, y); hit != 0 {
			return hit
		}
	}
	if originalWndProc != 0 {
		r, _, _ := procCallWindowProcW.Call(originalWndProc, hwnd, msg, wParam, lParam)
		return r
	}
	return 0
}

func hitTestWindow(hwnd uintptr, sx, sy int32) uintptr {
	zoomed, _, _ := procIsZoomed.Call(hwnd)
	if zoomed != 0 {
		return 0
	}
	var rc winRect
	procGetWindowRect := user32.NewProc("GetWindowRect")
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))

	x := sx - rc.Left
	y := sy - rc.Top
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top

	atLeft := x < resizeBorderPx
	atRight := w-x <= resizeBorderPx
	atTop := y < resizeBorderPx
	atBottom := h-y <= resizeBorderPx

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

	if y >= resizeBorderPx && y < titleBarHeight && x >= resizeBorderPx && x < w-titleRightSkip {
		return htCaption
	}
	return 0
}

func winHWND(w fyne.Window) uintptr {
	if cachedMainHWND != 0 {
		if ok, _, _ := procIsWindow.Call(cachedMainHWND); ok != 0 {
			return cachedMainHWND
		}
		cachedMainHWND = 0
	}
	for _, hwnd := range []uintptr{
		winHWNDFromGLFW(w),
		winHWNDFromNative(w),
		winHWNDByClass(),
		winHWNDByTitle(w.Title()),
		winHWNDByTitle(i18n.T(i18n.KeyAppTitle)),
	} {
		if hwnd != 0 {
			cachedMainHWND = hwnd
			winEnsureWndProcHook(hwnd)
			return hwnd
		}
	}
	return 0
}

func winHWNDFromGLFW(w fyne.Window) uintptr {
	vp, ok := fyneGLFWViewport(w)
	if !ok {
		return 0
	}
	hwnd := uintptr(unsafe.Pointer(vp.GetWin32Window()))
	if hwnd == 0 {
		return 0
	}
	if okWin, _, _ := procIsWindow.Call(hwnd); okWin == 0 {
		return 0
	}
	return hwnd
}

func winHWNDFromNative(w fyne.Window) uintptr {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return 0
	}
	var hwnd uintptr
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if ok && c.HWND != 0 {
			hwnd = c.HWND
		}
	})
	if hwnd == 0 {
		return 0
	}
	if okWin, _, _ := procIsWindow.Call(hwnd); okWin == 0 {
		return 0
	}
	return hwnd
}

func winHWNDByClass() uintptr {
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
		if !winClassIs(hwnd, glfwClassName) {
			return 1
		}
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}
		found = hwnd
		return 0
	})
	procEnumWindows.Call(cb, 0)
	return found
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

func winClassIs(hwnd uintptr, want string) bool {
	buf := make([]uint16, 64)
	n, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return false
	}
	return syscall.UTF16ToString(buf[:n]) == want
}

func winEnsureWndProcHook(hwnd uintptr) {
	if hwnd == 0 || wndProcHookedHWND == hwnd {
		return
	}
	prev, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlpWndProc)
	if prev == 0 {
		return
	}
	if prev == mainWndProcCallback {
		wndProcHookedHWND = hwnd
		return
	}
	ret, _, _ := procSetWindowLongPtrW.Call(hwnd, gwlpWndProc, mainWndProcCallback)
	if ret == 0 && prev != mainWndProcCallback {
		return
	}
	originalWndProc = prev
	wndProcHookedHWND = hwnd
}

func winInstallResizeHook(w fyne.Window) bool {
	return winHWND(w) != 0
}

func wrapWindowResizePlatform(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	_ = winHWND(w)
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
	win       fyne.Window
	edge      uintptr
	startX    int
	startY    int
	startW    int
	startH    int
	resizing  bool
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
	x, y, ok := fyneWindowPos(g.win)
	if !ok {
		return
	}
	w, h, ok := fyneWindowSize(g.win)
	if !ok {
		return
	}
	g.startX, g.startY, g.startW, g.startH = x, y, w, h
	g.resizing = true
}

func (g *resizeGrip) MouseUp(*desktop.MouseEvent) {
	g.resizing = false
}

func (g *resizeGrip) Dragged(e *fyne.DragEvent) {
	if !g.resizing || winIsMaximized(g.win) {
		return
	}
	dx := int(e.Dragged.DX)
	dy := int(e.Dragged.DY)
	x, y, w, h := g.startX, g.startY, g.startW, g.startH
	switch g.edge {
	case htLeft:
		x += dx
		w -= dx
	case htRight:
		w += dx
	case htTop:
		y += dy
		h -= dy
	case htBottom:
		h += dy
	case htTopLeft:
		x += dx
		y += dy
		w -= dx
		h -= dy
	case htTopRight:
		y += dy
		w += dx
		h -= dy
	case htBottomLeft:
		x += dx
		w -= dx
		h += dy
	case htBottomRight:
		w += dx
		h += dy
	}
	fyneSetWindowBounds(g.win, x, y, w, h)
}

func (g *resizeGrip) DragEnd() { g.resizing = false }

func (g *resizeGrip) MouseIn(*desktop.MouseEvent) {}
func (g *resizeGrip) MouseOut()                   {}

var _ desktop.Mouseable = (*resizeGrip)(nil)
var _ fyne.Draggable = (*resizeGrip)(nil)

func winIsMaximized(w fyne.Window) bool {
	hwnd := winHWND(w)
	if hwnd == 0 {
		return false
	}
	r, _, _ := procIsZoomed.Call(hwnd)
	return r != 0
}

func winToggleMaximize(w fyne.Window) {
	fyne.Do(func() {
		if winIsMaximized(w) {
			winRestoreWindows(w)
			return
		}
		winMaximizeWindows(w)
	})
}

func winBeginDrag(w fyne.Window) bool {
	if winIsMaximized(w) {
		winRestoreWindows(w)
	}
	_, ok := fyneGLFWViewport(w)
	return ok
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
