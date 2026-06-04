package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

func (a *App) handleDrop(_ fyne.Position, uris []fyne.URI, remoteArea bool) {
	if len(uris) == 0 {
		return
	}

	if remoteArea {
		if a.client == nil {
			dialogShow(a, "Not connected", "Connect to a server before uploading.")
			return
		}
		for _, u := range uris {
			localPath := u.Path()
			name := filepath.Base(localPath)
			remotePath := filepath.ToSlash(filepath.Join(a.remotePane.CurrentPath(), name))
			if err := a.client.Upload(localPath, remotePath); err != nil {
				dialogShowError(a, fmt.Errorf("upload %s: %w", name, err))
			}
		}
		a.remotePane.RefreshListing()
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
	if a.client == nil {
		dialogShow(a, "Not connected", "Connect to a server first.")
		return
	}
	path := a.localPane.SelectedPath()
	if path == "" {
		dialogShow(a, "Select a file", "Select a local file to upload.")
		return
	}
	name := filepath.Base(path)
	remotePath := filepath.ToSlash(filepath.Join(a.remotePane.CurrentPath(), name))
	go func() {
		err := a.client.Upload(path, remotePath)
		fyne.Do(func() {
			if err != nil {
				dialogShowError(a, err)
				return
			}
			a.remotePane.RefreshListing()
		})
	}()
}

func (a *App) downloadSelectedRemote() {
	if a.client == nil {
		return
	}
	entry := a.remotePane.SelectedEntry()
	if entry == nil || entry.IsDir {
		dialogShow(a, "Select a file", "Select a remote file to download.")
		return
	}
	localPath := filepath.Join(a.localPane.CurrentPath(), entry.Name)
	go func() {
		err := a.client.Download(entry.Path, localPath)
		fyne.Do(func() {
			if err != nil {
				dialogShowError(a, err)
				return
			}
			a.localPane.RefreshListing()
		})
	}()
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
