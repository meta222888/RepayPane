package walkui

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"github.com/lxn/walk"
)

type App struct {
	mw    *walk.MainWindow
	store *config.Store

	localPathEdit  *walk.LineEdit
	remotePathEdit *walk.LineEdit
	localTV        *walk.TableView
	remoteTV       *walk.TableView
	localModel     *dirModel
	remoteModel    *dirModel
	statusLabel    *walk.Label
	reconnectBtn   *walk.PushButton

	localPath  string
	remotePath string

	client     *remote.Client
	server     config.Server
	connected  bool
	connecting bool
}

func newApp(store *config.Store) *App {
	return &App{
		store:      store,
		localModel: newDirModel(),
		remoteModel: newDirModel(),
		localPath:  defaultLocalDir(),
		remotePath: "/",
	}
}

func (a *App) dialServer(s config.Server) (*remote.Client, error) {
	var pass []byte
	if remote.NeedsPassphrase(s.AutoSSHKey, s.PrivateKey) {
		pass = a.askPassphrase()
		if len(pass) == 0 {
			return nil, remote.ErrPassphraseRequired
		}
	}
	client, err := a.tryConnect(s, pass)
	if errors.Is(err, remote.ErrPassphraseRequired) {
		pass = a.askPassphrase()
		if len(pass) == 0 {
			return nil, err
		}
		return a.tryConnect(s, pass)
	}
	return client, err
}

func (a *App) tryConnect(s config.Server, passphrase []byte) (*remote.Client, error) {
	return remote.Connect(remote.ConnectOptions{
		Host:          s.Host,
		Port:          s.Port,
		Username:      s.Username,
		Password:      s.Password,
		AutoSSHKey:    s.AutoSSHKey,
		PrivateKey:    s.PrivateKey,
		KeyPassphrase: passphrase,
	})
}

func defaultRemoteRoot(s *config.Server) string {
	if s.RemoteRoot != "" {
		return s.RemoteRoot
	}
	return "/"
}

func defaultLocalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "C:\\"
	}
	return home
}

func (a *App) setStatus(text string) {
	if a.statusLabel == nil {
		return
	}
	a.statusLabel.SetText(text)
}

func (a *App) syncUI(fn func()) {
	if a.mw == nil {
		fn()
		return
	}
	a.mw.Synchronize(fn)
}

func (a *App) showError(title string, err error) {
	if err == nil {
		return
	}
	a.syncUI(func() {
		walk.MsgBox(a.mw, title, err.Error(), walk.MsgBoxIconError)
	})
}

func (a *App) refreshLocal() {
	entries, err := listLocalDir(a.localPath)
	if err != nil {
		a.showError(i18n.T(i18n.KeyLocal), err)
		return
	}
	a.localModel.setItems(entries)
	if a.localPathEdit != nil {
		a.localPathEdit.SetText(a.localPath)
	}
}

func (a *App) refreshRemote() {
	if !a.connected || a.client == nil {
		a.remoteModel.setItems(nil)
		return
	}
	entries, err := listRemoteDir(a.client, a.remotePath)
	if err != nil {
		a.showError(i18n.T(i18n.KeyRemote), err)
		return
	}
	a.remoteModel.setItems(entries)
	if a.remotePathEdit != nil {
		a.remotePathEdit.SetText(a.remotePath)
	}
}

func listLocalDir(dir string) ([]dirEntry, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	out := make([]dirEntry, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		out = append(out, dirEntry{
			name:     name,
			fullPath: full,
			size:     info.Size(),
			modTime:  info.ModTime(),
			isDir:    info.IsDir(),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].isDir != out[j].isDir {
			return out[i].isDir
		}
		return strings.ToLower(out[i].name) < strings.ToLower(out[j].name)
	})
	return out, nil
}

func listRemoteDir(client *remote.Client, dir string) ([]dirEntry, error) {
	files, err := client.ListDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]dirEntry, 0, len(files))
	for _, f := range files {
		out = append(out, dirEntry{
			name:     f.Name,
			fullPath: f.Path,
			size:     f.Size,
			modTime:  f.ModTime,
			isDir:    f.IsDir,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].isDir != out[j].isDir {
			return out[i].isDir
		}
		return strings.ToLower(out[i].name) < strings.ToLower(out[j].name)
	})
	return out, nil
}

