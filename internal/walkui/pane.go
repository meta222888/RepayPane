package walkui

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"github.com/lxn/walk"
)

func (a *App) navigateLocal(p string) {
	if p == "" {
		return
	}
	a.localPath = p
	a.refreshLocal()
}

func (a *App) navigateRemote(p string) {
	if p == "" || !a.connected {
		return
	}
	a.remotePath = p
	a.refreshRemote()
}

func (a *App) localUp() {
	parent := filepath.Dir(a.localPath)
	if parent == a.localPath {
		return
	}
	a.navigateLocal(parent)
}

func (a *App) remoteUp() {
	if !a.connected {
		return
	}
	dir := filepath.ToSlash(filepath.Dir(strings.ReplaceAll(a.remotePath, "\\", "/")))
	if dir == "." {
		dir = "/"
	}
	if !strings.HasPrefix(dir, "/") {
		dir = "/" + dir
	}
	a.navigateRemote(dir)
}

func (a *App) onLocalActivated() {
	for _, idx := range a.localTV.SelectedIndexes() {
		e, ok := a.localModel.entry(idx)
		if !ok {
			continue
		}
		if e.isDir {
			a.navigateLocal(e.fullPath)
			return
		}
		a.openLocalEditor(e.fullPath, e.name, e.size)
		return
	}
}

func (a *App) onRemoteActivated() {
	if !a.connected {
		return
	}
	for _, idx := range a.remoteTV.SelectedIndexes() {
		e, ok := a.remoteModel.entry(idx)
		if !ok {
			continue
		}
		if e.isDir {
			a.navigateRemote(e.fullPath)
			return
		}
		a.openRemoteEditor(remote.FileInfo{
			Name:    e.name,
			Path:    e.fullPath,
			Size:    e.size,
			IsDir:   false,
			ModTime: e.modTime,
		})
		return
	}
}

func (a *App) selectedLocalEntries() []dirEntry {
	var out []dirEntry
	for _, idx := range a.localTV.SelectedIndexes() {
		if e, ok := a.localModel.entry(idx); ok {
			out = append(out, e)
		}
	}
	return out
}

func (a *App) selectedRemoteEntries() []dirEntry {
	var out []dirEntry
	for _, idx := range a.remoteTV.SelectedIndexes() {
		if e, ok := a.remoteModel.entry(idx); ok {
			out = append(out, e)
		}
	}
	return out
}

func (a *App) uploadSelected() {
	client := a.activeClient()
	if client == nil {
		return
	}
	for _, e := range a.selectedLocalEntries() {
		if e.isDir {
			a.enqueueUploadTree(client, e.fullPath, joinRemote(a.remotePath, e.name), nil)
			continue
		}
		dst := joinRemote(a.remotePath, e.name)
		a.transfers.EnqueueUpload(client, e.fullPath, dst, func(err error) {
			if err != nil {
				a.showError(i18n.T(i18n.KeyUpload), err)
				return
			}
			a.refreshRemote()
		})
	}
}

func (a *App) downloadSelected() {
	client := a.activeClient()
	if client == nil {
		return
	}
	for _, e := range a.selectedRemoteEntries() {
		if e.isDir {
			a.enqueueDownloadTree(client, e.fullPath, filepath.Join(a.localPath, e.name), nil)
			continue
		}
		dst := filepath.Join(a.localPath, e.name)
		a.transfers.EnqueueDownload(client, e.fullPath, dst, func(err error) {
			if err != nil {
				a.showError(i18n.T(i18n.KeyDownload), err)
				return
			}
			a.refreshLocal()
		})
	}
}

func joinRemote(dir, name string) string {
	dir = strings.TrimSuffix(dir, "/")
	if dir == "" {
		dir = "/"
	}
	return dir + "/" + name
}

func (a *App) ctxCopyLocal() {
	items := a.selectedLocalEntries()
	if len(items) == 0 {
		return
	}
	clip := &paneClipboard{kind: paneLocal}
	for _, e := range items {
		clip.items = append(clip.items, clipItem{path: e.fullPath, name: e.name, isDir: e.isDir})
	}
	a.clipboard = clip
}

func (a *App) ctxCopyRemote() {
	items := a.selectedRemoteEntries()
	if len(items) == 0 {
		return
	}
	clip := &paneClipboard{kind: paneRemote}
	for _, e := range items {
		clip.items = append(clip.items, clipItem{path: e.fullPath, name: e.name, isDir: e.isDir})
	}
	a.clipboard = clip
}

func (a *App) ctxPasteLocal() {
	if a.clipboard == nil || len(a.clipboard.items) == 0 {
		return
	}
	switch a.clipboard.kind {
	case paneLocal:
		a.pasteLocalToLocal()
	case paneRemote:
		a.pasteRemoteToLocal()
	}
}

