package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window

	store        *config.Store
	client       *remote.Client
	activeServer *config.Server

	serverList *widget.List
	status     *widget.Label

	localPane  *FilePane
	remotePane *FilePane
}

func NewApp(a fyne.App, w fyne.Window) *App {
	store, err := config.Load()
	if err != nil {
		dialog.ShowError(err, w)
		store = &config.Store{}
	}

	appUI := &App{
		fyneApp: a,
		window:  w,
		store:   store,
		status:  widget.NewLabel("Not connected"),
	}

	appUI.localPane = NewLocalPane(appUI)
	appUI.remotePane = NewRemotePane(appUI)

	appUI.serverList = widget.NewList(
		func() int { return len(appUI.store.Servers) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(appUI.store.Servers[id].Name)
		},
	)
	appUI.serverList.OnSelected = func(id widget.ListItemID) {
		appUI.connectServer(&appUI.store.Servers[id])
	}

	sidebar := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Servers", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewButton("Add Server", appUI.showAddServer),
			widget.NewButton("Edit", appUI.showEditServer),
			widget.NewButton("Delete", appUI.showDeleteServer),
		),
		nil, nil, nil,
		appUI.serverList,
	)

	panes := container.NewHSplit(appUI.localPane.Container(), appUI.remotePane.Container())
	panes.SetOffset(0.5)

	transferBar := container.NewHBox(
		widget.NewButton("Upload  →", appUI.uploadSelectedLocal),
		widget.NewButton("←  Download", appUI.downloadSelectedRemote),
	)

	toolbar := container.NewHBox(
		widget.NewButton("Refresh", appUI.refreshPanes),
		widget.NewButton("Disconnect", appUI.disconnect),
		transferBar,
	)

	content := container.NewBorder(
		container.NewVBox(toolbar, appUI.status),
		nil, sidebar, nil,
		panes,
	)

	w.SetContent(content)
	w.SetOnDropped(appUI.onWindowDropped)
	return appUI
}

func (a *App) onWindowDropped(pos fyne.Position, uris []fyne.URI) {
	size := a.window.Content().Size()
	remoteArea := pos.X > size.Width/2
	a.handleDrop(pos, uris, remoteArea)
}
	if a.client != nil {
		_ = a.client.Close()
		a.client = nil
	}

	a.status.SetText(fmt.Sprintf("Connecting to %s…", s.Name))
	go func() {
		client, err := remote.Connect(remote.ConnectOptions{
			Host:       s.Host,
			Port:       s.Port,
			Username:   s.Username,
			Password:   s.Password,
			PrivateKey: s.PrivateKey,
		})
		fyne.CurrentApp().Driver().RunOnMain(func() {
			if err != nil {
				a.status.SetText("Connection failed")
				dialog.ShowError(err, a.window)
				return
			}
			a.client = client
			a.activeServer = s
			root := s.RemoteRoot
			if root == "" {
				root = "/"
			}
			a.remotePane.SetConnected(true)
			a.remotePane.Navigate(root)
			a.status.SetText(fmt.Sprintf("Connected: %s (%s@%s)", s.Name, s.Username, s.Host))
		})
	}()
}

func (a *App) disconnect() {
	if a.client != nil {
		_ = a.client.Close()
		a.client = nil
	}
	a.activeServer = nil
	a.remotePane.SetConnected(false)
	a.status.SetText("Disconnected")
}

func (a *App) refreshPanes() {
	a.localPane.RefreshListing()
	a.remotePane.RefreshListing()
}

func (a *App) connectServer(s *config.Server) {
	if a.client == nil {
		return
	}
	if entry.Size > config.MaxEditBytes {
		dialog.ShowConfirm(
			"File too large",
			fmt.Sprintf("%s is %.1f MB. Open anyway? (Ctrl+S saves back to server)", entry.Name, float64(entry.Size)/(1024*1024)),
			func(ok bool) {
				if ok {
					a.loadEditor(entry)
				}
			},
			a.window,
		)
		return
	}
	a.loadEditor(entry)
}

