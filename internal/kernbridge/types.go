package kernbridge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DirEntryJSON struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"isDir"`
	ModTime  string `json:"modTime"`
	Mode     string `json:"mode,omitempty"`
	Owner    string `json:"owner,omitempty"`
	TypeName string `json:"typeName,omitempty"`
}

type TabJSON struct {
	Index       int    `json:"index"`
	Label       string `json:"label"`
	State       string `json:"state"`
	LocalPath   string `json:"localPath"`
	RemotePath  string `json:"remotePath"`
	HeartbeatSec int   `json:"heartbeatSec"`
}

type StatusJSON struct {
	Text           string `json:"text"`
	Connected      bool   `json:"connected"`
	Connecting     bool   `json:"connecting"`
	ShowReconnect  bool   `json:"showReconnect"`
	ConnDotColor   string `json:"connDotColor"`
	WindowTitle    string `json:"windowTitle"`
	HeartbeatLabel string `json:"heartbeatLabel,omitempty"`
}

type TransferJSON struct {
	Active   bool    `json:"active"`
	Progress float64 `json:"progress"`
	Speed    string  `json:"speed"`
	Queue    int     `json:"queue"`
	FileName string  `json:"fileName"`
}

type MessageJSON struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type PassphraseRequestJSON struct {
	RequestID int `json:"requestId"`
}

type EditorOpenJSON struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Text    string `json:"text"`
	Encoding string `json:"encoding"`
	Remote  bool   `json:"remote"`
	Size    int64  `json:"size"`
}

type FeatureTextJSON struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"json marshal failed"}`
	}
	return string(b)
}

func listLocalDirJSON(dir string) ([]DirEntryJSON, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	out := make([]DirEntryJSON, 0, len(names)+1)
	parent := filepath.Dir(strings.TrimSuffix(dir, `\`))
	if parent != "" && !strings.HasSuffix(parent, ":") {
		parent = ensureTrailingSlash(parent)
	}
	if dir != parent {
		out = append(out, DirEntryJSON{Name: "..", Path: parent, IsDir: true, TypeName: "\u4e0a\u7ea7\u76ee\u5f55"})
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if info.IsDir() {
			full = ensureTrailingSlash(full)
		}
		out = append(out, DirEntryJSON{
			Name:     name,
			Path:     full,
			Size:     info.Size(),
			IsDir:    info.IsDir(),
			ModTime:  info.ModTime().Format("2006/1/2 15:04:05"),
			TypeName: fileTypeName(full, info.IsDir()),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].nameIsParent() {
			return true
		}
		if out[j].nameIsParent() {
			return false
		}
		if out[i].IsDir != out[j].IsDir {
			return out[i].IsDir
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (e DirEntryJSON) nameIsParent() bool {
	return e.Name == ".."
}

func ensureTrailingSlash(path string) string {
	if path == "" {
		return path
	}
	if path[len(path)-1] != '\\' && path[len(path)-1] != '/' {
		return path + `\`
	}
	return path
}

func fileTypeName(path string, folder bool) string {
	if folder {
		return "\u6587\u4ef6\u5939"
	}
	ext := filepath.Ext(path)
	if ext == "" {
		return "\u6587\u4ef6"
	}
	return strings.TrimPrefix(ext, ".") + " \u6587\u4ef6"
}

func formatModTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006/1/2 15:04:05")
}

func formatMode(mode os.FileMode) string {
	if mode == 0 {
		return ""
	}
	return mode.String()
}
