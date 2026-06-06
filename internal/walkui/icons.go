package walkui

import (
	"os"
	"path/filepath"

	"github.com/relaypane/relaypane/internal/assets"

	"github.com/lxn/walk"
)

type paneFileIcons struct {
	remoteFolder string
	remoteFile   string
}

var fileIcons paneFileIcons

func initFileIcons() {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = `C:\`
	}
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		winDir = `C:\Windows`
	}
	fileIcons.remoteFolder = home
	fileIcons.remoteFile = filepath.Join(winDir, "notepad.exe")
}

func (a *App) ensureAppIcon() *walk.Icon {
	if a.appIcon != nil {
		return a.appIcon
	}
	dpi := 96
	if a.mw != nil {
		dpi = a.mw.DPI()
	}
	icon, err := assets.WalkLogoIcon(dpi)
	if err != nil {
		return nil
	}
	a.appIcon = icon
	return icon
}

func (a *App) applyWindowIcon(form walk.Form) {
	if form == nil {
		return
	}
	if icon := a.ensureAppIcon(); icon != nil {
		_ = form.SetIcon(icon)
	}
}

func paneStyleCell(local bool, model *dirModel) func(style *walk.CellStyle) {
	return func(style *walk.CellStyle) {
		if style.Col() != 0 {
			return
		}
		e, ok := model.entry(style.Row())
		if !ok {
			return
		}
		if local {
			style.Image = e.fullPath
			return
		}
		if e.isDir {
			style.Image = fileIcons.remoteFolder
		} else {
			style.Image = fileIcons.remoteFile
		}
	}
}

func duStyleCell(model *duTableModel) func(style *walk.CellStyle) {
	return func(style *walk.CellStyle) {
		if style.Col() != 0 {
			return
		}
		if style.Row() < 0 || style.Row() >= len(model.rows) {
			return
		}
		r := model.rows[style.Row()]
		if r.isDir {
			style.Image = fileIcons.remoteFolder
		} else {
			style.Image = fileIcons.remoteFile
		}
	}
}
