package fileopen

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"
)

var imageExts = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".bmp": {},
	".webp": {}, ".ico": {}, ".tif": {}, ".tiff": {}, ".svg": {},
}

// IsImageName reports whether the file name has a common image extension.
func IsImageName(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	_, ok := imageExts[ext]
	return ok
}

// IsImageData reports whether the content looks like an image.
func IsImageData(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if magicKind(data) == kindImage {
		return true
	}
	sample := data
	if len(sample) > 512 {
		sample = sample[:512]
	}
	ct := http.DetectContentType(sample)
	return strings.HasPrefix(ct, "image/")
}

// IsLikelyText reports whether content is safe to open in a text editor.
func IsLikelyText(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	switch magicKind(data) {
	case kindImage, kindBinary:
		return false
	}
	n := len(data)
	if n > 8192 {
		n = 8192
	}
	sample := data[:n]
	if bytes.IndexByte(sample, 0) >= 0 {
		return false
	}
	ct := http.DetectContentType(sample)
	switch {
	case strings.HasPrefix(ct, "text/"):
		return true
	case ct == "application/json", ct == "application/xml", ct == "application/javascript",
		ct == "application/x-yaml", ct == "application/yaml", ct == "application/x-sh",
		ct == "application/x-httpd-php", ct == "application/x-perl":
		return true
	case strings.HasPrefix(ct, "image/"), strings.HasPrefix(ct, "audio/"), strings.HasPrefix(ct, "video/"):
		return false
	case ct == "application/pdf", ct == "application/zip", ct == "application/gzip",
		ct == "application/x-gzip", ct == "application/x-rar-compressed",
		ct == "application/vnd.ms-cab-compressed", ct == "application/x-msdownload",
		ct == "application/x-executable", ct == "application/x-sharedlib":
		return false
	case ct == "application/octet-stream":
		return utf8.Valid(sample)
	default:
		return utf8.Valid(sample)
	}
}

type fileKind int

const (
	kindUnknown fileKind = iota
	kindImage
	kindBinary
)

func magicKind(data []byte) fileKind {
	switch {
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return kindImage
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return kindImage
	case len(data) >= 6 && (bytes.Equal(data[:6], []byte("GIF87a")) || bytes.Equal(data[:6], []byte("GIF89a"))):
		return kindImage
	case len(data) >= 2 && data[0] == 'B' && data[1] == 'M':
		return kindImage
	case len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return kindImage
	case len(data) >= 4 && bytes.Equal(data[:4], []byte{0x00, 0x00, 0x01, 0x00}):
		return kindImage
	case len(data) >= 4 && bytes.Equal(data[:4], []byte("%PDF")):
		return kindBinary
	case len(data) >= 2 && data[0] == 'P' && data[1] == 'K':
		return kindBinary
	case len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b:
		return kindBinary
	case len(data) >= 2 && data[0] == 'M' && data[1] == 'Z':
		return kindBinary
	case len(data) >= 4 && data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F':
		return kindBinary
	default:
		return kindUnknown
	}
}

// OpenPath opens path with the system default application.
func OpenPath(path string) error {
	path = filepath.Clean(path)
	if _, err := os.Stat(path); err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/C", "start", "", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
