package walkui

import (
	"github.com/lxn/walk"
)

func (a *App) registerShortcuts() {
	if a.mw == nil {
		return
	}
	a.attachPaneKeys(a.localTV, true)
	a.attachPaneKeys(a.remoteTV, false)
	a.registerGlobalShortcut(walk.ModControl, walk.KeyE, a.showRemoteShell)
	a.mw.KeyDown().Attach(func(key walk.Key) {
		if a.handleRenameKey(key) {
			return
		}
		mods := walk.ModifiersDown()
		switch {
		case key == walk.KeyA && mods&walk.ModControl != 0:
			a.shortcutSelectAll()
		case key == walk.KeyF5:
			a.shortcutRefresh()
		case key == walk.KeyF2:
			a.shortcutRename()
		case key == walk.KeyDelete:
			a.shortcutDelete()
		case key == walk.KeyE && mods&walk.ModControl != 0:
			a.showRemoteShell()
		}
	})
}

func (a *App) registerGlobalShortcut(mods walk.Modifiers, key walk.Key, fn func()) {
	if a.mw == nil || fn == nil {
		return
	}
	action := walk.NewAction()
	action.SetShortcut(walk.Shortcut{Modifiers: mods, Key: key})
	action.SetVisible(true)
	action.SetEnabled(true)
	action.Triggered().Attach(fn)
	a.mw.Menu().Actions().Add(action)
}

func (a *App) attachPaneKeys(tv *walk.TableView, local bool) {
	if tv == nil {
		return
	}
	tv.KeyDown().Attach(func(key walk.Key) {
		if a.handleRenameKey(key) {
			return
		}
		mods := walk.ModifiersDown()
		switch {
		case key == walk.KeyA && mods&walk.ModControl != 0:
			a.setFocusPane(local)
			a.selectAllInPane(local)
		case key == walk.KeyF5:
			a.refreshLocal()
			a.refreshRemote()
		case key == walk.KeyF2:
			a.setFocusPane(local)
			a.startInlineRename(local)
		case key == walk.KeyDelete:
			a.setFocusPane(local)
			if local {
				a.ctxDeleteLocal()
			} else {
				a.ctxDeleteRemote()
			}
		case key == walk.KeyE && mods&walk.ModControl != 0:
			a.showRemoteShell()
		}
	})
	tv.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			a.setFocusPane(local)
			a.trackSlowDoubleClick(tv, local, x, y)
		}
	})
}

func (a *App) shortcutSelectAll() {
	if a.focusLocal {
		a.selectAllInPane(true)
	} else {
		a.selectAllInPane(false)
	}
}

func (a *App) shortcutRefresh() {
	a.refreshLocal()
	a.refreshRemote()
}

func (a *App) shortcutRename() {
	if a.focusLocal {
		a.startInlineRename(true)
	} else {
		a.startInlineRename(false)
	}
}

func (a *App) shortcutDelete() {
	if a.focusLocal {
		a.ctxDeleteLocal()
	} else {
		a.ctxDeleteRemote()
	}
}

func (a *App) setFocusPane(local bool) {
	a.focusLocal = local
}

func (a *App) selectAllInPane(local bool) {
	tv, model := a.paneTable(local)
	if tv == nil || model == nil {
		return
	}
	n := model.RowCount()
	if n == 0 {
		return
	}
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	_ = tv.SetSelectedIndexes(idx)
}

func (a *App) paneTable(local bool) (*walk.TableView, *dirModel) {
	if local {
		return a.localTV, a.localModel
	}
	return a.remoteTV, a.remoteModel
}
