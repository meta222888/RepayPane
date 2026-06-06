package walkui

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
)

const paneDoubleClickInterval = 450 * time.Millisecond

type paneClickState struct {
	row int
	at  time.Time
	gen int
}

func (a *App) trackSlowDoubleClick(tv *walk.TableView, local bool, x, y int) {
	row := tv.IndexAt(x, y)
	if row < 0 {
		return
	}
	st := &a.localClick
	if !local {
		st = &a.remoteClick
	}
	now := time.Now()
	elapsed := now.Sub(st.at)
	if row == st.row && elapsed > paneDoubleClickInterval {
		sel := tv.SelectedIndexes()
		if len(sel) == 1 && sel[0] == row {
			a.schedulePendingRename(local, row)
			st.at = now
			return
		}
	}
	a.cancelPendingRename(local)
	st.row = row
	st.at = now
}

func (a *App) cancelPendingRename(local bool) {
	if local {
		a.localClick.gen++
	} else {
		a.remoteClick.gen++
	}
}

func (a *App) schedulePendingRename(local bool, row int) {
	st := &a.localClick
	if !local {
		st = &a.remoteClick
	}
	st.gen++
	gen := st.gen
	time.AfterFunc(paneDoubleClickInterval, func() {
		a.syncUI(func() {
			cur := &a.localClick
			if !local {
				cur = &a.remoteClick
			}
			if cur.gen != gen {
				return
			}
			if a.renamingLocal && local || a.renamingRemote && !local {
				return
			}
			tv, _ := a.paneTable(local)
			if tv == nil {
				return
			}
			sel := tv.SelectedIndexes()
			if len(sel) != 1 || sel[0] != row {
				return
			}
			a.beginInlineRename(local, row)
		})
	})
}

func (a *App) startInlineRename(local bool) {
	tv, model := a.paneTable(local)
	if tv == nil || model == nil {
		return
	}
	sel := tv.SelectedIndexes()
	if len(sel) != 1 {
		return
	}
	a.beginInlineRename(local, sel[0])
}

func (a *App) beginInlineRename(local bool, row int) {
	_, model := a.paneTable(local)
	e, ok := model.entry(row)
	if !ok {
		return
	}
	edit, panel := a.renameWidgets(local)
	if edit == nil || panel == nil {
		return
	}
	if local {
		a.renamingLocal = true
		a.localRenameRow = row
	} else {
		a.renamingRemote = true
		a.remoteRenameRow = row
	}
	panel.SetVisible(true)
	edit.SetText(e.name)
	edit.SetFocus()
	edit.SetTextSelection(0, len(e.name))
}

func (a *App) renameWidgets(local bool) (*walk.LineEdit, *walk.Composite) {
	if local {
		return a.localRenameEdit, a.localRenamePanel
	}
	return a.remoteRenameEdit, a.remoteRenamePanel
}

func (a *App) cancelInlineRename(local bool) {
	edit, panel := a.renameWidgets(local)
	if edit == nil || panel == nil {
		return
	}
	panel.SetVisible(false)
	if local {
		a.renamingLocal = false
		a.localRenameRow = -1
	} else {
		a.renamingRemote = false
		a.remoteRenameRow = -1
	}
}

func (a *App) commitInlineRename(local bool) {
	edit, _ := a.renameWidgets(local)
	if edit == nil {
		return
	}
	row := a.localRenameRow
	if !local {
		row = a.remoteRenameRow
	}
	_, model := a.paneTable(local)
	e, ok := model.entry(row)
	if !ok {
		a.cancelInlineRename(local)
		return
	}
	newName := strings.TrimSpace(edit.Text())
	if newName == "" || newName == e.name {
		a.cancelInlineRename(local)
		return
	}
	if !validRenameName(newName) {
		a.showError(i18n.T(i18n.KeyRename), errors.New(i18n.T(i18n.KeyRenameInvalidName)))
		return
	}
	a.cancelInlineRename(local)
	if local {
		a.commitLocalRename(e, newName)
	} else {
		a.commitRemoteRename(e, newName)
	}
}

func validRenameName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return false
	}
	return !strings.ContainsAny(name, `/\:*?"<>|`)
}

func (a *App) commitLocalRename(e dirEntry, newName string) {
	dst := filepath.Join(filepath.Dir(e.fullPath), newName)
	if pathExistsLocal(dst) {
		a.showMsg(i18n.T(i18n.KeyFileExistsTitle), i18n.Tf(i18n.KeyFileExists, newName))
		return
	}
	if err := os.Rename(e.fullPath, dst); err != nil {
		a.showError(i18n.T(i18n.KeyRename), err)
		return
	}
	a.refreshLocal()
}

func (a *App) commitRemoteRename(e dirEntry, newName string) {
	client := a.activeClient()
	if client == nil {
		return
	}
	dst := joinRemote(path.Dir(strings.ReplaceAll(e.fullPath, "\\", "/")), newName)
	go func() {
		err := client.Rename(e.fullPath, dst)
		a.syncUI(func() {
			if err != nil {
				a.showError(i18n.T(i18n.KeyRename), err)
				return
			}
			a.refreshRemote()
		})
	}()
}

func (a *App) handleRenameKey(key walk.Key) bool {
	if key == walk.KeyEscape {
		if a.renamingLocal {
			a.cancelInlineRename(true)
			return true
		}
		if a.renamingRemote {
			a.cancelInlineRename(false)
			return true
		}
	}
	if key == walk.KeyReturn {
		if a.renamingLocal {
			a.commitInlineRename(true)
			return true
		}
		if a.renamingRemote {
			a.commitInlineRename(false)
			return true
		}
	}
	return a.renamingLocal || a.renamingRemote
}

func (a *App) setupRenameEdit(edit *walk.LineEdit, local bool) {
	if edit == nil {
		return
	}
	edit.KeyDown().Attach(func(key walk.Key) {
		switch key {
		case walk.KeyReturn:
			a.commitInlineRename(local)
		case walk.KeyEscape:
			a.cancelInlineRename(local)
		}
	})
}
