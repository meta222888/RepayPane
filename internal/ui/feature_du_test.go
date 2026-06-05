package ui

import (
	"path"
	"testing"
)

func TestParseDuLinesTyped(t *testing.T) {
	out := "D\t1.2G\t/usr/share/doc\nF\t4.0K\t/usr/share/file\n"
	entries := parseDuLines(out, "/usr/share")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if !entries[0].isDir || entries[0].name != "doc" || entries[0].size != "1.2G" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].isDir || entries[1].name != "file" {
		t.Fatalf("unexpected second entry: %+v", entries[1])
	}
}

func TestParseDuLinesLegacy(t *testing.T) {
	out := "1.2G\t/usr/share/doc\n"
	entries := parseDuLines(out, "/usr/share")
	if len(entries) != 1 || entries[0].name != "doc" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}

func TestNormalizeDuPath(t *testing.T) {
	if got := normalizeDuPath("//usr"); got != "/usr" {
		t.Fatalf("//usr = %q", got)
	}
	if got := normalizeDuPath("/"); got != "/" {
		t.Fatalf("/ = %q", got)
	}
	if got := normalizeDuPath(path.Join("/", "usr")); got != "/usr" {
		t.Fatalf("join / + usr = %q", got)
	}
}
