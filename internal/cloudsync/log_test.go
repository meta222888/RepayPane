package cloudsync

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIErrorDetail(t *testing.T) {
	err := newResponseError(502, []byte("<html>bad gateway</html>"), errors.New("invalid response: bad json"))
	detail := APIErrorDetail(err)
	if !strings.Contains(detail, "HTTP 502") {
		t.Fatalf("expected status in detail: %q", detail)
	}
	if !strings.Contains(detail, "<html>") {
		t.Fatalf("expected body in detail: %q", detail)
	}
}

func TestAppendLog(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	defer func() {
		_ = os.Setenv("HOME", oldHome)
		_ = os.Setenv("USERPROFILE", oldUserProfile)
	}()

	path, err := LogUploadError(newResponseError(500, []byte("error body"), errors.New("upload failed")))
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(dir, ".relaypane", logFileName)
	if path != expected {
		t.Fatalf("path = %q want %q", path, expected)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "upload") || !strings.Contains(text, "upload failed") {
		t.Fatalf("unexpected log: %q", text)
	}
	if !strings.Contains(text, "error body") {
		t.Fatalf("expected response body in log: %q", text)
	}
}