func (a *App) ctxPasteRemote() {
	if a.clipboard == nil || len(a.clipboard.items) == 0 {
		return
	}
	switch a.clipboard.kind {
	case paneRemote:
		a.pasteRemoteToRemote()
	case paneLocal:
		a.pasteLocalToRemote()
	}
}

func (a *App) pasteLocalToLocal() {
	for _, item := range a.clipboard.items {
		dst := filepath.Join(a.localPath, item.name)
		if pathExistsLocal(dst) {
			a.resolveFileConflict(item.name, pathExistsLocal, func(name string) {
				if name != "" {
					_ = copyPathLocal(item.path, filepath.Join(a.localPath, name))
					a.refreshLocal()
				}
			})
			continue
		}
		_ = copyPathLocal(item.path, dst)
	}
	a.refreshLocal()
}

func (a *App) pasteRemoteToRemote() {
	client := a.activeClient()
	if client == nil {
		return
	}
	for _, item := range a.clipboard.items {
		dst := joinRemote(a.remotePath, item.name)
		if destExistsRemote(client, dst) {
			a.resolveFileConflict(item.name, func(name string) bool {
				return destExistsRemote(client, joinRemote(a.remotePath, name))
			}, func(name string) {
				if name != "" {
					_ = client.CopyPath(item.path, joinRemote(a.remotePath, name))
					a.refreshRemote()
				}
			})
			continue
		}
		_ = client.CopyPath(item.path, dst)
	}
	a.refreshRemote()
}

func (a *App) pasteLocalToRemote() {
	client := a.activeClient()
	if client == nil {
		return
	}
	a.uploadClipboardItems(a.clipboard.items, a.remotePath, 0, func() { a.refreshRemote() })
}

func (a *App) pasteRemoteToLocal() {
	client := a.activeClient()
	if client == nil {
		return
	}
	a.downloadClipboardItems(a.clipboard.items, a.localPath, 0, func() { a.refreshLocal() })
}

func pathExistsLocal(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func destExistsRemote(client *remote.Client, p string) bool {
	_, err := client.Stat(p)
	return err == nil
}

func (a *App) ctxDeleteLocal() {
	items := a.selectedLocalEntries()
	if len(items) == 0 {
		return
	}
	a.showConfirm(i18n.T(i18n.KeyDelete), i18n.Tf(i18n.KeyDeleteMultiConfirm, len(items)), func() {
		for _, e := range items {
			_ = removePathLocal(e.fullPath)
		}
		a.refreshLocal()
	})
}

func (a *App) ctxDeleteRemote() {
	client := a.activeClient()
	items := a.selectedRemoteEntries()
	if client == nil || len(items) == 0 {
		return
	}
	a.showConfirm(i18n.T(i18n.KeyDelete), i18n.Tf(i18n.KeyDeleteMultiConfirm, len(items)), func() {
		go func() {
			for _, e := range items {
				_ = client.RemoveAll(e.fullPath)
			}
			a.syncUI(func() { a.refreshRemote() })
		}()
	})
}

func (a *App) ctxRenameLocal() {
	items := a.selectedLocalEntries()
	if len(items) != 1 {
		return
	}
	e := items[0]
	name, ok := a.promptInput(i18n.T(i18n.KeyRename), i18n.T(i18n.KeyRenamePrompt), e.name)
	if !ok || strings.TrimSpace(name) == "" || name == e.name {
		return
	}
	dst := filepath.Join(filepath.Dir(e.fullPath), name)
	if err := os.Rename(e.fullPath, dst); err != nil {
		a.showError(i18n.T(i18n.KeyRename), err)
		return
	}
	a.refreshLocal()
}

func (a *App) ctxRenameRemote() {
	client := a.activeClient()
	items := a.selectedRemoteEntries()
	if client == nil || len(items) != 1 {
		return
	}
	e := items[0]
	name, ok := a.promptInput(i18n.T(i18n.KeyRename), i18n.T(i18n.KeyRenamePrompt), e.name)
	if !ok || strings.TrimSpace(name) == "" || name == e.name {
		return
	}
	dst := joinRemote(path.Dir(strings.ReplaceAll(e.fullPath, "\\", "/")), name)
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

func (a *App) ctxNewFolderLocal() {
	name, ok := a.promptInput(i18n.T(i18n.KeyNewFolder), i18n.T(i18n.KeyRenamePrompt), "New Folder")
	if !ok || strings.TrimSpace(name) == "" {
		return
	}
	if err := os.Mkdir(filepath.Join(a.localPath, name), 0o755); err != nil {
		a.showError(i18n.T(i18n.KeyNewFolder), err)
		return
	}
	a.refreshLocal()
}

func (a *App) ctxNewFolderRemote() {
	client := a.activeClient()
	if client == nil {
		return
	}
	name, ok := a.promptInput(i18n.T(i18n.KeyNewFolder), i18n.T(i18n.KeyRenamePrompt), "New Folder")
	if !ok || strings.TrimSpace(name) == "" {
		return
	}
	go func() {
		err := client.Mkdir(joinRemote(a.remotePath, name))
		a.syncUI(func() {
			if err != nil {
				a.showError(i18n.T(i18n.KeyNewFolder), err)
				return
			}
			a.refreshRemote()
		})
	}()
}

func (a *App) uploadClipboardItems(items []clipItem, remoteDir string, idx int, onDone func()) {
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
		return destExistsRemote(client, joinRemote(remoteDir, name))
	}
	start := func(name string) {
		if name == "" {
			a.uploadClipboardItems(items, remoteDir, idx+1, onDone)
			return
		}
		a.doUploadClipItem(item, remoteDir, name, func() {
			a.uploadClipboardItems(items, remoteDir, idx+1, onDone)
		})
	}
	if exists(item.name) {
		a.resolveFileConflict(item.name, exists, start)
		return
	}
	start(item.name)
}

func (a *App) doUploadClipItem(item clipItem, remoteDir, destName string, onDone func()) {
	client := a.activeClient()
	if client == nil {
		return
	}
	dst := joinRemote(remoteDir, destName)
	if item.isDir {
		a.enqueueUploadTree(client, item.path, dst, onDone)
		return
	}
	a.transfers.EnqueueUpload(client, item.path, dst, func(err error) {
		if err != nil {
			a.showError(i18n.T(i18n.KeyUpload), err)
			return
		}
		if onDone != nil {
			onDone()
		}
	})
}

func (a *App) downloadClipboardItems(items []clipItem, localDir string, idx int, onDone func()) {
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
		return pathExistsLocal(filepath.Join(localDir, name))
	}
	start := func(name string) {
		if name == "" {
			a.downloadClipboardItems(items, localDir, idx+1, onDone)
			return
		}
		a.doDownloadClipItem(item, localDir, name, func() {
			a.downloadClipboardItems(items, localDir, idx+1, onDone)
		})
	}
	if exists(item.name) {
		a.resolveFileConflict(item.name, exists, start)
		return
	}
	start(item.name)
}

