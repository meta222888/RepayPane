package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"strings"
	"unsafe"

	"github.com/relaypane/relaypane/internal/kernbridge"
)

//export rp_free
func rp_free(p *C.char) {
	if p != nil {
		C.free(unsafe.Pointer(p))
	}
}

func cstr(s string) *C.char {
	return C.CString(s)
}

//export rp_init
func rp_init(hwnd C.uintptr_t) *C.char {
	if err := kernbridge.InitApp(); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	kernbridge.SetUIHost(uintptr(hwnd), 0)
	kernbridge.AppInstance().Startup()
	return cstr(`{"ok":true}`)
}

//export rp_set_ui_host
func rp_set_ui_host(hwnd C.uintptr_t) {
	kernbridge.SetUIHost(uintptr(hwnd), 0)
}

//export rp_get_status
func rp_get_status() *C.char {
	return cstr(kernbridge.AppInstance().StatusJSON())
}

//export rp_get_tabs
func rp_get_tabs() *C.char {
	return cstr(kernbridge.AppInstance().TabsJSON())
}

//export rp_get_servers
func rp_get_servers() *C.char {
	return cstr(kernbridge.AppInstance().ServersJSON())
}

//export rp_get_settings
func rp_get_settings() *C.char {
	return cstr(kernbridge.AppInstance().SettingsJSON())
}

