package kernbridge

import (
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/relaypane/relaypane/internal/i18n"
)

const (
	EventRefreshLocal = iota + 1
	EventRefreshRemote
	EventStatus
	EventTransfer
	EventMessage
	EventPassphrase
	EventEditorOpen
	EventFeatureText
	EventTabs
)

var (
	hostMu       sync.Mutex
	hostHWND     uintptr
	hostCallback uintptr

	user32      = syscall.NewLazyDLL("user32.dll")
	postMessage = user32.NewProc("PostMessageW")
)

const wmAppEvent = 0x8000 + 42

func WMAppEvent() int { return wmAppEvent }

func SetUIHost(hwnd uintptr, callback uintptr) {
	hostMu.Lock()
	hostHWND = hwnd
	hostCallback = callback
	hostMu.Unlock()
}

func postEvent(eventType int, jsonPayload string) {
	hostMu.Lock()
	hwnd := hostHWND
	cb := hostCallback
	hostMu.Unlock()
	if hwnd == 0 {
		return
	}
	payload := syscall.StringToUTF16Ptr(jsonPayload)
	eventTypeU := uintptr(eventType)
	_, _, _ = postMessage.Call(hwnd, wmAppEvent, eventTypeU, uintptr(unsafe.Pointer(payload)))
	if cb != 0 {
		// optional direct callback for synchronous paths
	}
}

type uiHost struct {
	app *App
}

func (h *uiHost) sync(fn func()) {
	fn()
	postEvent(EventStatus, h.app.statusJSON())
}

func (h *uiHost) showError(title string, err error) {
	if err == nil {
		return
	}
	postEvent(EventMessage, mustJSON(MessageJSON{Kind: "error", Title: title, Text: err.Error()}))
}

func (h *uiHost) showMsg(title, msg string) {
	postEvent(EventMessage, mustJSON(MessageJSON{Kind: "info", Title: title, Text: msg}))
}

var passphraseReqID atomic.Int32

func (h *uiHost) askPassphrase() []byte {
	id := int(passphraseReqID.Add(1))
	postEvent(EventPassphrase, mustJSON(PassphraseRequestJSON{RequestID: id}))
	// C++ must call SubmitPassphrase before connect continues; blocking via channel
	return h.app.waitPassphrase(id)
}

func (h *uiHost) refreshLocal() {
	entries, err := listLocalDirJSON(h.app.localPath)
	if err != nil {
		h.showError("Local", err)
		return
	}
	if tab := h.app.activeSession(); tab != nil {
		tab.localPath = h.app.localPath
	}
	postEvent(EventRefreshLocal, mustJSON(map[string]any{
		"path":    h.app.localPath,
		"entries": entries,
	}))
}

func (h *uiHost) refreshRemote() {
	if !h.app.connected || h.app.client == nil {
		postEvent(EventRefreshRemote, mustJSON(map[string]any{
			"path":      h.app.remotePath,
			"entries":   []DirEntryJSON{},
			"loading":   false,
			"empty":     true,
			"connected": false,
		}))
		return
	}
	client := h.app.client
	dir := h.app.remotePath
	postEvent(EventRefreshRemote, mustJSON(map[string]any{
		"path":      dir,
		"entries":   nil,
		"loading":   true,
		"empty":     false,
		"connected": true,
	}))
	go func() {
		entries, err := listRemoteDirJSON(client, dir)
		h.sync(func() {
			if err != nil {
				h.showError("Remote", err)
				return
			}
			if tab := h.app.activeSession(); tab != nil {
				tab.remotePath = h.app.remotePath
			}
			postEvent(EventRefreshRemote, mustJSON(map[string]any{
				"path":      h.app.remotePath,
				"entries":   entries,
				"loading":   false,
				"empty":     false,
				"connected": true,
			}))
		})
	}()
}

func (h *uiHost) refreshTransferUI() {
	if h.app.transfers == nil {
		return
	}
	active, progress, speed, queue, fileName := h.app.transfers.Snapshot()
	postEvent(EventTransfer, mustJSON(TransferJSON{
		Active: active, Progress: progress, Speed: speed, Queue: queue, FileName: fileName,
	}))
}

func (h *uiHost) refreshTabs() {
	postEvent(EventTabs, h.app.tabsJSON())
}

func (h *uiHost) updateStatusBar() {
	postEvent(EventStatus, h.app.statusJSON())
}

func (h *uiHost) showFeatureDialog(title string, load func(set func(string))) {
	postEvent(EventFeatureText, mustJSON(FeatureTextJSON{Title: title, Text: i18n.T(i18n.KeyFeatLoading)}))
	go func() {
		load(func(text string) {
			postEvent(EventFeatureText, mustJSON(FeatureTextJSON{Title: title, Text: text}))
		})
	}()
}

func (h *uiHost) showEditor(title, path, text, enc string, remote bool, size int64) {
	postEvent(EventEditorOpen, mustJSON(EditorOpenJSON{
		Title: title, Path: path, Text: text, Encoding: enc, Remote: remote, Size: size,
	}))
}