func (a *App) doDownloadClipItem(item clipItem, localDir, destName string, onDone func()) {
	client := a.activeClient()
	if client == nil {
		return
	}
	dst := filepath.Join(localDir, destName)
	if item.isDir {
		a.enqueueDownloadTree(client, item.path, dst, onDone)
		return
	}
	a.transfers.EnqueueDownload(client, item.path, dst, func(err error) {
		if err != nil {
			a.showError(i18n.T(i18n.KeyDownload), err)
			return
		}
		if onDone != nil {
			onDone()
		}
	})
}

func (a *App) enqueueUploadTree(client *remote.Client, localPath, remotePath string, onDone func()) {
	go func() {
		err := filepath.Walk(localPath, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(localPath, p)
			if err != nil {
				return nil
			}
			rp := filepath.ToSlash(filepath.Join(remotePath, rel))
			done := make(chan struct{})
			a.transfers.EnqueueUpload(client, p, rp, func(error) { close(done) })
			<-done
			return nil
		})
		a.syncUI(func() {
			if err != nil {
				a.showError(i18n.T(i18n.KeyUpload), err)
			}
			if onDone != nil {
				onDone()
			}
			a.refreshRemote()
		})
	}()
}

func (a *App) enqueueDownloadTree(client *remote.Client, remotePath, localPath string, onDone func()) {
	go func() {
		var walkDir func(dir string)
		walkDir = func(dir string) {
			entries, err := client.ListDir(dir)
			if err != nil {
				return
			}
			for _, e := range entries {
				if e.IsDir {
					subLocal := filepath.Join(localPath, relRemotePath(remotePath, e.Path))
					_ = os.MkdirAll(subLocal, 0o755)
					walkDir(e.Path)
					continue
				}
				rel, _ := filepath.Rel(remotePath, filepath.FromSlash(e.Path))
				lp := filepath.Join(localPath, rel)
				done := make(chan struct{})
				a.transfers.EnqueueDownload(client, e.Path, lp, func(error) { close(done) })
				<-done
			}
		}
		_ = os.MkdirAll(localPath, 0o755)
		walkDir(remotePath)
		a.syncUI(func() {
			if onDone != nil {
				onDone()
			}
			a.refreshLocal()
		})
	}()
}