func (a *App) loadEditor(entry remote.FileInfo) {
	go func() {
		data, err := a.client.ReadFile(entry.Path)
		fyne.CurrentApp().Driver().RunOnMain(func() {
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			ShowEditor(a, entry, string(data))
		})
	}()
}

func (a *App) saveServers() {
	if err := config.Save(a.store); err != nil {
		dialog.ShowError(err, a.window)
	}
	a.serverList.Refresh()
}

func (a *App) showAddServer() {
	showServerForm(a, config.Server{Port: config.DefaultSFTPPort}, func(s config.Server) {
		s.ID = fmt.Sprintf("srv-%d", time.Now().UnixNano())
		a.store.Servers = append(a.store.Servers, s)
		a.saveServers()
	})
}

func (a *App) showEditServer() {
	id := a.serverList.Selected()
	if id < 0 || id >= len(a.store.Servers) {
		dialog.ShowInformation("Select server", "Choose a server to edit.", a.window)
		return
	}
	s := a.store.Servers[id]
	showServerForm(a, s, func(updated config.Server) {
		updated.ID = s.ID
		a.store.Servers[id] = updated
		a.saveServers()
	})
}

func (a *App) showDeleteServer() {
	id := a.serverList.Selected()
	if id < 0 || id >= len(a.store.Servers) {
		dialog.ShowInformation("Select server", "Choose a server to delete.", a.window)
		return
	}
	name := a.store.Servers[id].Name
	dialog.ShowConfirm("Delete server", fmt.Sprintf("Delete %q?", name), func(ok bool) {
		if !ok {
			return
		}
		a.store.Servers = append(a.store.Servers[:id], a.store.Servers[id+1:]...)
		a.saveServers()
	}, a.window)
}

func showServerForm(a *App, initial config.Server, onSave func(config.Server)) {
	name := widget.NewEntry()
	name.SetText(initial.Name)
	host := widget.NewEntry()
	host.SetText(initial.Host)
	port := widget.NewEntry()
	if initial.Port == 0 {
		initial.Port = config.DefaultSFTPPort
	}
	port.SetText(fmt.Sprintf("%d", initial.Port))
	user := widget.NewEntry()
	user.SetText(initial.Username)
	pass := widget.NewPasswordEntry()
	pass.SetText(initial.Password)
	keyPath := widget.NewEntry()
	keyPath.SetText(initial.PrivateKey)
	root := widget.NewEntry()
	root.SetText(initial.RemoteRoot)
	root.SetPlaceHolder("/")

	form := dialog.NewForm("Server", "Save", "Cancel", []*widget.FormItem{
		{Text: "Name", Widget: name},
		{Text: "Host", Widget: host},
		{Text: "Port", Widget: port},
		{Text: "Username", Widget: user},
		{Text: "Password", Widget: pass},
		{Text: "Private key file", Widget: keyPath},
		{Text: "Remote root", Widget: root},
	}, func(ok bool) {
		if !ok {
			return
		}
		var p int
		fmt.Sscanf(port.Text, "%d", &p)
		onSave(config.Server{
			Name:       strings.TrimSpace(name.Text),
			Host:       strings.TrimSpace(host.Text),
			Port:       p,
			Username:   strings.TrimSpace(user.Text),
			Password:   pass.Text,
			PrivateKey: strings.TrimSpace(keyPath.Text),
			RemoteRoot: strings.TrimSpace(root.Text),
		})
	}, a.window)
	form.Resize(fyne.NewSize(480, 420))
	form.Show()
}

// Local file listing helpers used by FilePane.

type localEntry struct {
	name  string
	path  string
	size  int64
	isDir bool
	mod   time.Time
}

func listLocal(dir string) ([]localEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]localEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, localEntry{
			name:  e.Name(),
			path:  filepath.Join(dir, e.Name()),
			size:  info.Size(),
			isDir: e.IsDir(),
			mod:   info.ModTime(),
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

func defaultLocalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "C:\\"
	}
	return home
}
