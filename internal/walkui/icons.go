package walkui

import (
	"github.com/relaypane/relaypane/internal/assets"

	"github.com/lxn/walk"
)

type paneFileIcons struct{}

var fileIcons paneFileIcons

func initFileIcons() {}

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
			style.Image = iconPathForEntry(e, true)
			return
		}
		style.Image = iconPathForEntry(e, false)
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
		style.Image = shellIconPath(r.name, r.isDir)
	}
}