func relRemotePath(base, full string) string {
	base = strings.TrimSuffix(base, "/")
	full = strings.TrimPrefix(full, base)
	full = strings.TrimPrefix(full, "/")
	return filepath.FromSlash(full)
}

func (a *App) resolveFileConflict(fileName string, exists func(string) bool, onProceed func(string)) {
	msg := i18n.Tf(i18n.KeyFileExistsConflict, fileName)
	a.syncUI(func() {
		switch walk.MsgBox(a.mw, i18n.T(i18n.KeyFileExistsTitle), msg+"\n\nYes=Overwrite, No=Cancel, Cancel=Rename",
			walk.MsgBoxYesNoCancel|walk.MsgBoxIconWarning) {
		case walk.DlgCmdYes:
			onProceed(fileName)
		case walk.DlgCmdCancel:
			name, ok := a.promptInput(i18n.T(i18n.KeyRename), i18n.T(i18n.KeyRenamePrompt), suggestCopyName(fileName, exists))
			if ok && name != "" {
				if exists(name) {
					a.resolveFileConflict(name, exists, onProceed)
					return
				}
				onProceed(name)
			} else {
				onProceed("")
			}
		default:
			onProceed("")
		}
	})
}

func suggestCopyName(original string, exists func(string) bool) string {
	for i := 1; i < 1000; i++ {
		candidate := appendCopySuffix(original, i)
		if !exists(candidate) {
			return candidate
		}
	}
	return original + " copy"
}

func appendCopySuffix(name string, n int) string {
	dot := strings.LastIndex(name, ".")
	if dot > 0 {
		return fmt.Sprintf("%s (%d)%s", name[:dot], n, name[dot:])
	}
	return fmt.Sprintf("%s (%d)", name, n)
}

func (a *App) syncDriveCombo() {
	// drive combo updates via user selection; path edit shows full path
}

func localDriveLabel(p string) string {
	if len(p) >= 2 && p[1] == ':' {
		return strings.ToUpper(p[:2]) + `\`
	}
	return p
}

func listWindowsDrives() []string {
	var drives []string
	for c := 'A'; c <= 'Z'; c++ {
		root := string(c) + `:\`
		if _, err := os.Stat(root); err == nil {
			drives = append(drives, root)
		}
	}
	return drives
}

type placeEntry struct {
	label string
	path  string
}

func commonPlaces() []placeEntry {
	home, _ := os.UserHomeDir()
	type candidate struct {
		key  string
		sub  string
		fallback string
	}
	candidates := []candidate{
		{i18n.KeyPlaceDesktop, "Desktop", filepath.Join(home, "Desktop")},
		{i18n.KeyPlaceDocuments, "Documents", filepath.Join(home, "Documents")},
		{i18n.KeyPlacePictures, "Pictures", filepath.Join(home, "Pictures")},
		{i18n.KeyPlaceDownloads, "Downloads", filepath.Join(home, "Downloads")},
		{i18n.KeyPlaceMusic, "Music", filepath.Join(home, "Music")},
		{i18n.KeyPlaceVideos, "Videos", filepath.Join(home, "Videos")},
		{i18n.KeyPlaceHome, "", home},
	}
	var out []placeEntry
	for _, c := range candidates {
		p := c.fallback
		if c.sub == "" {
			p = home
		}
		if st, err := os.Stat(p); err != nil || !st.IsDir() {
			continue
		}
		out = append(out, placeEntry{label: i18n.T(c.key), path: p})
	}
	return out
}

func (a *App) refreshTabBar() {
	if a.tabBar == nil {
		return
	}
	a.syncUI(func() {
		a.tabBar.Children().Clear()
		for i, tab := range a.tabs {
			idx := i
			btn, _ := walk.NewPushButton(a.tabBar)
			btn.SetText(tab.tabLabel())
			if i == a.activeTab {
				btn.SetEnabled(true)
			}
			btn.Clicked().Attach(func() { a.activateTab(idx) })
			closeBtn, _ := walk.NewPushButton(a.tabBar)
			closeBtn.SetText("×")
			closeBtn.SetMinMaxSize(walk.Size{Width: 28, Height: 0}, walk.Size{Width: 28, Height: 0})
			closeBtn.Clicked().Attach(func() { a.closeTab(idx) })
			switch tab.state {
			case tabConnected:
				// default
			case tabConnecting:
				btn.SetText(tab.tabLabel() + " …")
			default:
				btn.SetText(tab.tabLabel() + " !")
			}
			_ = tab
		}
		addBtn, _ := walk.NewPushButton(a.tabBar)
		addBtn.SetText(i18n.T(i18n.KeyNewTabConnect))
		addBtn.Clicked().Attach(a.onNewTab)
	})
}
