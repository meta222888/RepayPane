package ui

import (
	"fmt"
	"path/filepath"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
)

func (a *App) handleDrop(_ fyne.Position, uris []fyne.URI, remoteArea bool) {
	if len(uris) == 0 {
		return
	}
	client := a.activeClient()

	if remoteArea {
		if client == nil {
			dialogShow(a, i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedUpload))
			return
		}
		for _, u := range uris {
			localPath := u.Path()
			name := filepath.Base(localPath)
			remotePath := filepath.ToSlash(filepath.Join(a.remotePane.CurrentPath(), name))
			a.transfers.EnqueueUpload(client, localPath, remotePath, func(err error) {
				fyne.Do(func() {
					if err != nil {
						dialogShowError(a, fmt.Errorf("upload %s: %w", name, err))
						return
					}
					a.remotePane.RefreshListing()
				})
			})
		}
		return
	}

	for _, u := range uris {
		src := u.Path()
		name := filepath.Base(src)
		dst := filepath.Join(a.localPane.CurrentPath(), name)
		if err := copyFile(src, dst); err != nil {
			dialogShowError(a, fmt.Errorf("copy %s: %w", name, err))
		}
	}
	a.localPane.RefreshListing()
}

func (a *App) uploadSelectedLocal() {
	client := a.activeClient()
	if client == nil {
		dialogShow(a, i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedFirst))
		return
	}
	path := a.localPane.SelectedPath()
	if path == "" {
		dialogShow(a, i18n.T(i18n.KeySelectFile), i18n.T(i18n.KeySelectLocalUpload))
		return
	}
	name := filepath.Base(path)
	remotePath := filepath.ToSlash(filepath.Join(a.remotePane.CurrentPath(), name))
	a.transfers.EnqueueUpload(client, path, remotePath, func(err error) {
		fyne.Do(func() {
			if err != nil {
				dialogShowError(a, err)
				return
			}
			a.remotePane.RefreshListing()
		})
	})
}

func (a *App) downloadSelectedRemote() {
	client := a.activeClient()
	if client == nil {
		return
	}
	entry := a.remotePane.SelectedEntry()
	if entry == nil || entry.IsDir {
		dialogShow(a, i18n.T(i18n.KeySelectFile), i18n.T(i18n.KeySelectRemoteDownload))
		return
	}
	localPath := filepath.Join(a.localPane.CurrentPath(), entry.Name)
	a.transfers.EnqueueDownload(client, entry.Path, localPath, func(err error) {
		fyne.Do(func() {
			if err != nil {
				dialogShowError(a, err)
				return
			}
			a.localPane.RefreshListing()
		})
	})
}

func copyFile(src, dst string) error {
	return copyFileLocal(src, dst)
}
