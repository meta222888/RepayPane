//go:build windows

package ui

import (
	"syscall"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	procGetCursorPos   = user32.NewProc("GetCursorPos")
	procGetWindowRect  = user32.NewProc("GetWindowRect")
	procSetWindowPos   = user32.NewProc("SetWindowPos")
	procSetCapture     = user32.NewProc("SetCapture")
	procReleaseCapture = user32.NewProc("ReleaseCapture")
	procShowWindow     = user32.NewProc("ShowWindow")
)

const (
	swMinimize  = 6
	swMaximize  = 3
	swRestore   = 9
	swpNoSize   = 0x0001
	swpNoZOrder = 0x0004
)

type winPoint struct {
	X, Y int32
}

type winRect struct {
	Left, Top, Right, Bottom int32
}

type dragSession struct {
	hwnd         uintptr
	offsetX      int32
	offsetY      int32
}

var activeDrag dragSession

func winBeginDrag(d *dragRegion) bool {
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
