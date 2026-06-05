package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

type placeEntry struct {
	label string
	path  string
	short string
	icon  fyne.Resource
}

type LocalNav struct {
	pane     *FilePane
	driveBtn *localDriveButton
}

func NewLocalNav(pane *FilePane) *LocalNav {
	n := &LocalNav{pane: pane}
	n.driveBtn = newLocalDriveButton(n.showRootMenu)
	n.driveBtn.SetLabel(n.rootLabel())
	return n
}

func (n *LocalNav) Widget() fyne.CanvasObject {
	return newPaneFixedHeight(paneBandInnerHeight(), n.driveBtn)
}

func (n *LocalNav) ApplyLanguage() {
	n.driveBtn.SetLabel(n.rootLabel())
}

func (n *LocalNav) syncFromPath(path string) {
	_ = path
	n.driveBtn.SetLabel(n.rootLabel())
}

func (n *LocalNav) rootLabel() string {
	path := n.pane.path
	if len(path) >= 2 && path[1] == ':' {
		return strings.ToUpper(path[:2]) + `\`
	}
	return path
}

func (n *LocalNav) showRootMenu() {
	c := n.pane.app.window.Canvas()
	showLocalNavPopup(c, n.driveBtn, func(dismiss func()) fyne.CanvasObject {
		current := n.rootLabel()
		var rows []fyne.CanvasObject
		rows = append(rows, localNavSectionHeader(i18n.T(i18n.KeySidebarDrive)))
		for _, d := range listWindowsDrives() {
			drive := d
			active := strings.EqualFold(drive, current)
			rows = append(rows, newLocalNavMenuRow(theme.ComputerIcon(), drive, "", active, func() {
				dismiss()
				target := drive
				fyne.Do(func() { n.pane.Navigate(target) })
			}))
		}
		rows = append(rows, localNavPopupSeparator())
		rows = append(rows, localNavSectionHeader(i18n.T(i18n.KeySidebarPlaces)))
		for _, place := range commonPlaces() {
			p := place
			rows = append(rows, newLocalNavMenuRow(p.icon, p.label, p.short, false, func() {
				dismiss()
				target := p.path
				fyne.Do(func() { n.pane.Navigate(target) })
			}))
		}
		return container.NewVBox(rows...)
	})
}

func commonPlaces() []placeEntry {
	home, _ := os.UserHomeDir()
	type candidate struct {
		key      string
		sub      string
		fallback string
		icon     fyne.Resource
	}
	candidates := []candidate{
		{i18n.KeyPlaceDesktop, "Desktop", filepath.Join(home, "Desktop"), theme.ComputerIcon()},
		{i18n.KeyPlaceDocuments, "Documents", filepath.Join(home, "Documents"), theme.DocumentIcon()},
		{i18n.KeyPlacePictures, "Pictures", filepath.Join(home, "Pictures"), theme.MediaPhotoIcon()},
		{i18n.KeyPlaceDownloads, "Downloads", filepath.Join(home, "Downloads"), theme.DownloadIcon()},
		{i18n.KeyPlaceMusic, "Music", filepath.Join(home, "Music"), theme.MediaMusicIcon()},
		{i18n.KeyPlaceVideos, "Videos", filepath.Join(home, "Videos"), theme.MediaVideoIcon()},
		{i18n.KeyPlaceHome, "", home, theme.HomeIcon()},
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
		short := "~"
		if c.sub != "" {
			short = `~\` + c.sub
		}
		out = append(out, placeEntry{
			label: i18n.T(c.key),
			path:  p,
			short: short,
			icon:  c.icon,
		})
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
