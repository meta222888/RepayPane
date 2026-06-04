package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type PaneKind int

const (
	PaneLocal PaneKind = iota
	PaneRemote
)

type FilePane struct {
	app  *App
	kind PaneKind

	path    string
	local   []localEntry
	remote  []remote.FileInfo
	connected bool

	title   *widget.Label
	pathBar *widget.Entry
	list    *widget.List
	root    *fyne.Container

	selectedID int
	lastTap     time.Time
	lastTapID   int
}

func NewLocalPane(app *App) *FilePane {
	p := &FilePane{app: app, kind: PaneLocal, path: defaultLocalDir(), selectedID: -1}
	p.build()
	p.RefreshListing()
	return p
}

func NewRemotePane(app *App) *FilePane {
	p := &FilePane{app: app, kind: PaneRemote, path: "/", selectedID: -1}
	p.build()
	return p
}

func (p *FilePane) build() {
	p.title = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	p.pathBar = widget.NewEntry()
	p.pathBar.OnSubmitted = func(s string) { p.Navigate(s) }

	p.list = widget.NewList(
		func() int {
			if p.kind == PaneLocal {
				return len(p.local)
			}
			return len(p.remote)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("icon"), widget.NewLabel("name"))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			box := obj.(*fyne.Container)
			icon := box.Objects[0].(*widget.Label)
			name := box.Objects[1].(*widget.Label)
			if p.kind == PaneLocal {
				e := p.local[id]
				icon.SetText(entryIcon(e.isDir))
				name.SetText(formatEntry(e.name, e.size, e.isDir))
			} else {
				e := p.remote[id]
				icon.SetText(entryIcon(e.IsDir))
				name.SetText(formatEntry(e.Name, e.Size, e.IsDir))
			}
		},
	)

	p.list.OnSelected = func(id widget.ListItemID) {
		p.selectedID = int(id)
		p.onSelect(id)
	}

	upBtn := widget.NewButton("Up", func() { p.goUp() })
	refreshBtn := widget.NewButton("Refresh", func() { p.RefreshListing() })

	header := container.NewBorder(nil, nil, p.title, container.NewHBox(upBtn, refreshBtn), p.pathBar)
	p.root = container.NewBorder(header, nil, nil, nil, p.list)

	if p.kind == PaneLocal {
		p.title.SetText("Local")
	} else {
		p.title.SetText("Remote")
	}
}

func (p *FilePane) Container() *fyne.Container {
	return p.root
}

func (p *FilePane) CurrentPath() string {
	return p.path
}

func (p *FilePane) SetConnected(v bool) {
	p.connected = v
	if !v {
		p.remote = nil
		p.list.Refresh()
	}
}

func (p *FilePane) Navigate(path string) {
	if p.kind == PaneLocal {
		path = filepath.Clean(path)
		st, err := os.Stat(path)
		if err != nil || !st.IsDir() {
			dialog.ShowError(fmt.Errorf("invalid local path: %s", path), p.app.window)
			return
		}
	} else {
		if !p.connected || p.app.client == nil {
			return
		}
		path = strings.ReplaceAll(path, "\\", "/")
		if path == "" {
			path = "/"
		}
	}
	p.path = path
	p.pathBar.SetText(path)
	p.RefreshListing()
}

func (p *FilePane) goUp() {
	if p.kind == PaneLocal {
		parent := filepath.Dir(p.path)
		if parent == p.path {
			return
		}
		p.Navigate(parent)
		return
	}
	if p.path == "/" {
		return
	}
	parent := filepath.Dir(strings.ReplaceAll(p.path, "\\", "/"))
	p.Navigate(parent)
}

func (p *FilePane) RefreshListing() {
	if p.kind == PaneLocal {
		entries, err := listLocal(p.path)
		if err != nil {
			dialog.ShowError(err, p.app.window)
			return
		}
		p.local = entries
		p.list.Refresh()
		return
	}
	if !p.connected || p.app.client == nil {
		return
	}
	entries, err := p.app.client.ListDir(p.path)
	if err != nil {
		dialog.ShowError(err, p.app.window)
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	p.remote = entries
	p.list.Refresh()
}

func (p *FilePane) onSelect(id widget.ListItemID) {
	now := time.Now()
	isDouble := int(id) == p.lastTapID && now.Sub(p.lastTap) < 450*time.Millisecond
	p.lastTap = now
	p.lastTapID = int(id)

	if !isDouble {
		return
	}

	if p.kind == PaneLocal {
		e := p.local[id]
		if e.isDir {
			p.Navigate(e.path)
		}
		return
	}

	e := p.remote[id]
	if e.IsDir {
		p.Navigate(e.Path)
		return
	}
	p.app.openRemoteEditor(e)
}

func entryIcon(isDir bool) string {
	if isDir {
		return "[D]"
	}
	return "[F]"
}

func formatEntry(name string, size int64, isDir bool) string {
	if isDir {
		return name + "/"
	}
	if size < 1024 {
		return fmt.Sprintf("%s  (%d B)", name, size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%s  (%.1f KB)", name, float64(size)/1024)
	}
	return fmt.Sprintf("%s  (%.1f MB)", name, float64(size)/(1024*1024))
}

func (p *FilePane) SelectedPath() string {
	if p.kind != PaneLocal || p.selectedID < 0 || p.selectedID >= len(p.local) {
		return ""
	}
	e := p.local[p.selectedID]
	if e.isDir {
		return ""
	}
	return e.path
}

func (p *FilePane) SelectedEntry() *remote.FileInfo {
	if p.kind != PaneRemote || p.selectedID < 0 || p.selectedID >= len(p.remote) {
		return nil
	}
	e := p.remote[p.selectedID]
	return &e
}