//export rp_save_settings
func rp_save_settings(jsonData *C.char) *C.char {
	if err := kernbridge.AppInstance().SaveSettingsJSON(C.GoString(jsonData)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	return cstr(`{"ok":true}`)
}

//export rp_set_language
func rp_set_language(lang *C.char) {
	kernbridge.AppInstance().SetLanguage(C.GoString(lang))
}

//export rp_open_server_tab
func rp_open_server_tab(index C.int) {
	kernbridge.AppInstance().OpenServerTab(int(index))
}

//export rp_activate_tab
func rp_activate_tab(index C.int) {
	kernbridge.AppInstance().ActivateTab(int(index))
}

//export rp_close_tab
func rp_close_tab(index C.int) {
	kernbridge.AppInstance().CloseTab(int(index))
}

//export rp_reconnect
func rp_reconnect() {
	kernbridge.AppInstance().ReconnectActiveTab()
}

//export rp_navigate_local
func rp_navigate_local(path *C.char) {
	kernbridge.AppInstance().NavigateLocal(C.GoString(path))
}

//export rp_navigate_remote
func rp_navigate_remote(path *C.char) {
	kernbridge.AppInstance().NavigateRemote(C.GoString(path))
}

//export rp_local_up
func rp_local_up() {
	kernbridge.AppInstance().LocalUp()
}

//export rp_remote_up
func rp_remote_up() {
	kernbridge.AppInstance().RemoteUp()
}

//export rp_refresh
func rp_refresh() {
	kernbridge.AppInstance().RefreshBoth()
}

//export rp_upload_paths
func rp_upload_paths(jsonPaths *C.char) {
	var paths []string
	_ = json.Unmarshal([]byte(C.GoString(jsonPaths)), &paths)
	kernbridge.AppInstance().UploadPaths(paths)
}

//export rp_download_paths
func rp_download_paths(jsonPaths *C.char) {
	var paths []string
	_ = json.Unmarshal([]byte(C.GoString(jsonPaths)), &paths)
	kernbridge.AppInstance().DownloadPaths(paths)
}

//export rp_delete_local
func rp_delete_local(jsonPaths *C.char) {
	var paths []string
	_ = json.Unmarshal([]byte(C.GoString(jsonPaths)), &paths)
	kernbridge.AppInstance().DeleteLocal(paths)
}

//export rp_delete_remote
func rp_delete_remote(jsonPaths *C.char) {
	var paths []string
	_ = json.Unmarshal([]byte(C.GoString(jsonPaths)), &paths)
	kernbridge.AppInstance().DeleteRemote(paths)
}

//export rp_rename_local
func rp_rename_local(oldPath *C.char, newName *C.char) *C.char {
	if err := kernbridge.AppInstance().RenameLocal(C.GoString(oldPath), C.GoString(newName)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	kernbridge.AppInstance().RefreshBoth()
	return cstr(`{"ok":true}`)
}

//export rp_rename_remote
func rp_rename_remote(oldPath *C.char, newName *C.char) *C.char {
	if err := kernbridge.AppInstance().RenameRemote(C.GoString(oldPath), C.GoString(newName)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	kernbridge.AppInstance().RefreshBoth()
	return cstr(`{"ok":true}`)
}

//export rp_add_server
func rp_add_server(jsonServer *C.char) *C.char {
	if err := kernbridge.AppInstance().AddServerJSON(C.GoString(jsonServer)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	return cstr(`{"ok":true}`)
}

//export rp_update_server
func rp_update_server(index C.int, jsonServer *C.char) *C.char {
	if err := kernbridge.AppInstance().UpdateServerJSON(int(index), C.GoString(jsonServer)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	return cstr(`{"ok":true}`)
}

//export rp_delete_server
func rp_delete_server(index C.int) *C.char {
	if err := kernbridge.AppInstance().DeleteServer(int(index)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	return cstr(`{"ok":true}`)
}

//export rp_submit_passphrase
func rp_submit_passphrase(id C.int, pass *C.char) {
	kernbridge.AppInstance().SubmitPassphrase(int(id), C.GoString(pass))
}

//export rp_open_editor
func rp_open_editor(path *C.char, remote C.int) {
	kernbridge.AppInstance().OpenEditor(C.GoString(path), remote != 0)
}

//export rp_save_editor
func rp_save_editor(path *C.char, text *C.char, remote C.int) *C.char {
	if err := kernbridge.AppInstance().SaveEditor(C.GoString(path), C.GoString(text), remote != 0); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	return cstr(`{"ok":true}`)
}

//export rp_sync_upload
func rp_sync_upload() { kernbridge.AppInstance().SyncUpload() }

//export rp_sync_download
func rp_sync_download() { kernbridge.AppInstance().SyncDownload() }

//export rp_run_shell
func rp_run_shell(cmd *C.char) *C.char {
	return cstr(kernbridge.AppInstance().RunShell(C.GoString(cmd)))
}

//export rp_show_system_info
func rp_show_system_info() { kernbridge.AppInstance().ShowSystemInfo() }

//export rp_show_disk_space
func rp_show_disk_space() { kernbridge.AppInstance().ShowDiskSpace() }

//export rp_show_network_info
func rp_show_network_info() { kernbridge.AppInstance().ShowNetworkInfo() }

//export rp_show_resource_usage
func rp_show_resource_usage() { kernbridge.AppInstance().ShowResourceUsage() }

//export rp_show_du_tree
func rp_show_du_tree(dir *C.char) *C.char {
	return cstr(kernbridge.AppInstance().ShowDuTree(C.GoString(dir)))
}

//export rp_check_update
func rp_check_update() *C.char {
	return cstr(kernbridge.AppInstance().CheckUpdate())
}

//export rp_cloud_sync_status
func rp_cloud_sync_status() *C.char {
	return cstr(kernbridge.AppInstance().CloudSyncStatus())
}

//export rp_cloud_sync_upload
func rp_cloud_sync_upload() *C.char {
	return cstr(kernbridge.AppInstance().CloudSyncUpload())
}

//export rp_cloud_sync_download
func rp_cloud_sync_download() *C.char {
	return cstr(kernbridge.AppInstance().CloudSyncDownload())
}

//export rp_cloud_sync_delete
func rp_cloud_sync_delete() *C.char {
	return cstr(kernbridge.AppInstance().CloudSyncDelete())
}

//export rp_list_drives
func rp_list_drives() *C.char {
	return cstr(kernbridge.AppInstance().ListDrivesJSON())
}

//export rp_set_local_drive
func rp_set_local_drive(drive *C.char) {
	kernbridge.AppInstance().SetLocalDrive(C.GoString(drive))
}

//export rp_copy_clipboard
func rp_copy_clipboard(local C.int, jsonItems *C.char) {
	var items []struct {
		Path  string `json:"path"`
		Name  string `json:"name"`
		IsDir bool   `json:"isDir"`
	}
	_ = json.Unmarshal([]byte(C.GoString(jsonItems)), &items)
	paths := make([]string, len(items))
	names := make([]string, len(items))
	dirs := make([]bool, len(items))
	for i, it := range items {
		paths[i], names[i], dirs[i] = it.Path, it.Name, it.IsDir
	}
	kernbridge.AppInstance().CopyClipboard(local != 0, paths, names, dirs)
}

//export rp_paste_local
func rp_paste_local() { kernbridge.AppInstance().PasteToLocal() }

//export rp_paste_remote
func rp_paste_remote() { kernbridge.AppInstance().PasteToRemote() }

//export rp_new_folder_local
func rp_new_folder_local(name *C.char) *C.char {
	if err := kernbridge.AppInstance().NewFolderLocal(C.GoString(name)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	kernbridge.AppInstance().RefreshBoth()
	return cstr(`{"ok":true}`)
}

//export rp_new_folder_remote
func rp_new_folder_remote(name *C.char) *C.char {
	if err := kernbridge.AppInstance().NewFolderRemote(C.GoString(name)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	kernbridge.AppInstance().RefreshBoth()
	return cstr(`{"ok":true}`)
}

//export rp_new_file_local
func rp_new_file_local(name *C.char) *C.char {
	if err := kernbridge.AppInstance().NewFileLocal(C.GoString(name)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	kernbridge.AppInstance().RefreshBoth()
	return cstr(`{"ok":true}`)
}

//export rp_new_file_remote
func rp_new_file_remote(name *C.char) *C.char {
	if err := kernbridge.AppInstance().NewFileRemote(C.GoString(name)); err != nil {
		return cstr(`{"error":"` + err.Error() + `"}`)
	}
	kernbridge.AppInstance().RefreshBoth()
	return cstr(`{"ok":true}`)
}

//export rp_get_shell_history
func rp_get_shell_history() *C.char {
	return cstr(kernbridge.AppInstance().GetShellHistoryJSON())
}

//export rp_clear_shell_history
func rp_clear_shell_history() {
	kernbridge.AppInstance().ClearShellHistory()
}

//export rp_about
func rp_about() *C.char {
	return cstr(kernbridge.AppInstance().AboutJSON())
}

//export rp_wm_app_event
func rp_wm_app_event() C.int {
	return C.int(kernbridge.WMAppEvent())
}

func init() {
	_ = strings.Builder{}
}

func main() {}
