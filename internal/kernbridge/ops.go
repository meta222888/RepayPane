package kernbridge

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"
)

func (a *App) UploadPaths(localPaths []string) {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	remoteDir := a.remotePath
	for _, lp := range localPaths {
		info, err := os.Stat(lp)
		if err != nil {
			continue
		}
		if info.IsDir() {
			a.enqueueUploadTree(client, lp, filepath.ToSlash(path.Join(remoteDir, filepath.Base(lp))), nil)
			continue
		}
		rp := filepath.ToSlash(path.Join(remoteDir, filepath.Base(lp)))
		a.transfers.EnqueueUpload(client, lp, rp, func(err error) {
			if err != nil {
				a.host.showError(i18n.T(i18n.KeyUpload), err)
			}
			a.host.refreshRemote()
		})
	}
}

func (a *App) DownloadPaths(remotePaths []string) {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	localDir := a.localPath
	for _, rp := range remotePaths {
		st, err := client.Stat(rp)
		if err != nil {
			continue
		}
		if st.IsDir {
			a.enqueueDownloadTree(client, rp, filepath.Join(localDir, filepath.Base(rp)), nil)
			continue
		}
		lp := filepath.Join(localDir, filepath.Base(rp))
		a.transfers.EnqueueDownload(client, rp, lp, func(err error) {
			if err != nil {
				a.host.showError(i18n.T(i18n.KeyDownload), err)
			}
			a.host.refreshLocal()
		})
	}
}

func (a *App) enqueueUploadTree(client *remote.Client, localPath, remotePath string, onDone func()) {
	_ = filepath.Walk(localPath, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(localPath, p)
		if err != nil {
			return nil
		}
		rp := filepath.ToSlash(path.Join(remotePath, rel))
		a.transfers.EnqueueUpload(client, p, rp, nil)
		return nil
	})
	if onDone != nil {
		onDone()
	}
}

func (a *App) enqueueDownloadTree(client *remote.Client, remotePath, localPath string, onDone func()) {
	walkRemoteTree(client, remotePath, func(p string, isDir bool) {
		if isDir {
			return
		}
		rel, err := filepath.Rel(remotePath, filepath.FromSlash(p))
		if err != nil {
			return
		}
		lp := filepath.Join(localPath, rel)
		a.transfers.EnqueueDownload(client, p, lp, nil)
	})
	if onDone != nil {
		onDone()
	}
}

func walkRemoteTree(client *remote.Client, dir string, fn func(path string, isDir bool)) {
	entries, err := client.ListDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		fn(e.Path, e.IsDir)
		if e.IsDir {
			walkRemoteTree(client, e.Path, fn)
		}
	}
}

func (a *App) DeleteLocal(paths []string) {
	go func() {
		for _, p := range paths {
			if err := removePathLocal(p); err != nil {
				a.host.showError(i18n.T(i18n.KeyLocal), err)
			}
		}
		a.host.refreshLocal()
	}()
}

func (a *App) DeleteRemote(paths []string) {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	go func() {
		for _, p := range paths {
			if err := client.RemoveAll(p); err != nil {
				a.host.showError(i18n.T(i18n.KeyRemote), err)
			}
		}
		a.host.refreshRemote()
	}()
}

func (a *App) RenameLocal(oldPath, newName string) error {
	newPath := filepath.Join(filepath.Dir(oldPath), newName)
	return os.Rename(oldPath, newPath)
}

func (a *App) RenameRemote(oldPath, newName string) error {
	client, ok := a.requireClient()
	if !ok {
		return fmt.Errorf("not connected")
	}
	newPath := path.Join(path.Dir(oldPath), newName)
	return client.Rename(oldPath, newPath)
}

func (a *App) NewFolderLocal(name string) error {
	return os.MkdirAll(filepath.Join(a.localPath, name), 0o755)
}

func (a *App) NewFolderRemote(name string) error {
	client, ok := a.requireClient()
	if !ok {
		return fmt.Errorf("not connected")
	}
	return client.Mkdir(path.Join(a.remotePath, name))
}

func (a *App) NewFileLocal(name string) error {
	f, err := os.Create(filepath.Join(a.localPath, name))
	if err != nil {
		return err
	}
	return f.Close()
}

func (a *App) NewFileRemote(name string) error {
	client, ok := a.requireClient()
	if !ok {
		return fmt.Errorf("not connected")
	}
	return client.WriteFile(path.Join(a.remotePath, name), nil)
}

