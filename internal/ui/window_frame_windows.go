//go:build windows

package ui

import (
	"reflect"
	"sync"
	"syscall"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetCurrentProcessId      = kernel32.NewProc("GetCurrentProcessId")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW          = user32.NewProc("CallWindowProcW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procIsZoomed                 = user32.NewProc("IsZoomed")
	procIsWindow                 = user32.NewProc("IsWindow")
	procReleaseCapture           = user32.NewProc("ReleaseCapture")
	procPostMessageW             = user32.NewProc("PostMessageW")
)

const (
	gwlpWndProc    = ^uintptr(3) // GWLP_WNDPROC (-4)
	wmNcHitTest    = 0x0084
	wmNcLButtonDown = 0x00A1
	swMinimize     = 6
	swMaximize     = 3
	swRestore      = 9
	htCaption      = 2
	htLeft         = 10
	htRight        = 11
	htTop          = 12
	htTopLeft      = 13
	htTopRight     = 14
	htBottom       = 15
	htBottomLeft   = 16
	htBottomRight  = 17
	edgeBorderPx   = 8
)

type winRect struct {
	Left, Top, Right, Bottom int32
}

var (
	frameMainWindow     fyne.Window
	frameHookedHWND     uintptr
	frameOriginalProc   uintptr
	frameWndProcCB      = syscall.NewCallback(frameWndProc)
	hwndCache           sync.Map
)

func frameWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if msg == wmNcHitTest {
		x := int32(int16(uint16(lParam & 0xffff)))
		y := int32(int16(uint16(lParam >> 16)))
		if hit := hitTestResizeEdges(hwnd, x, y); hit != 0 {
			return hit
		}
	}
	if frameOriginalProc != 0 {
		r, _, _ := procCallWindowProcW.Call(frameOriginalProc, hwnd, msg, wParam, lParam)
		return r
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return r
}

func scaledEdgeBorder() int32 {
	if frameMainWindow == nil {
		return edgeBorderPx
	}
	c := frameMainWindow.Canvas()
	if c == nil {
		return edgeBorderPx
	}
	s := c.Scale()
	if s <= 0 {
		return edgeBorderPx
	}
	return int32(float32(edgeBorderPx) * s)
}

func hitTestResizeEdges(hwnd uintptr, sx, sy int32) uintptr {
	if hwnd != frameHookedHWND {
		return 0
	}
	zoomed, _, _ := procIsZoomed.Call(hwnd)
	if zoomed != 0 {
		return 0
	}
	var rc winRect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))

	border := scaledEdgeBorder()
	x := sx - rc.Left
	y := sy - rc.Top
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top

	atLeft := x < border
	atRight := w-x <= border
	atTop := y < border
	atBottom := h-y <= border

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

func ensureFrameWndProcHook(hwnd uintptr) {
	if hwnd == 0 || frameHookedHWND == hwnd {
		return
	}
	prev, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlpWndProc)
	if prev == 0 {
		return
	}
	if prev == frameWndProcCB {
		frameHookedHWND = hwnd
		return
	}
	ret, _, _ := procSetWindowLongPtrW.Call(hwnd, gwlpWndProc, frameWndProcCB)
	if ret == 0 && prev != frameWndProcCB {
		return
	}
	frameOriginalProc = prev
	frameHookedHWND = hwnd
}

func winInstallResizeHook(w fyne.Window) bool {
	frameMainWindow = w
	hwnd := winHWND(w)
	if hwnd == 0 {
		return false
	}
	ensureFrameWndProcHook(hwnd)
	return frameHookedHWND != 0
}

func wrapWindowResizePlatform(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	_ = winInstallResizeHook(w)
	return content
}

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
	procReleaseCapture.Call()
	procPostMessageW.Call(hwnd, wmNcLButtonDown, htCaption, 0)
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
