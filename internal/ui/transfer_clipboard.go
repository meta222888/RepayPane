package ui

import (
	"os"
	"path/filepath"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
)

func (a *App) paneAtPosition(pos fyne.Position) *FilePane {
	for _, pane := range []*FilePane{a.localPane, a.remotePane} {
		if pane == nil || pane.root == nil {
			continue
		}
		origin := a.fyneApp.Driver().AbsolutePositionForObject(pane.root)
		rel := pos.Subtract(origin)
		size := pane.root.Size()
		if rel.X >= 0 && rel.Y >= 0 && rel.X <= size.Width && rel.Y <= size.Height {
			return pane
		}
	}
	return nil
}

func (a *App) completePaneDrop(source *FilePane, absPos fyne.Position) {
	clip := a.clipboard
	if clip == nil || len(clip.Items) == 0 {
		return
	}
	target := a.paneAtPosition(absPos)
	if target == nil || target == source {
		return
	}
	// Drag-copy only between local (left) and remote (right) panes.
	if clip.Kind == target.kind {
		return
	}
	a.transferClipboardToPane(clip, target)
}

func (a *App) transferClipboardToPane(clip *PaneClipboard, dest *FilePane) {
	if clip == nil || dest == nil || len(clip.Items) == 0 {
		return
	}
	if clip.Kind == PaneLocal && dest.kind == PaneRemote {
		a.uploadClipboardToRemote(clip, dest.CurrentPath(), func() {
			fyne.Do(func() { dest.RefreshListing() })
		})
		return
	}
	if clip.Kind == PaneRemote && dest.kind == PaneLocal {
		a.downloadClipboardToLocal(clip, dest.CurrentPath(), func() {
			fyne.Do(func() { dest.RefreshListing() })
		})
	}
}

func (a *App) uploadClipboardToRemote(clip *PaneClipboard, remoteDir string, onDone func()) {
	client := a.activeClient()
	if client == nil {
		dialogShow(a, i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedUpload))
		return
	}
	a.uploadClipboardItems(clip.Items, remoteDir, 0, onDone)
}

func (a *App) uploadClipboardItems(items []PaneClipItem, remoteDir string, idx int, onDone func()) {
	if idx >= len(items) {
		if onDone != nil {
			onDone()
		}
		return
	}
	item := items[idx]
	client := a.activeClient()
	if client == nil {
		return
	}
	exists := func(name string) bool {
		dst := filepath.ToSlash(filepath.Join(remoteDir, name))
		return destExistsRemote(client, dst)
	}
	start := func(name string) {
		if name == "" {
			a.uploadClipboardItems(items, remoteDir, idx+1, onDone)
			return
		}
		a.doUploadClipboardItem(item, remoteDir, name, func() {
			a.uploadClipboardItems(items, remoteDir, idx+1, onDone)
		})
	}
	if exists(item.Name) {
		a.resolveFileConflict(item.Name, exists, start)
		return
	}
	start(item.Name)
}

func (a *App) doUploadClipboardItem(item PaneClipItem, remoteDir, destName string, onDone func()) {
	client := a.activeClient()
	if client == nil {
		return
	}
	dst := filepath.ToSlash(filepath.Join(remoteDir, destName))
	if item.IsDir {
		a.enqueueUploadTree(client, item.Path, dst, onDone)
		return
	}
	a.transfers.EnqueueUpload(client, item.Path, dst, func(err error) {
		fyne.Do(func() {
			if err != nil {
				dialogShowError(a, err)
				return
			}
			if onDone != nil {
				onDone()
			}
		})
	})
}

func (a *App) downloadClipboardToLocal(clip *PaneClipboard, localDir string, onDone func()) {
	client := a.activeClient()
	if client == nil {
		dialogShow(a, i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedFirst))
		return
	}
	a.downloadClipboardItems(clip.Items, localDir, 0, onDone)
}

func (a *App) downloadClipboardItems(items []PaneClipItem, localDir string, idx int, onDone func()) {
	if idx >= len(items) {
		if onDone != nil {
			onDone()
		}
		return
	}
	item := items[idx]
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(localDir, name))
		return err == nil
	}
	start := func(name string) {
		if name == "" {
			a.downloadClipboardItems(items, localDir, idx+1, onDone)
			return
		}
		a.doDownloadClipboardItem(item, localDir, name, func() {
			a.downloadClipboardItems(items, localDir, idx+1, onDone)
		})
	}
	if exists(item.Name) {
		a.resolveFileConflict(item.Name, exists, start)
		return
	}
	start(item.Name)
}

func (a *App) doDownloadClipboardItem(item PaneClipItem, localDir, destName string, onDone func()) {
	client := a.activeClient()
	if client == nil {
		return
	}
	dst := filepath.Join(localDir, destName)
	if item.IsDir {
		a.enqueueDownloadTree(client, item.Path, dst, onDone)
		return
	}
	a.transfers.EnqueueDownload(client, item.Path, dst, func(err error) {
		fyne.Do(func() {
			if err != nil {
				dialogShowError(a, err)
				return
			}
			if onDone != nil {
				onDone()
			}
		})
	})
}

func (a *App) enqueueUploadTree(client *remote.Client, localPath, remotePath string, onDone func()) {
	var pending int
	enqueue := func(localFile, remoteFile string) {
		pending++
		a.transfers.EnqueueUpload(client, localFile, remoteFile, func(err error) {
			fyne.Do(func() {
				if err != nil {
					dialogShowError(a, err)
				}
				pending--
				if pending == 0 && onDone != nil {
					onDone()
				}
			})
		})
	}
	_ = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(localPath, path)
		if err != nil {
			return nil
		}
		enqueue(path, filepath.ToSlash(filepath.Join(remotePath, rel)))
		return nil
	})
	if pending == 0 && onDone != nil {
		onDone()
	}
}

func (a *App) enqueueDownloadTree(client *remote.Client, remotePath, localPath string, onDone func()) {
	var pending int
	var walk func(dir string, localBase string)
	walk = func(dir, localBase string) {
		entries, err := client.ListDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir {
				walk(e.Path, filepath.Join(localBase, e.Name))
				continue
			}
			pending++
			dst := filepath.Join(localBase, e.Name)
			a.transfers.EnqueueDownload(client, e.Path, dst, func(err error) {
				fyne.Do(func() {
					if err != nil {
						dialogShowError(a, err)
					}
					pending--
					if pending == 0 && onDone != nil {
						onDone()
					}
				})
			})
		}
	}
	walk(remotePath, localPath)
	if pending == 0 && onDone != nil {
		onDone()
	}
}

func destExistsRemote(client *remote.Client, path string) bool {
	_, err := client.Stat(path)
	return err == nil
}