func (a *App) CopyClipboard(local bool, paths []string, names []string, isDirs []bool) {
	items := make([]clipItem, len(paths))
	for i := range paths {
		items[i] = clipItem{path: paths[i], name: names[i], isDir: isDirs[i]}
	}
	a.clipboard = &paneClipboard{local: local, items: items}
}

func (a *App) PasteToLocal() {
	if a.clipboard == nil {
		return
	}
	if a.clipboard.local {
		for _, item := range a.clipboard.items {
			dst := filepath.Join(a.localPath, item.name)
			_ = copyPathLocal(item.path, dst)
		}
		a.host.refreshLocal()
		return
	}
	client, ok := a.requireClient()
	if !ok {
		return
	}
	for _, item := range a.clipboard.items {
		if item.isDir {
			a.enqueueDownloadTree(client, item.path, filepath.Join(a.localPath, item.name), nil)
		} else {
			lp := filepath.Join(a.localPath, item.name)
			a.transfers.EnqueueDownload(client, item.path, lp, nil)
		}
	}
}

func (a *App) PasteToRemote() {
	if a.clipboard == nil {
		return
	}
	client, ok := a.requireClient()
	if !ok {
		return
	}
	if !a.clipboard.local {
		for _, item := range a.clipboard.items {
			dst := path.Join(a.remotePath, item.name)
			_ = client.CopyPath(item.path, dst)
		}
		a.host.refreshRemote()
		return
	}
	for _, item := range a.clipboard.items {
		if item.isDir {
			a.enqueueUploadTree(client, item.path, path.Join(a.remotePath, item.name), nil)
		} else {
			rp := path.Join(a.remotePath, item.name)
			a.transfers.EnqueueUpload(client, item.path, rp, nil)
		}
	}
}

func (a *App) SyncUpload() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	localDir := a.localPath
	remoteDir := a.remotePath
	go func() {
		_ = filepath.Walk(localDir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(localDir, p)
			if err != nil {
				return nil
			}
			rp := filepath.ToSlash(path.Join(remoteDir, rel))
			a.transfers.EnqueueUpload(client, p, rp, nil)
			return nil
		})
		a.host.refreshRemote()
	}()
}

func (a *App) SyncDownload() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	localDir := a.localPath
	remoteDir := a.remotePath
	go func() {
		walkRemoteTree(client, remoteDir, func(p string, isDir bool) {
			if isDir {
				return
			}
			rel, err := filepath.Rel(remoteDir, filepath.FromSlash(p))
			if err != nil {
				return
			}
			lp := filepath.Join(localDir, rel)
			a.transfers.EnqueueDownload(client, p, lp, nil)
		})
		a.host.refreshLocal()
	}()
}

func (a *App) RunShell(cmd string) string {
	client, ok := a.requireClient()
	if !ok {
		return mustJSON(map[string]string{"error": "not connected"})
	}
	out, err := client.RunCombined(cmd)
	if err != nil {
		if strings.TrimSpace(out) == "" {
			return mustJSON(map[string]string{"error": err.Error()})
		}
		return mustJSON(map[string]string{"output": strings.TrimSpace(out), "error": err.Error()})
	}
	a.pushShellHistory(cmd)
	return mustJSON(map[string]string{"output": strings.TrimSpace(out)})
}

func (a *App) pushShellHistory(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	h := a.settings.ShellHistory
	for i, c := range h {
		if c == cmd {
			h = append(h[:i], h[i+1:]...)
			break
		}
	}
	h = append(h, cmd)
	if len(h) > 100 {
		h = h[len(h)-100:]
	}
	a.settings.ShellHistory = h
	_ = config.SaveSettings(a.settings)
}

func copyPathLocal(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDirLocal(src, dst, info.Mode())
	}
	return copyFileLocal(src, dst)
}

func copyFileLocal(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyDirLocal(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(dst, mode); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyPathLocal(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func removePathLocal(p string) error {
	info, err := os.Stat(p)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return os.RemoveAll(p)
	}
	return os.Remove(p)
}

func (a *App) ListDrivesJSON() string {
	var drives []string
	for letter := 'A'; letter <= 'Z'; letter++ {
		root := string(letter) + `:\`
		if _, err := os.Stat(root); err == nil {
			drives = append(drives, string(letter)+":")
		}
	}
	return mustJSON(drives)
}

func (a *App) RefreshBoth() {
	a.host.refreshLocal()
	a.host.refreshRemote()
}
