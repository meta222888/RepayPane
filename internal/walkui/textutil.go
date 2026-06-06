package walkui

import (
	"strings"

	"github.com/lxn/walk"
)

// setMultilineText sets TextEdit content with CRLF line endings (required by Win32 RichEdit).
func setMultilineText(te *walk.TextEdit, text string) {
	if te == nil {
		return
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\n", "\r\n")
	_ = te.SetText(text)
}

func normalizeRemotePath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + strings.TrimPrefix(p, "/")
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}
