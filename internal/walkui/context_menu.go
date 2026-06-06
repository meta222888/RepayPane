package walkui

import (
	"os"
	"path/filepath"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"github.com/lxn/walk"
)

func (a *App) prepareLocalContextMenu() {
	if a.localTV == nil {
		return
	}
	menu, _ := walk.NewMenu()
	a.populatePaneMenu(menu, true)
	a.localTV.SetContextMenu(menu)
}

func (a *App) prepareRemoteContextMenu() {
	if a.remoteTV == nil {
		return
	}
	menu, _ := walk.NewMenu()
	a.populatePaneMenu(menu, false)
	a.remoteTV.SetContextMenu(menu)
}

func (a *App) populatePaneMenu(menu *walk.Menu, local bool) {
	items := a.selectedClipItems(local)
	hasSel := len(items) > 0
	single := len(items) == 1
	clip := a.clipboard
	connected := a.activeClient() != nil

	if !hasSel {
		addMenuAction(menu, i18n.T(i18n.KeyRefresh), func() {
			if local {
				a.refreshLocal()
			} else {
				a.refreshRemote()
			}
		}, local || connected)
		addMenuAction(menu, i18n.T(i18n.KeyCtxPaste), func() {
			if local {
				a.ctxPasteLocal()
			} else {
				a.ctxPasteRemote()
			}
		}, a.canPaste(local, clip, connected))
		addMenuAction(menu, i18n.T(i18n.KeyCtxNewFolder), func() {
			if local {
				a.ctxNewFolderLocal()
			} else {
				a.ctxNewFolderRemote()
			}
		}, local || connected)
		addMenuAction(menu, i18n.T(i18n.KeyCtxNewFile), func() {
			if local {
				a.ctxNewFileLocal()
			} else {
				a.ctxNewFileRemote()
			}
		}, local || connected)
		return
	}

	addMenuAction(menu, i18n.T(i18n.KeyUpload), a.uploadSelected, local && hasSel && connected)
	addMenuAction(menu, i18n.T(i18n.KeyDownload), a.downloadSelected, !local && hasSel && connected)
	menu.Actions().Add(walk.NewSeparatorAction())
	addMenuAction(menu, i18n.T(i18n.KeyCtxCopy), func() {
		if local {
			a.ctxCopyLocal()
		} else {
			a.ctxCopyRemote()
		}
	}, hasSel)
	addMenuAction(menu, i18n.T(i18n.KeyCtxPaste), func() {
		if local {
			a.ctxPasteLocal()
		} else {
			a.ctxPasteRemote()
		}
	}, a.canPaste(local, clip, connected))
	menu.Actions().Add(walk.NewSeparatorAction())
	addMenuAction(menu, i18n.T(i18n.KeyEdit), func() {
		if local {
			a.ctxEditLocal()
		} else {
			a.ctxEditRemote()
		}
	}, single && !items[0].isDir)
	addMenuAction(menu, i18n.T(i18n.KeyRename), func() {
		if local {
			a.startInlineRename(true)
		} else {
			a.startInlineRename(false)
		}
	}, single)
	menu.Actions().Add(walk.NewSeparatorAction())
	addMenuAction(menu, i18n.T(i18n.KeyCtxNewFolder), func() {
		if local {
			a.ctxNewFolderLocal()
		} else {
			a.ctxNewFolderRemote()
		}
	}, local || connected)
	addMenuAction(menu, i18n.T(i18n.KeyCtxNewFile), func() {
		if local {
			a.ctxNewFileLocal()
		} else {
			a.ctxNewFileRemote()
		}
	}, local || connected)
	menu.Actions().Add(walk.NewSeparatorAction())
	addMenuAction(menu, i18n.T(i18n.KeyCtxDelete), func() {
		if local {
			a.ctxDeleteLocal()
		} else {
			a.ctxDeleteRemote()
		}
	}, hasSel && (local || connected))
}

func (a *App) canPaste(local bool, clip *paneClipboard, connected bool) bool {
	if clip == nil || len(clip.items) == 0 {
		return false
	}
	if clip.kind == paneLocal && local {
		return true
	}
	if clip.kind == paneRemote && !local {
		return connected
	}
	if clip.kind == paneLocal && !local {
		return connected
	}
	if clip.kind == paneRemote && local {
		return connected
	}
	return false
}

func addMenuAction(menu *walk.Menu, text string, fn func(), enabled bool) {
	action := walk.NewAction()
	action.SetText(text)
	action.SetEnabled(enabled)
	if enabled {
		action.Triggered().Attach(fn)
	}
	menu.Actions().Add(action)
}

func (a *App) ctxEditLocal() {
	items := a.selectedLocalEntries()
	if len(items) != 1 || items[0].isDir {
		return
	}
	e := items[0]
	a.openLocalEditor(e.fullPath, e.name, e.size)
}

func (a *App) ctxEditRemote() {
	items := a.selectedRemoteEntries()
	if len(items) != 1 || items[0].isDir {
		return
	}
	e := items[0]
	a.openRemoteEditor(remote.FileInfo{
		Name: e.name, Path: e.fullPath, Size: e.size, IsDir: false, ModTime: e.modTime,
	})
}

func (a *App) ctxNewFileLocal() {
	name, ok := a.promptInput(i18n.T(i18n.KeyCtxNewFile), i18n.T(i18n.KeyRenamePrompt), "new.txt")
	if !ok || name == "" {
		return
	}
	dst := filepath.Join(a.localPath, name)
	if pathExistsLocal(dst) {
		a.showMsg(i18n.T(i18n.KeyFileExistsTitle), i18n.Tf(i18n.KeyFileExists, name))
		return
	}
	if err := os.WriteFile(dst, nil, 0o644); err != nil {
		a.showError(i18n.T(i18n.KeyCtxNewFile), err)
		return
	}
	a.refreshLocal()
	a.openLocalEditor(dst, name, 0)
}

func (a *App) ctxNewFileRemote() {
	client := a.activeClient()
	if client == nil {
		return
	}
	name, ok := a.promptInput(i18n.T(i18n.KeyCtxNewFile), i18n.T(i18n.KeyRenamePrompt), "new.txt")
	if !ok || name == "" {
		return
	}
	dst := joinRemote(a.remotePath, name)
	go func() {
		if err := client.WriteFile(dst, nil); err != nil {
			a.showError(i18n.T(i18n.KeyCtxNewFile), err)
			return
		}
		a.syncUI(func() {
			a.refreshRemote()
			a.openRemoteEditor(remote.FileInfo{Name: name, Path: dst, Size: 0})
		})
	}()
}
