package kernbridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/relaypane/relaypane/internal/cloudsync"
	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/fileopen"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"
	"github.com/relaypane/relaypane/internal/textencoding"
	"github.com/relaypane/relaypane/internal/update"
	"github.com/relaypane/relaypane/internal/version"
)

type tabState int

const (
	tabDisconnected tabState = iota
	tabConnecting
	tabConnected
)

type TabSession struct {
	server     config.Server
	client     *remote.Client
	state      tabState
	remotePath string
	localPath  string
}

type App struct {
	mu sync.Mutex

	store    *config.Store
	settings *config.Settings
	host     *uiHost

	tabs      []*TabSession
	activeTab int

	localPath  string
	remotePath string
	server     config.Server
	client     *remote.Client
	connected  bool

	transfers *TransferQueue
	clipboard *paneClipboard

	passphraseCh map[int]chan []byte
	passphraseMu sync.Mutex

	editorEnc textencoding.Info
}

type paneClipboard struct {
	local bool
	items []clipItem
}

type clipItem struct {
	path  string
	name  string
	isDir bool
}

var globalApp *App

func InitApp() error {
	store, err := config.Load()
	if err != nil {
		return err
	}
	settings, err := config.LoadSettings()
	if err != nil {
		settings = &config.Settings{}
	}
	initLanguage(settings)
	a := &App{
		store:        store,
		settings:     settings,
		activeTab:    -1,
		localPath:    defaultLocalDir(),
		remotePath:   "/",
		passphraseCh: make(map[int]chan []byte),
	}
	a.host = &uiHost{app: a}
	a.transfers = NewTransferQueue(a)
	globalApp = a
	return nil
}

func AppInstance() *App {
	return globalApp
}

func initLanguage(settings *config.Settings) {
	switch settings.Language {
	case "en":
		i18n.SetLanguage(i18n.EN)
	case "zh":
		i18n.SetLanguage(i18n.ZH)
	default:
		if lang := os.Getenv("LANG"); strings.HasPrefix(strings.ToLower(lang), "en") {
			i18n.SetLanguage(i18n.EN)
			settings.Language = "en"
		} else {
			i18n.SetLanguage(i18n.ZH)
			settings.Language = "zh"
		}
	}
}

func defaultLocalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return `C:\`
	}
	return home
}

func defaultRemoteRoot(s *config.Server) string {
	if s.RemoteRoot != "" {
		return s.RemoteRoot
	}
	return "/"
}

func serverDisplayName(s config.Server) string {
	if s.Name != "" {
		return s.Name
	}
	return s.Host
}

func (a *App) activeSession() *TabSession {
	if a.activeTab < 0 || a.activeTab >= len(a.tabs) {
		return nil
	}
	return a.tabs[a.activeTab]
}

func (a *App) activeClient() *remote.Client {
	if s := a.activeSession(); s != nil {
		return s.client
	}
	return nil
}

func (a *App) requireClient() (*remote.Client, bool) {
	c := a.activeClient()
	if c == nil {
		a.host.showMsg(i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedFirst))
		return nil, false
	}
	return c, true
}

func (a *App) waitPassphrase(id int) []byte {
	ch := make(chan []byte, 1)
	a.passphraseMu.Lock()
	a.passphraseCh[id] = ch
	a.passphraseMu.Unlock()
	defer func() {
		a.passphraseMu.Lock()
		delete(a.passphraseCh, id)
		a.passphraseMu.Unlock()
	}()
	return <-ch
}

func (a *App) SubmitPassphrase(id int, pass string) {
	a.passphraseMu.Lock()
	ch := a.passphraseCh[id]
	a.passphraseMu.Unlock()
	if ch != nil {
		ch <- []byte(pass)
	}
}

func (a *App) dialServer(s config.Server) (*remote.Client, error) {
	var pass []byte
	if remote.NeedsPassphrase(s.AutoSSHKey, s.PrivateKey) {
		pass = a.host.askPassphrase()
		if len(pass) == 0 {
			return nil, remote.ErrPassphraseRequired
		}
	}
	client, err := a.tryConnect(s, pass)
	if errors.Is(err, remote.ErrPassphraseRequired) {
		pass = a.host.askPassphrase()
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

func (a *App) StatusJSON() string  { return a.statusJSON() }
func (a *App) TabsJSON() string    { return a.tabsJSON() }

func (a *App) statusJSON() string {
	tab := a.activeSession()
	st := StatusJSON{WindowTitle: i18n.T(i18n.KeyAppTitle)}
	switch {
	case tab == nil:
		st.Text = i18n.T(i18n.KeyNotConnected)
		st.ConnDotColor = "red"
	case tab.state == tabConnecting:
		st.Text = i18n.Tf(i18n.KeyConnecting, serverDisplayName(tab.server))
		st.Connecting = true
		st.ConnDotColor = "yellow"
	case tab.state == tabConnected:
		st.Text = i18n.Tf(i18n.KeyConnected, serverDisplayName(tab.server), tab.server.Username, tab.server.Host)
		st.Connected = true
		st.ConnDotColor = "green"
		if tab.server.HeartbeatSec > 0 {
			st.HeartbeatLabel = fmt.Sprintf("%ds", tab.server.HeartbeatSec)
		}
	default:
		st.Text = i18n.T(i18n.KeyDisconnected)
		st.ShowReconnect = true
		st.ConnDotColor = "red"
	}
	if tab != nil {
		st.WindowTitle += " - " + serverDisplayName(tab.server)
	}
	return mustJSON(st)
}

func (a *App) tabsJSON() string {
	out := make([]TabJSON, len(a.tabs))
	for i, t := range a.tabs {
		state := "disconnected"
		switch t.state {
		case tabConnecting:
			state = "connecting"
		case tabConnected:
			state = "connected"
		}
		out[i] = TabJSON{
			Index:        i,
			Label:        serverDisplayName(t.server),
			State:        state,
			LocalPath:    t.localPath,
			RemotePath:   t.remotePath,
			HeartbeatSec: t.server.HeartbeatSec,
		}
	}
	return mustJSON(out)
}

func (a *App) ServersJSON() string {
	return mustJSON(a.store.Servers)
}

func (a *App) SettingsJSON() string {
	return mustJSON(a.settings)
}

func (a *App) SaveSettingsJSON(data string) error {
	var s config.Settings
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return err
	}
	a.settings = &s
	return config.SaveSettings(a.settings)
}

func (a *App) SetLanguage(lang string) {
	var l i18n.Lang
	switch lang {
	case "en":
		l = i18n.EN
	default:
		l = i18n.ZH
	}
	if i18n.Current() == l {
		return
	}
	i18n.SetLanguage(l)
	a.settings.Language = string(l)
	_ = config.SaveSettings(a.settings)
	a.host.updateStatusBar()
}

func (a *App) CheckUpdate() string {
	rel, err := update.FetchLatestRelease()
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()})
	}
	newer := update.IsNewer(version.Version, rel.Version)
	return mustJSON(map[string]any{
		"current": version.Version,
		"remote":  rel.Version,
		"url":     rel.HTMLURL,
		"newer":   newer,
	})
}

func (a *App) CloudSyncStatus() string {
	if a.settings.CloudSyncAPISecret == "" {
		return mustJSON(map[string]string{"error": "missing api secret"})
	}
	st, err := cloudsync.NewClient(a.settings.CloudSyncAPISecret).QueryStatus()
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()})
	}
	return mustJSON(st)
}

func (a *App) CloudSyncUpload() string {
	if a.settings.CloudSyncAPISecret == "" || a.settings.CloudSyncPassword == "" {
		return mustJSON(map[string]string{"error": "missing credentials"})
	}
	updatedAt, err := cloudsync.Upload(a.store, a.settings.CloudSyncAPISecret, a.settings.CloudSyncPassword)
	if err != nil {
		logPath, _ := cloudsync.LogUploadError(err)
		return mustJSON(map[string]string{"error": err.Error(), "log": logPath})
	}
	if updatedAt != "" {
		a.settings.CloudSyncLastSyncAt = updatedAt
	} else {
		a.settings.CloudSyncLastSyncAt = time.Now().Format("2006-01-02 15:04:05")
	}
	_ = config.SaveSettings(a.settings)
	return mustJSON(map[string]string{"ok": "uploaded", "at": a.settings.CloudSyncLastSyncAt})
}

func (a *App) CloudSyncDownload() string {
	if a.settings.CloudSyncAPISecret == "" || a.settings.CloudSyncPassword == "" {
		return mustJSON(map[string]string{"error": "missing credentials"})
	}
	plain, st, err := cloudsync.Download(a.settings.CloudSyncAPISecret, a.settings.CloudSyncPassword)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()})
	}
	if !st.Exists || len(plain) == 0 {
		return mustJSON(map[string]string{"error": "no cloud data"})
	}
	newStore := &config.Store{}
	if err := cloudsync.ApplyPayload(plain, newStore); err != nil {
		return mustJSON(map[string]string{"error": err.Error()})
	}
	a.store.Servers = newStore.Servers
	_ = config.Save(a.store)
	a.settings.CloudSyncLastSyncAt = time.Now().Format("2006-01-02 15:04:05")
	_ = config.SaveSettings(a.settings)
	a.host.refreshTabs()
	return mustJSON(map[string]string{"ok": "downloaded", "at": a.settings.CloudSyncLastSyncAt})
}

func (a *App) CloudSyncDelete() string {
	if a.settings.CloudSyncAPISecret == "" {
		return mustJSON(map[string]string{"error": "missing api secret"})
	}
	if err := cloudsync.NewClient(a.settings.CloudSyncAPISecret).Delete(); err != nil {
		return mustJSON(map[string]string{"error": err.Error()})
	}
	return mustJSON(map[string]string{"ok": "deleted"})
}

func (a *App) OpenEditor(path string, remotePane bool) {
	if remotePane {
		client, ok := a.requireClient()
		if !ok {
			return
		}
		st, err := client.Stat(path)
		if err != nil {
			a.host.showError(i18n.T(i18n.KeyRemote), err)
			return
		}
		if st.IsDir {
			return
		}
		if fileopen.IsImageName(filepath.Base(path)) {
			go a.loadRemoteImage(client, path, filepath.Base(path))
			return
		}
		if st.Size > config.MaxEditBytes {
			a.host.showMsg(i18n.T(i18n.KeyFileTooLarge), i18n.Tf(i18n.KeyFileTooLargeMsg, filepath.Base(path), float64(st.Size)/(1024*1024)))
			return
		}
		go a.loadRemoteEditor(client, path)
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		a.host.showError(i18n.T(i18n.KeyLocal), err)
		return
	}
	if info.IsDir() {
		return
	}
	if fileopen.IsImageName(filepath.Base(path)) {
		_ = fileopen.OpenPath(path)
		return
	}
	if info.Size() > config.MaxEditBytes {
		a.host.showMsg(i18n.T(i18n.KeyFileTooLarge), i18n.Tf(i18n.KeyFileTooLargeMsg, filepath.Base(path), float64(info.Size())/(1024*1024)))
		return
	}
	go a.loadLocalEditor(path)
}

func (a *App) loadLocalEditor(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		a.host.showError(i18n.T(i18n.KeyEditTitle), err)
		return
	}
	if !fileopen.IsLikelyText(data) {
		a.host.showMsg(i18n.T(i18n.KeyNotTextFileTitle), i18n.Tf(i18n.KeyNotTextFileMsg, filepath.Base(path)))
		return
	}
	text, enc, err := textencoding.Decode(data)
	if err != nil {
		a.host.showError(i18n.T(i18n.KeyEditTitle), err)
		return
	}
	a.editorEnc = enc
	a.host.showEditor(filepath.Base(path), path, text, enc.Label(), false, int64(len(data)))
}

func (a *App) loadRemoteEditor(client *remote.Client, path string) {
	data, err := client.ReadFile(path)
	if err != nil {
		a.host.showError(i18n.T(i18n.KeyEditTitle), err)
		return
	}
	if !fileopen.IsLikelyText(data) {
		a.host.showMsg(i18n.T(i18n.KeyNotTextFileTitle), i18n.Tf(i18n.KeyNotTextFileMsg, filepath.Base(path)))
		return
	}
	text, enc, err := textencoding.Decode(data)
	if err != nil {
		a.host.showError(i18n.T(i18n.KeyEditTitle), err)
		return
	}
	a.editorEnc = enc
	a.host.showEditor(filepath.Base(path), path, text, enc.Label(), true, int64(len(data)))
}

func (a *App) loadRemoteImage(client *remote.Client, path, name string) {
	data, err := client.ReadFile(path)
	if err != nil {
		a.host.showError(i18n.T(i18n.KeyRemote), err)
		return
	}
	tmp, err := os.CreateTemp("", "relaypane-view-*"+filepath.Ext(name))
	if err != nil {
		a.host.showError(i18n.T(i18n.KeyRemote), err)
		return
	}
	_, _ = tmp.Write(data)
	_ = tmp.Close()
	_ = fileopen.OpenPath(tmp.Name())
}

func (a *App) SaveEditor(path, text string, remotePane bool) error {
	data, err := textencoding.Encode(text, a.editorEnc)
	if err != nil {
		return err
	}
	if remotePane {
		client, ok := a.requireClient()
		if !ok {
			return errors.New("not connected")
		}
		return client.WriteFile(path, data)
	}
	return os.WriteFile(path, data, 0o644)
}

func (a *App) saveServers() error {
	return config.Save(a.store)
}

func (a *App) Startup() {
	a.host.refreshLocal()
	a.host.refreshRemote()
	a.host.updateStatusBar()
	a.host.refreshTabs()
}
