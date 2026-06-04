package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type placeEntry struct {
	labelKey string
	path     string
}

type LocalSidebar struct {
	pane   *FilePane
	places []placeEntry
	list   *widget.List
	drive  *widget.Select
	root   *fyne.Container
}

func NewLocalSidebar(pane *FilePane) *LocalSidebar {
	s := &LocalSidebar{pane: pane}
	s.rebuildPlaces()
	s.list = widget.NewList(
		func() int { return len(s.places) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if int(id) >= len(s.places) {
				return
			}
			obj.(*widget.Label).SetText(s.places[id].labelKey)
		},
	)
	s.list.OnSelected = func(id widget.ListItemID) {
		if int(id) < len(s.places) {
			s.pane.Navigate(s.places[id].path)
		}
	}

	drives := listWindowsDrives()
	s.drive = widget.NewSelect(drives, func(path string) {
		if path != "" {
			s.pane.Navigate(path)
		}
	})
	if len(drives) > 0 {
		s.drive.SetSelected(drives[0])
	}

	placesTitle := widget.NewLabelWithStyle(i18n.T(i18n.KeySidebarPlaces), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	driveTitle := widget.NewLabelWithStyle(i18n.T(i18n.KeySidebarDrive), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	bg := canvas.NewRectangle(colorPanel)
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(1, 0))

	content := container.NewVBox(
		container.NewPadded(placesTitle),
		s.list,
		container.NewPadded(driveTitle),
		container.NewPadded(s.drive),
	)
	s.root = container.NewBorder(nil, nil, nil, line, container.NewStack(bg, content))
	s.syncFromPath(pane.path)
	return s
}

func (s *LocalSidebar) Container() fyne.CanvasObject { return s.root }

func (s *LocalSidebar) ApplyLanguage() {
	s.rebuildPlaces()
	s.list.Refresh()
}

func (s *LocalSidebar) syncFromPath(path string) {
	if len(path) >= 2 && path[1] == ':' {
		root := strings.ToUpper(path[:2]) + `\`
		for _, d := range s.drive.Options {
			if strings.EqualFold(d, root) {
				s.drive.SetSelected(d)
				return
			}
		}
	}
}

func (s *LocalSidebar) rebuildPlaces() {
	home, _ := os.UserHomeDir()
	type candidate struct {
		key  string
		join string
		fallback string
	}
	candidates := []candidate{
		{i18n.KeyPlaceHome, "", home},
		{i18n.KeyPlaceDesktop, "Desktop", filepath.Join(home, "Desktop")},
		{i18n.KeyPlaceDocuments, "Documents", filepath.Join(home, "Documents")},
		{i18n.KeyPlaceDownloads, "Downloads", filepath.Join(home, "Downloads")},
	}
	s.places = s.places[:0]
	for _, c := range candidates {
		p := c.fallback
		if c.join == "" {
			p = home
		}
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			s.places = append(s.places, placeEntry{labelKey: i18n.T(c.key), path: p})
		}
	}
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
