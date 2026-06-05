package ui

import (
	"path"
	"testing"
)

func TestParseDuLinesTyped(t *testing.T) {
	out := "D\t1258291\t/usr/share/doc\nF\t4\t/usr/share/file\n"
	entries := parseDuLines(out, "/usr/share")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if !entries[0].isDir || entries[0].name != "doc" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if entries[0].sizeKB <= entries[1].sizeKB {
		t.Fatalf("expected descending size order, got %+v then %+v", entries[0], entries[1])
	}
	if entries[1].isDir || entries[1].name != "file" {
		t.Fatalf("unexpected second entry: %+v", entries[1])
	}
}

func TestParseDuLinesLegacy(t *testing.T) {
	out := "1.2G\t/usr/share/doc\n512M\t/usr/share/lib\n"
	entries := parseDuLines(out, "/usr/share")
	if len(entries) != 2 || entries[0].name != "doc" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
	if entries[0].sizeKB <= entries[1].sizeKB {
		t.Fatalf("expected doc before lib by size, got %+v", entries)
	}
}

func TestDuSizeRank(t *testing.T) {
	if duSizeRank("1G") <= duSizeRank("500M") {
		t.Fatalf("1G should rank above 500M")
	}
	if duSizeRank("2.0g") <= duSizeRank("1.0G") {
		t.Fatalf("case-insensitive suffix should work")
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
