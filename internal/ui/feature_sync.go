package ui

import (
	"os"
	"path/filepath"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) syncLocalToRemote() {
	a.confirmSync(true)
}

func (a *App) syncRemoteToLocal() {
	a.confirmSync(false)
}

func (a *App) confirmSync(upload bool) {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	localPath := a.localPane.CurrentPath()
	remotePath := a.remotePane.CurrentPath()

	var msg string
	if upload {
		msg = i18n.Tf(i18n.KeyFeatSyncConfirmUp, localPath, remotePath)
	} else {
		msg = i18n.Tf(i18n.KeyFeatSyncConfirmDown, remotePath, localPath)
	}
	lbl := widget.NewLabel(msg)
	lbl.Wrapping = fyne.TextWrapWord

	title := i18n.T(i18n.KeyFeatSync)
	var dlg *modalDialog
	okBtn := newAccentButton(i18n.T(i18n.KeyOK), func() {
		dlg.Close()
		if upload {
			a.runSyncUpload(client, localPath, remotePath)
		} else {
			a.runSyncDownload(client, remotePath, localPath)
		}
	})
	cancelBtn := newAccentButton(i18n.T(i18n.KeyCancel), func() { dlg.Close() })
	body := container.NewBorder(nil, container.NewHBox(cancelBtn, okBtn), nil, nil, lbl)
	dlg = newModalDialog(a, title, fyne.NewSize(520, 240), body)
}

func (a *App) runSyncUpload(client *remote.Client, localDir, remoteDir string) {
	go func() {
		_ = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(localDir, path)
			if err != nil {
				return nil
			}
			remotePath := filepath.ToSlash(filepath.Join(remoteDir, rel))
			a.transfers.EnqueueUpload(client, path, remotePath, func(err error) {
				if err != nil {
					fyne.Do(func() { dialogShowError(a, err) })
				}
			})
			return nil
		})
		fyne.Do(func() { a.remotePane.RefreshListing() })
	}()
}

func (a *App) runSyncDownload(client *remote.Client, remoteDir, localDir string) {
	go func() {
		walkRemote(client, remoteDir, func(p string, isDir bool) {
			if isDir {
				return
			}
			rel, err := filepath.Rel(remoteDir, filepath.FromSlash(p))
			if err != nil {
				return
			}
			localPath := filepath.Join(localDir, rel)
			a.transfers.EnqueueDownload(client, p, localPath, func(err error) {
				if err != nil {
					fyne.Do(func() { dialogShowError(a, err) })
				}
			})
		})
		fyne.Do(func() { a.localPane.RefreshListing() })
	}()
}

func walkRemote(client *remote.Client, dir string, fn func(path string, isDir bool)) {
	entries, err := client.ListDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		fn(e.Path, e.IsDir)
		if e.IsDir {
			walkRemote(client, e.Path, fn)
		}
	}
}
