//go:build windows

package ui

import (
	"image/color"
	"reflect"
	"sync"
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
	user32                       = syscall.NewLazyDLL("user32.dll")
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetCurrentProcessId      = kernel32.NewProc("GetCurrentProcessId")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procShowWindow               = user32.NewProc("ShowWindow")
	procIsZoomed                 = user32.NewProc("IsZoomed")
	procIsWindow                 = user32.NewProc("IsWindow")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procGetDoubleClickTime       = user32.NewProc("GetDoubleClickTime")
)

const (
	gwlStyle        = ^uintptr(15) // GWL_STYLE (-16)
	wmNcLButtonDown = 0x00A1
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
	wsThickFrame    = 0x00040000
	wsMaximizeBox   = 0x00010000
	swpNoMove       = 0x0002
	swpNoSize       = 0x0001
	swpNoZOrder     = 0x0004
	swpFrameChanged = 0x0020
	gripThickness   float32 = 8
)

var (
	frameMainWindow    fyne.Window
	frameResizeStyled  uintptr
	hwndCache          sync.Map
)

func doubleClickInterval() time.Duration {
	ms, _, _ := procGetDoubleClickTime.Call()
	if ms == 0 {
		return 500 * time.Millisecond
	}
	return time.Duration(ms) * time.Millisecond
}

func windowKey(w fyne.Window) uintptr {
	if w == nil {
		return 0
	}
	v := reflect.ValueOf(w)
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return 0
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0
	}
	return v.Pointer()
}

func hwndOwnedByProcess(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)), 0)
	myPID, _, _ := procGetCurrentProcessId.Call()
	return uintptr(pid) == myPID
}

func winHWND(w fyne.Window) uintptr {
	if key := windowKey(w); key != 0 {
		if v, ok := hwndCache.Load(key); ok {
			hwnd := v.(uintptr)
			if ok, _, _ := procIsWindow.Call(hwnd); ok != 0 && hwndOwnedByProcess(hwnd) {
				return hwnd
			}
			hwndCache.Delete(key)
		}
	}
	hwnd := resolveHWND(w)
	if hwnd == 0 {
		return 0
	}
	if key := windowKey(w); key != 0 {
		hwndCache.Store(key, hwnd)
	}
	return hwnd
}

func resolveHWND(w fyne.Window) uintptr {
	if nw, ok := w.(driver.NativeWindow); ok {
		var hwnd uintptr
		nw.RunNative(func(ctx any) {
			c, ok := ctx.(driver.WindowsWindowContext)
			if ok {
				hwnd = c.HWND
			}
		})
		if hwnd != 0 && hwndOwnedByProcess(hwnd) {
			if ok, _, _ := procIsWindow.Call(hwnd); ok != 0 {
				return hwnd
			}
		}
	}
	if vp, ok := fyneGLFWViewport(w); ok {
		hwnd := uintptr(unsafe.Pointer(vp.GetWin32Window()))
		if hwnd != 0 && hwndOwnedByProcess(hwnd) {
			if ok, _, _ := procIsWindow.Call(hwnd); ok != 0 {
				return hwnd
			}
		}
	}
	return 0
}

func enableWindowResizeFrame(hwnd uintptr) {
	if hwnd == 0 || frameResizeStyled == hwnd {
		return
	}
	style, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlStyle)
	if style == 0 {
		return
	}
	procSetWindowLongPtrW.Call(hwnd, gwlStyle, style|wsThickFrame|wsMaximizeBox)
	procSetWindowPos.Call(
		hwnd, 0, 0, 0, 0, 0,
		swpNoMove|swpNoSize|swpNoZOrder|swpFrameChanged,
	)
	frameResizeStyled = hwnd
}

func winInstallResizeHook(w fyne.Window) bool {
	frameMainWindow = w
	hwnd := winHWND(w)
	if hwnd == 0 {
		return false
	}
	enableWindowResizeFrame(hwnd)
	return true
}

func wrapWindowResizePlatform(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	_ = winInstallResizeHook(w)
	return wrapResizeEdges(w, content)
}

func wrapResizeEdges(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	mk := func(edge uintptr) fyne.CanvasObject {
		return newResizeEdge(w, edge)
	}
	top := container.NewGridWithColumns(3, mk(htTopLeft), mk(htTop), mk(htTopRight))
	bottom := container.NewGridWithColumns(3, mk(htBottomLeft), mk(htBottom), mk(htBottomRight))
	mid := container.NewBorder(nil, nil, mk(htLeft), mk(htRight), content)
	return container.NewBorder(top, bottom, nil, nil, mid)
}

type resizeEdge struct {
	widget.BaseWidget
	win  fyne.Window
	edge uintptr
}

func newResizeEdge(w fyne.Window, edge uintptr) *resizeEdge {
	g := &resizeEdge{win: w, edge: edge}
	g.ExtendBaseWidget(g)
	return g
}

func (g *resizeEdge) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(bg)
}

func (g *resizeEdge) MinSize() fyne.Size {
	if g.edge == htTop || g.edge == htBottom {
		return fyne.NewSize(0, gripThickness)
	}
	if g.edge == htLeft || g.edge == htRight {
		return fyne.NewSize(gripThickness, 0)
	}
	return fyne.NewSize(gripThickness, gripThickness)
}

func (g *resizeEdge) MouseDown(e *desktop.MouseEvent) {
	if e.Button != desktop.MouseButtonPrimary || winIsMaximized(g.win) {
		return
	}
	winStartEdgeResize(g.win, g.edge)
}

func (g *resizeEdge) MouseUp(*desktop.MouseEvent) {}
func (g *resizeEdge) MouseIn(*desktop.MouseEvent) {}
func (g *resizeEdge) MouseOut()                   {}

var _ desktop.Mouseable = (*resizeEdge)(nil)

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

func winPostNcMouseDown(hwnd, edge uintptr) {
	procPostMessageW.Call(hwnd, wmNcLButtonDown, edge, 0)
	procSetForegroundWindow.Call(hwnd)
}

func winStartCaptionDrag(w fyne.Window) bool {
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
	winPostNcMouseDown(hwnd, htCaption)
	return true
}

func winStartEdgeResize(w fyne.Window, edge uintptr) bool {
	hwnd := winHWND(w)
	if hwnd == 0 {
		return false
	}
	winPostNcMouseDown(hwnd, edge)
	return true
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
