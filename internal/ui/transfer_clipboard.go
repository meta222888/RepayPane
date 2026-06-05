package ui

import (
	"fmt"
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
	if clip == nil {
		return
	}
	target := a.paneAtPosition(absPos)
	if target == nil {
		return
	}
	a.transferClipboardToPane(clip, target)
}

func (a *App) transferClipboardToPane(clip *PaneClipboard, dest *FilePane) {
	if clip == nil || dest == nil {
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
		return
	}
	if clip.Kind == dest.kind {
		dest.pasteClipboard(clip)
	}
}

func (a *App) uploadClipboardToRemote(clip *PaneClipboard, remoteDir string, onDone func()) {
	client := a.activeClient()
	if client == nil {
		dialogShow(a, i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedUpload))
		return
	}
	dst := filepath.ToSlash(filepath.Join(remoteDir, clip.Name))
	if destExistsRemote(client, dst) {
		dialogShowError(a, fmt.Errorf(i18n.Tf(i18n.KeyFileExists, clip.Name)))
		return
	}
	if clip.IsDir {
		a.enqueueUploadTree(client, clip.Path, dst, onDone)
		return
	}
	a.transfers.EnqueueUpload(client, clip.Path, dst, func(err error) {
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
	dst := filepath.Join(localDir, clip.Name)
	if _, err := os.Stat(dst); err == nil {
		dialogShowError(a, fmt.Errorf(i18n.Tf(i18n.KeyFileExists, clip.Name)))
		return
	}
	if clip.IsDir {
		a.enqueueDownloadTree(client, clip.Path, dst, onDone)
		return
	}
	a.transfers.EnqueueDownload(client, clip.Path, dst, func(err error) {
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
