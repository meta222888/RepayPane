package walkui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (a *App) ownDialog(dlg *walk.Dialog) {
	if a.mw == nil || dlg == nil {
		return
	}
	setWindowOwner(dlg.Handle(), a.mw.Handle())
}

func (a *App) showMsg(title, msg string) {
	a.syncUI(func() {
		walk.MsgBox(a.mw, title, msg, walk.MsgBoxIconInformation)
	})
}

func (a *App) showConfirm(title, msg string, onOK func()) {
	a.syncUI(func() {
		if walk.MsgBox(a.mw, title, msg, walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdYes {
			if onOK != nil {
				onOK()
			}
		}
	})
}

func (a *App) showConfirmSync(title, msg string) bool {
	if a.mw == nil {
		return false
	}
	return walk.MsgBox(a.mw, title, msg, walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdYes
}

func (a *App) showFeatureDialog(title string, load func(setText func(string))) {
	var dlg *walk.Dialog
	var te *walk.TextEdit

	refresh := func() {
		if load == nil {
			return
		}
		load(func(text string) {
			a.syncUI(func() {
				if te != nil {
					te.SetText(text)
				}
			})
		})
	}

	if err := (Dialog{
		AssignTo: &dlg,
		Title:    title,
		MinSize:  Size{640, 480},
		Layout:   VBox{},
		Children: []Widget{
			TextEdit{
				AssignTo: &te,
				ReadOnly: true,
				VScroll:  true,
				Font:     Font{Family: "Consolas", PointSize: 9},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyRefresh), OnClicked: refresh},
					PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}).Create(a.mw); err != nil {
		return
	}
	a.ownDialog(dlg)
	te.SetText(i18n.T(i18n.KeyFeatLoading))
	go refresh()
	dlg.Run()
}

func (a *App) promptInput(title, label, initial string) (string, bool) {
	var dlg *walk.Dialog
	var edit *walk.LineEdit
	ok := false

	_, err := Dialog{
		AssignTo: &dlg,
		Title:    title,
		MinSize:  Size{360, 120},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: label},
			LineEdit{AssignTo: &edit, Text: initial},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: "OK", OnClicked: func() { ok = true; dlg.Accept() }},
					PushButton{Text: "Cancel", OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}.Run(a.mw)
	if err != nil || !ok {
		return "", false
	}
	return edit.Text(), true
}
