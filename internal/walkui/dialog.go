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
				setMultilineText(te, text)
			})
		})
	}

	if err := (Dialog{
		AssignTo: &dlg,
		Title:    title,
		MinSize:  Size{640, 480},
		Font:     uiFont(),
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			dlgBody(
				TextEdit{
					AssignTo: &te,
					ReadOnly: true,
					VScroll:  true,
					Font:     monoFont(),
				},
			),
			dlgFooter(
				PushButton{Text: i18n.T(i18n.KeyRefresh), OnClicked: refresh},
				PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
			),
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
		MinSize:  Size{360, 140},
		Font:     uiFont(),
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			dlgBody(
				Label{Text: label},
				LineEdit{AssignTo: &edit, Text: initial},
			),
			dlgFooter(
				PushButton{Text: i18n.T(i18n.KeyCancel), OnClicked: func() { dlg.Cancel() }},
				PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { ok = true; dlg.Accept() }},
			),
		},
	}.Run(a.mw)
	if err != nil || !ok {
		return "", false
	}
	return edit.Text(), true
}
