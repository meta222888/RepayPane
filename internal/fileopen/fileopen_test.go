package fileopen

import (
	"os"
	"testing"
)

func TestIsImageName(t *testing.T) {
	if !IsImageName("photo.PNG") {
		t.Fatal("expected png")
	}
	if IsImageName("readme.txt") {
		t.Fatal("txt is not image")
	}
}

func TestIsImageData(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if !IsImageData(png) {
		t.Fatal("expected png magic")
	}
	if IsLikelyText(png) {
		t.Fatal("png should not be text")
	}
}

func TestIsLikelyText(t *testing.T) {
	if !IsLikelyText([]byte("hello world\n")) {
		t.Fatal("plain text")
	}
	if !IsLikelyText(nil) {
		t.Fatal("empty is text")
	}
	if IsLikelyText([]byte{0x00, 'a'}) {
		t.Fatal("nul byte is binary")
	}
	if IsLikelyText([]byte("%PDF-1.4")) {
		t.Fatal("pdf is binary")
	}
}

func TestOpenPathMissing(t *testing.T) {
	err := OpenPath(os.TempDir() + string(os.PathSeparator) + "relaypane-nonexistent-file-open-test")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
