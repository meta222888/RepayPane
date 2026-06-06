package walkui

import (
	"image"
	"sync"
	"syscall"

	"github.com/relaypane/relaypane/internal/assets"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

const vkLButton = 0x01

var getAsyncKeyState = syscall.NewLazyDLL("user32.dll").NewProc("GetAsyncKeyState")

var (
	fileDragWalkCur     walk.Cursor
	fileDragWalkCurOnce sync.Once
)

func fileDragWalkCursor() walk.Cursor {
	fileDragWalkCurOnce.Do(func() {
		img, hotX, hotY, err := assets.DecodeCursorImageForTest(assets.CopyCURBytes())
		if err != nil {
			return
		}
		cur, err := walk.NewCursorFromImage(img, image.Pt(hotX, hotY))
		if err != nil {
			return
		}
		fileDragWalkCur = cur
	})
	return fileDragWalkCur
}

func setFileDragCursor(a *App, active bool) {
	var cur walk.Cursor
	if active {
		cur = fileDragWalkCursor()
	}
	if a == nil {
		return
	}
	if a.mw != nil {
		a.mw.SetCursor(cur)
	}
	if a.localTV != nil {
		a.localTV.SetCursor(cur)
	}
	if a.remoteTV != nil {
		a.remoteTV.SetCursor(cur)
	}
}

func isLeftButtonDown() bool {
	r, _, _ := getAsyncKeyState.Call(vkLButton)
	return r&0x8000 != 0
}

func hwndContainsCursor(hwnd win.HWND) bool {
	if hwnd == 0 {
		return false
	}
	var pt win.POINT
	if !win.GetCursorPos(&pt) {
		return false
	}
	var rect win.RECT
	if !win.GetWindowRect(hwnd, &rect) {
		return false
	}
	return pt.X >= rect.Left && pt.X < rect.Right && pt.Y >= rect.Top && pt.Y < rect.Bottom
}

func (a *App) dragTargetPaneAtCursor() (local bool, ok bool) {
	if a.remoteTV != nil && hwndContainsCursor(a.remoteTV.Handle()) {
		return false, true
	}
	if a.localTV != nil && hwndContainsCursor(a.localTV.Handle()) {
		return true, true
	}
	return false, false
}

func setWindowOwner(child, owner win.HWND) {
	if child == 0 || owner == 0 {
		return
	}
	win.SetWindowLongPtr(child, win.GWL_HWNDPARENT, uintptr(owner))
}
