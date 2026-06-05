package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type placeEntry struct {
	label string
	path  string
}

type LocalNav struct {
	pane    *FilePane
	rootBtn *widget.Button
}

func NewLocalNav(pane *FilePane) *LocalNav {
	n := &LocalNav{pane: pane}
	n.rootBtn = widget.NewButtonWithIcon(n.rootLabel(), theme.ComputerIcon(), func() {
		n.showRootMenu()
	})
	n.rootBtn.Importance = widget.LowImportance
	return n
}

func (n *LocalNav) Button() *widget.Button { return n.rootBtn }

func (n *LocalNav) ApplyLanguage() {
	n.rootBtn.SetText(n.rootLabel())
}

func (n *LocalNav) syncFromPath(path string) {
	_ = path
	n.rootBtn.SetText(n.rootLabel())
}

func (n *LocalNav) rootLabel() string {
	path := n.pane.path
	if len(path) >= 2 && path[1] == ':' {
		return strings.ToUpper(path[:2]) + `\`
	}
	return path
}

func (n *LocalNav) showRootMenu() {
	var items []*fyne.MenuItem
	for _, d := range listWindowsDrives() {
		drive := d
		items = append(items, fyne.NewMenuItem(drive, func() { n.pane.Navigate(drive) }))
	}
	items = append(items, fyne.NewMenuItemSeparator())
	for _, place := range commonPlaces() {
		p := place
		items = append(items, fyne.NewMenuItem(p.label, func() { n.pane.Navigate(p.path) }))
	}
	menu := fyne.NewMenu("", items...)
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(n.rootBtn)
	showWidePopUpMenu(n.pane.app.window.Canvas(), menu, pos.Add(fyne.NewPos(0, n.rootBtn.MinSize().Height)))
}

func commonPlaces() []placeEntry {
	home, _ := os.UserHomeDir()
	type candidate struct {
		key      string
		sub      string
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
		label := i18n.T(c.key)
		if c.sub != "" {
			label += "  ~\\" + c.sub
		} else {
			label += "  ~"
		}
		out = append(out, placeEntry{label: label, path: p})
	}
	return out
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
