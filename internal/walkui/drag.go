package walkui

import (
	"os"
	"path/filepath"

	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type dragState struct {
	active      bool
	sourceLocal bool
	items       []clipItem
}

func (d *dragState) reset() {
	*d = dragState{}
}

func (a *App) selectedClipItems(local bool) []clipItem {
	var out []clipItem
	if local {
		for _, e := range a.selectedLocalEntries() {
			out = append(out, clipItem{path: e.fullPath, name: e.name, isDir: e.isDir})
		}
	} else {
		for _, e := range a.selectedRemoteEntries() {
			out = append(out, clipItem{path: e.fullPath, name: e.name, isDir: e.isDir})
		}
	}
	return out
}

func (a *App) cancelFileDrag() {
	if !a.drag.active {
		return
	}
	setFileDragCursor(a, false)
	a.drag.reset()
}

func (a *App) attachPaneDrag(tv *walk.TableView, local bool) {
	if tv == nil {
		return
	}
	var down bool
	var startX, startY int
	const threshold = 10

	tv.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		down = true
		startX, startY = x, y
	})
	tv.MouseMove().Attach(func(x, y int, button walk.MouseButton) {
		if a.drag.active {
			if !isLeftButtonDown() {
				a.cancelFileDrag()
				down = false
			}
			return
		}
		if !down {
			return
		}
		dx, dy := x-startX, y-startY
		if dx*dx+dy*dy < threshold*threshold {
			return
		}
		items := a.selectedClipItems(local)
		if len(items) == 0 {
			return
		}
		a.drag.active = true
		a.drag.sourceLocal = local
		a.drag.items = items
		setFileDragCursor(a, true)
	})
	tv.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		if a.drag.active {
			setFileDragCursor(a, false)
			targetLocal, onPane := a.dragTargetPaneAtCursor()
			if onPane && a.drag.sourceLocal != targetLocal {
				a.completeDragDrop(targetLocal)
			}
		}
		down = false
		a.drag.reset()
	})
}

func (a *App) completeDragDrop(targetLocal bool) {
	if len(a.drag.items) == 0 {
		return
	}
	if a.drag.sourceLocal && !targetLocal {
		client := a.activeClient()
		if client == nil {
			a.showMsg(i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedUpload))
			return
		}
		a.clipboard = &paneClipboard{kind: paneLocal, items: append([]clipItem(nil), a.drag.items...)}
		a.pasteLocalToRemote()
		return
	}
	if !a.drag.sourceLocal && targetLocal {
		if a.activeClient() == nil {
			return
		}
		a.clipboard = &paneClipboard{kind: paneRemote, items: append([]clipItem(nil), a.drag.items...)}
		a.pasteRemoteToLocal()
	}
}

func (a *App) handleDropFiles(files []string) {
	if len(files) == 0 {
		return
	}
	remoteArea := dropCursorOnRightHalf(a.mw)
	client := a.activeClient()

	if remoteArea {
		if client == nil {
			a.showMsg(i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedUpload))
			return
		}
		for _, localPath := range files {
			name := filepath.Base(localPath)
			dst := joinRemote(a.remotePath, name)
			info, err := os.Stat(localPath)
			if err != nil {
				continue
			}
			if info.IsDir() {
				a.enqueueUploadTree(client, localPath, dst, func() { a.refreshRemote() })
				continue
			}
			lp, rp := localPath, dst
			a.transfers.EnqueueUpload(client, lp, rp, func(err error) {
				if err != nil {
					a.showError(i18n.T(i18n.KeyUpload), err)
					return
				}
				a.refreshRemote()
			})
		}
		return
	}

	for _, src := range files {
		name := filepath.Base(src)
		dst := filepath.Join(a.localPath, name)
		if err := copyPathLocal(src, dst); err != nil {
			a.showError(i18n.T(i18n.KeyLocal), err)
		}
	}
	a.refreshLocal()
}

func dropCursorOnRightHalf(mw *walk.MainWindow) bool {
	if mw == nil {
		return true
	}
	hwnd := mw.Handle()
	if hwnd == 0 {
		return true
	}
	var pt win.POINT
	if !win.GetCursorPos(&pt) {
		return true
	}
	var rect win.RECT
	if !win.GetWindowRect(hwnd, &rect) {
		return true
	}
	mid := (rect.Left + rect.Right) / 2
	return pt.X >= mid
}