func (a *App) disconnect() {
	if a.client != nil {
		_ = a.client.Close()
		a.client = nil
	}
	a.connected = false
	a.connecting = false
	a.remoteModel.setItems(nil)
	a.setStatus(i18n.T(i18n.KeyNotConnected))
	if a.reconnectBtn != nil {
		a.reconnectBtn.SetVisible(false)
	}
}

func (a *App) connectServer(s config.Server) {
	if strings.TrimSpace(s.Host) == "" || strings.TrimSpace(s.Username) == "" {
		a.showError(i18n.T(i18n.KeyServerFormTitle), errors.New(i18n.T(i18n.KeyFormRequired)))
		return
	}
	if a.connecting {
		return
	}
	a.disconnect()
	a.server = s
	a.remotePath = defaultRemoteRoot(&s)
	a.connecting = true
	a.setStatus(i18n.Tf(i18n.KeyConnecting, serverDisplayName(s)))

	go func() {
		client, err := a.dialServer(s)
		a.syncUI(func() {
			a.connecting = false
			if err != nil {
				a.setStatus(i18n.T(i18n.KeyConnectionFailed))
				a.showError(i18n.T(i18n.KeyConnectionFailed), err)
				return
			}
			a.client = client
			a.connected = true
			a.setStatus(i18n.Tf(i18n.KeyConnected, serverDisplayName(s), s.Username, s.Host))
			if a.reconnectBtn != nil {
				a.reconnectBtn.SetVisible(false)
			}
			if interval := s.HeartbeatInterval(); interval > 0 {
				client.StartHeartbeat(interval, func(err error) {
					a.syncUI(func() {
						if err != nil {
							a.handleDisconnect(err)
						}
					})
				})
			}
			a.refreshRemote()
		})
	}()
}

func (a *App) handleDisconnect(err error) {
	a.disconnect()
	a.setStatus(i18n.T(i18n.KeyConnectionLost))
	if a.reconnectBtn != nil {
		a.reconnectBtn.SetVisible(true)
	}
	if err != nil {
		a.showError(i18n.T(i18n.KeyDisconnected), err)
	}
}

func serverDisplayName(s config.Server) string {
	if s.Name != "" {
		return s.Name
	}
	return s.Host
}

func (a *App) navigateLocal(path string) {
	if path == "" {
		return
	}
	a.localPath = path
	a.refreshLocal()
}

func (a *App) navigateRemote(path string) {
	if path == "" || !a.connected {
		return
	}
	a.remotePath = path
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
	idx := a.localTV.CurrentIndex()
	e, ok := a.localModel.entry(idx)
	if !ok {
		return
	}
	if e.isDir {
		a.navigateLocal(e.fullPath)
	}
}

func (a *App) onRemoteActivated() {
	if !a.connected {
		return
	}
	idx := a.remoteTV.CurrentIndex()
	e, ok := a.remoteModel.entry(idx)
	if !ok {
		return
	}
	if e.isDir {
		a.navigateRemote(e.fullPath)
	}
}

func (a *App) uploadSelected() {
	if !a.connected || a.client == nil {
		return
	}
	idx := a.localTV.CurrentIndex()
	e, ok := a.localModel.entry(idx)
	if !ok || e.isDir {
		return
	}
	remoteDest := strings.TrimSuffix(a.remotePath, "/") + "/" + e.name
	a.setStatus(i18n.T(i18n.KeyUpload) + " " + e.name)
	go func() {
		err := a.client.Upload(e.fullPath, remoteDest)
		a.syncUI(func() {
			if err != nil {
				a.showError(i18n.T(i18n.KeyUpload), err)
			} else {
				a.setStatus(i18n.T(i18n.KeyConnected))
				a.refreshRemote()
			}
		})
	}()
}

func (a *App) downloadSelected() {
	if !a.connected || a.client == nil {
		return
	}
	idx := a.remoteTV.CurrentIndex()
	e, ok := a.remoteModel.entry(idx)
	if !ok || e.isDir {
		return
	}
	localDest := filepath.Join(a.localPath, e.name)
	a.setStatus(i18n.T(i18n.KeyDownload) + " " + e.name)
	go func() {
		err := a.client.Download(e.fullPath, localDest)
		a.syncUI(func() {
			if err != nil {
				a.showError(i18n.T(i18n.KeyDownload), err)
			} else {
				a.setStatus(i18n.T(i18n.KeyConnected))
				a.refreshLocal()
			}
		})
	}()
}
