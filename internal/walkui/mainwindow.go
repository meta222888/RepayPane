package walkui

import (
	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func tableColumns() []TableViewColumn {
	return []TableViewColumn{
		{Title: "Name", Width: 220},
		{Title: "Size", Width: 80},
		{Title: "Modified", Width: 140},
	}
}

func Run() error {
	store, err := config.Load()
	if err != nil {
		return err
	}
	settings, err := config.LoadSettings()
	if err == nil && settings.Language == "en" {
		i18n.SetLanguage(i18n.EN)
	}

	app := newApp(store)
	var mw *walk.MainWindow

	if err := (MainWindow{
		AssignTo: &mw,
		Title:    i18n.T(i18n.KeyAppTitle) + " (Win32)",
		MinSize:  Size{960, 600},
		Size:     Size{1280, 760},
		Layout:   VBox{MarginsZero: true},
		MenuItems: []MenuItem{
			Menu{
				Text: "&File",
				Items: []MenuItem{
					Action{
						Text:        "&Connect...",
						OnTriggered: app.showConnectDialog,
					},
					Separator{},
					Action{
						Text:        "E&xit",
						OnTriggered: func() { mw.Close() },
					},
				},
			},
			Menu{
				Text: "&Help",
				Items: []MenuItem{
					Action{
						Text: "About",
						OnTriggered: func() {
							walk.MsgBox(mw, "RelayPane Win32", "Native Win32 preview build (walk).\nNo OpenGL required.", walk.MsgBoxIconInformation)
						},
					},
				},
			},
		},
		Children: []Widget{
			ToolBar{
				Items: []MenuItem{
					Action{Text: i18n.T(i18n.KeyConnect), OnTriggered: app.showConnectDialog},
					Action{Text: i18n.T(i18n.KeyRefresh), OnTriggered: func() {
						app.refreshLocal()
						app.refreshRemote()
					}},
					Action{Text: i18n.T(i18n.KeyUpload), OnTriggered: app.uploadSelected},
					Action{Text: i18n.T(i18n.KeyDownload), OnTriggered: app.downloadSelected},
				},
			},
			HSplitter{
				Children: []Widget{
					paneComposite(app, true),
					paneComposite(app, false),
				},
			},
			Composite{
				Layout: HBox{Margins: Margins{6, 4, 6, 4}},
				Children: []Widget{
					Label{AssignTo: &app.statusLabel, Text: i18n.T(i18n.KeyNotConnected)},
					HSpacer{},
					PushButton{
						AssignTo:  &app.reconnectBtn,
						Text:      i18n.T(i18n.KeyReconnect),
						Visible:   false,
						OnClicked: func() { app.connectServer(app.server) },
					},
				},
			},
		},
	}).Create(); err != nil {
		return err
	}

	app.mw = mw
	app.refreshLocal()

	if len(store.Servers) > 0 {
		app.showConnectDialog()
	}

	mw.Run()
	return nil
}

func paneComposite(app *App, local bool) Widget {
	var title string
	if local {
		title = i18n.T(i18n.KeyLocal)
	} else {
		title = i18n.T(i18n.KeyRemote)
	}

	var pathEdit **walk.LineEdit
	var tv **walk.TableView
	var model *dirModel
	var upFn, refreshFn, activatedFn func()
	var onPathReturn func()

	if local {
		pathEdit = &app.localPathEdit
		tv = &app.localTV
		model = app.localModel
		upFn = app.localUp
		refreshFn = app.refreshLocal
		activatedFn = app.onLocalActivated
		onPathReturn = func() {
			app.navigateLocal(app.localPathEdit.Text())
		}
	} else {
		pathEdit = &app.remotePathEdit
		tv = &app.remoteTV
		model = app.remoteModel
		upFn = app.remoteUp
		refreshFn = app.refreshRemote
		activatedFn = app.onRemoteActivated
		onPathReturn = func() {
			app.navigateRemote(app.remotePathEdit.Text())
		}
	}

	return Composite{
		Layout: VBox{Margins: Margins{4, 4, 4, 4}},
		Children: []Widget{
			Label{Text: title, Font: Font{Bold: true}},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{Text: i18n.T(i18n.KeyUp), OnClicked: upFn, MaxSize: Size{48, 0}},
					PushButton{Text: i18n.T(i18n.KeyRefresh), OnClicked: refreshFn, MaxSize: Size{56, 0}},
					LineEdit{
						AssignTo: pathEdit,
						OnKeyDown: func(key walk.Key) {
							if key == walk.KeyReturn {
								onPathReturn()
							}
						},
					},
				},
			},
			TableView{
				AssignTo:         tv,
				AlternatingRowBG: true,
				Columns:          tableColumns(),
				Model:            model,
				OnItemActivated:  activatedFn,
			},
		},
	}
}
