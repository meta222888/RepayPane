package walkui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/fileopen"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"
	"github.com/relaypane/relaypane/internal/textencoding"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (a *App) openRemoteEditor(entry remote.FileInfo) {
	if fileopen.IsImageName(entry.Name) {
		a.loadRemoteImage(entry)
		return
	}
	if entry.Size > config.MaxEditBytes {
		a.showConfirm(i18n.T(i18n.KeyFileTooLarge), i18n.Tf(i18n.KeyFileTooLargeMsg, entry.Name, float64(entry.Size)/(1024*1024)), func() {
			a.loadEditor(entry)
		})
		return
	}
	a.loadEditor(entry)
}

func (a *App) openLocalEditor(path, name string, size int64) {
	if fileopen.IsImageName(name) {
		if err := fileopen.OpenPath(path); err != nil {
			a.showError(i18n.T(i18n.KeyEditTitle), err)
		}
		return
	}
	if size > config.MaxEditBytes {
		a.showConfirm(i18n.T(i18n.KeyFileTooLarge), i18n.Tf(i18n.KeyFileTooLargeMsg, name, float64(size)/(1024*1024)), func() {
			a.loadLocalEditor(path, name)
		})
		return
	}
	a.loadLocalEditor(path, name)
}

func (a *App) loadLocalEditor(path, name string) {
	go func() {
		data, err := os.ReadFile(path)
		a.syncUI(func() {
			if err != nil {
				a.showError(i18n.T(i18n.KeyEditTitle), err)
				return
			}
			if fileopen.IsImageData(data) {
				_ = fileopen.OpenPath(path)
				return
			}
			if !fileopen.IsLikelyText(data) {
				a.showMsg(i18n.T(i18n.KeyNotTextFileTitle), i18n.Tf(i18n.KeyNotTextFileMsg, name))
				return
			}
			text, enc, err := textencoding.Decode(data)
			if err != nil {
				a.showError(i18n.T(i18n.KeyEditTitle), err)
				return
			}
			a.showTextEditor(i18n.Tf(i18n.KeyEditTitle, name), path, text, enc,
				func(raw []byte) error { return os.WriteFile(path, raw, 0o644) },
				func() ([]byte, error) { return os.ReadFile(path) },
			)
		})
	}()
}

func (a *App) loadEditor(entry remote.FileInfo) {
	client := a.activeClient()
	if client == nil {
		return
	}
	go func() {
		data, err := client.ReadFile(entry.Path)
		a.syncUI(func() {
			if err != nil {
				a.showError(i18n.T(i18n.KeyEditTitle), err)
				return
			}
			if fileopen.IsImageData(data) {
				a.openRemoteImageData(entry.Name, data)
				return
			}
			if !fileopen.IsLikelyText(data) {
				a.showMsg(i18n.T(i18n.KeyNotTextFileTitle), i18n.Tf(i18n.KeyNotTextFileMsg, entry.Name))
				return
			}
			text, enc, err := textencoding.Decode(data)
			if err != nil {
				a.showError(i18n.T(i18n.KeyEditTitle), err)
				return
			}
			path := entry.Path
			a.showTextEditor(i18n.Tf(i18n.KeyEditTitle, entry.Name), path, text, enc,
				func(raw []byte) error {
					c := a.activeClient()
					if c == nil {
						return fmt.Errorf(i18n.T(i18n.KeyNotConnectedErr))
					}
					return c.WriteFile(path, raw)
				},
				func() ([]byte, error) {
					c := a.activeClient()
					if c == nil {
						return nil, fmt.Errorf(i18n.T(i18n.KeyNotConnectedErr))
					}
					return c.ReadFile(path)
				},
			)
		})
	}()
}

func (a *App) loadRemoteImage(entry remote.FileInfo) {
	client := a.activeClient()
	if client == nil {
		return
	}
	go func() {
		data, err := client.ReadFile(entry.Path)
		a.syncUI(func() {
			if err != nil {
				a.showError(i18n.T(i18n.KeyEditTitle), err)
				return
			}
			a.openRemoteImageData(entry.Name, data)
		})
	}()
}

func (a *App) openRemoteImageData(name string, data []byte) {
	ext := filepathExt(name)
	tmp, err := os.CreateTemp("", "relaypane-view-*"+ext)
	if err != nil {
		a.showError(i18n.T(i18n.KeyEditTitle), err)
		return
	}
	path := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(path)
		a.showError(i18n.T(i18n.KeyEditTitle), err)
		return
	}
	_ = tmp.Close()
	if err := fileopen.OpenPath(path); err != nil {
		os.Remove(path)
		a.showError(i18n.T(i18n.KeyEditTitle), err)
	}
}

func filepathExt(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i:]
	}
	return ".img"
}

func (a *App) showTextEditor(title, path, text string, enc textencoding.Info, saveFn func([]byte) error, loadFn func() ([]byte, error)) {
	var mw *walk.MainWindow
	var te *walk.TextEdit
	var status *walk.Label
	dirty := false

	save := func() {
		data, err := textencoding.Encode(te.Text(), enc)
		if err != nil {
			a.showError(i18n.T(i18n.KeySave), err)
			return
		}
		if err := saveFn(data); err != nil {
			a.showError(i18n.T(i18n.KeySave), err)
			return
		}
		dirty = false
		status.SetText(i18n.Tf(i18n.KeySaveSuccessAt, time.Now().Format("2006-01-02 15:04:05")))
	}

	revert := func() {
		go func() {
			raw, err := loadFn()
			a.syncUI(func() {
				if err != nil {
					a.showError(i18n.T(i18n.KeyEditorRevert), err)
					return
				}
				t, newEnc, err := textencoding.Decode(raw)
				if err != nil {
					a.showError(i18n.T(i18n.KeyEditorRevert), err)
					return
				}
				enc = newEnc
				te.SetText(t)
				dirty = false
				status.SetText(i18n.Tf(i18n.KeyEditorEncoding, enc.Label()))
			})
		}()
	}

	_ = MainWindow{
		AssignTo: &mw,
		Title:    title,
		MinSize:  Size{720, 520},
		Layout:   VBox{},
		MenuItems: []MenuItem{
			Menu{
				Text: "File",
				Items: []MenuItem{
					Action{Text: i18n.T(i18n.KeySave), Shortcut: Shortcut{Modifiers: walk.ModControl, Key: walk.KeyS}, OnTriggered: save},
					Action{Text: i18n.T(i18n.KeyEditorRevert), OnTriggered: revert},
				},
			},
		},
		Children: []Widget{
			Label{Text: path},
			TextEdit{
				AssignTo: &te,
				Text:     text,
				VScroll:  true,
				Font:     Font{Family: "Consolas", PointSize: 10},
				OnTextChanged: func() { dirty = true },
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{AssignTo: &status, Text: i18n.Tf(i18n.KeyEditorEncoding, enc.Label())},
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyEditorRevert), OnClicked: revert},
					PushButton{Text: i18n.T(i18n.KeySave), OnClicked: save},
				},
			},
		},
	}.Create()

	a.applyWindowIcon(mw)
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if dirty {
			*canceled = walk.MsgBox(mw, i18n.T(i18n.KeyUnsaved), i18n.T(i18n.KeyDiscard), walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes
		}
	})
	mw.Show()
}
