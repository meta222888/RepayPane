package walkui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (a *App) resolveFileConflict(fileName string, exists func(string) bool, onProceed func(string)) {
	a.syncUI(func() {
		var dlg *walk.Dialog
		msg := i18n.Tf(i18n.KeyFileExistsConflict, fileName)

		makeBtn := func(text string, fn func()) Widget {
			return PushButton{Text: text, OnClicked: func() {
				dlg.Cancel()
				fn()
			}}
		}

		if err := (Dialog{
			AssignTo: &dlg,
			Title:    i18n.T(i18n.KeyFileExistsTitle),
			MinSize:  Size{420, 160},
			Layout:   VBox{},
			Children: []Widget{
				Label{Text: msg},
				Composite{
					Layout: HBox{},
					Children: []Widget{
						makeBtn(i18n.T(i18n.KeyCancel), func() { onProceed("") }),
						HSpacer{},
						makeBtn(i18n.T(i18n.KeySkip), func() { onProceed("") }),
						makeBtn(i18n.T(i18n.KeyRename), func() {
							name, ok := a.promptInput(i18n.T(i18n.KeyRename), i18n.T(i18n.KeyRenamePrompt), suggestCopyName(fileName, exists))
							if !ok || name == "" {
								onProceed("")
								return
							}
							if exists(name) {
								a.resolveFileConflict(name, exists, onProceed)
								return
							}
							onProceed(name)
						}),
						makeBtn(i18n.T(i18n.KeyOverwrite), func() { onProceed(fileName) }),
					},
				},
			},
		}).Create(a.mw); err != nil {
			onProceed("")
			return
		}
		dlg.Run()
	})
}
