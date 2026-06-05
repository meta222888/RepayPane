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
	mw       *walk.MainWindow
	store    *config.Store
	settings *config.Settings

	tabs      []*TabSession
	activeTab int

	tabBar *walk.Composite

	localDriveCombo *walk.ComboBox
	localPathEdit   *walk.LineEdit
	remotePathEdit  *walk.LineEdit
	localTV         *walk.TableView
	remoteTV        *walk.TableView
	localModel      *dirModel
	remoteModel     *dirModel

	statusLabel    *walk.Label
	transferLabel  *walk.Label
	progressBar    *walk.ProgressBar
	reconnectBtn   *walk.PushButton

	localPath  string
	remotePath string

	client     *remote.Client
	server     config.Server
	connected  bool
	connecting bool

	transfers *TransferQueue
	clipboard *paneClipboard
}

func newApp(store *config.Store, settings *config.Settings) *App {
	return &App{
		store:      store,
		settings:   settings,
		localModel: newDirModel(),
		remoteModel: newDirModel(),
		localPath:  defaultLocalDir(),
		remotePath: "/",
		activeTab:  -1,
	}
}

func (a *App) initTransfers() {
	a.transfers = NewTransferQueue(a)
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

func (a *App) updateStatusBar() {
	tab := a.activeSession()
	var text string
	switch {
	case tab == nil:
		text = i18n.T(i18n.KeyNotConnected)
	case tab.state == tabConnecting:
		text = i18n.Tf(i18n.KeyConnecting, serverDisplayName(tab.server))
	case tab.state == tabConnected:
		text = i18n.Tf(i18n.KeyConnected, serverDisplayName(tab.server), tab.server.Username, tab.server.Host)
	default:
		text = i18n.T(i18n.KeyDisconnected)
	}
	a.setStatus(text)
	if a.reconnectBtn != nil {
		a.reconnectBtn.SetVisible(tab != nil && tab.state == tabDisconnected)
	}
}

func (a *App) setStatus(text string) {
	if a.statusLabel != nil {
		a.statusLabel.SetText(text)
	}
}

func (a *App) refreshTransferUI() {
	if a.transfers == nil {
		return
	}
	active, progress, speed, queue := a.transfers.Snapshot()
	a.syncUI(func() {
		if a.progressBar != nil {
			if active {
				a.progressBar.SetVisible(true)
				a.progressBar.SetValue(int(progress))
			} else {
				a.progressBar.SetVisible(false)
				a.progressBar.SetValue(0)
			}
		}
		if a.transferLabel != nil {
			if active || queue > 0 {
				a.transferLabel.SetText(speed + "  ·  " + i18n.Tf(i18n.KeyQueue, queue))
			} else {
				a.transferLabel.SetText(speed)
			}
		}
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
	if a.localDriveCombo != nil {
		a.syncDriveCombo()
	}
	if tab := a.activeSession(); tab != nil {
		tab.localPath = a.localPath
	}
}

func (a *App) refreshRemote() {
	if !a.connected || a.client == nil {
		a.remoteModel.setItems(nil)
		if a.remotePathEdit != nil {
			a.remotePathEdit.SetText(a.remotePath)
		}
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
	if tab := a.activeSession(); tab != nil {
		tab.remotePath = a.remotePath
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

func serverDisplayName(s config.Server) string {
	if s.Name != "" {
		return s.Name
	}
	return s.Host
}

func (a *App) setLanguage(lang i18n.Lang) {
	if i18n.Current() == lang {
		return
	}
	i18n.SetLanguage(lang)
	a.settings.Language = string(lang)
	_ = config.SaveSettings(a.settings)
	if a.mw != nil {
		a.mw.SetTitle(i18n.T(i18n.KeyAppTitle) + " (Win32)")
	}
}
