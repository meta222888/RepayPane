package ui

import "testing"

func TestParseDfOutput(t *testing.T) {
	out := "Filesystem      Size  Used Avail Use% Mounted on\n" +
		"/dev/sda1        50G   20G   28G  42% /\n" +
		"tmpfs           1.9G     0  1.9G   0% /dev/shm\n"
	rows := parseDfOutput(out)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].mount != "/" || rows[0].pcent != "42%" {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}
	if rows[1].mount != "/dev/shm" {
		t.Fatalf("unexpected second row: %+v", rows[1])
	}
}

func TestSortDiskRowsByUsed(t *testing.T) {
	rows := []diskRow{
		{mount: "/dev/shm", used: "0", size: "1.9G"},
		{mount: "/", used: "20G", size: "50G"},
		{mount: "/data", used: "100G", size: "200G"},
	}
	sortDiskRowsByUsed(rows)
	if rows[0].mount != "/data" || rows[1].mount != "/" || rows[2].mount != "/dev/shm" {
		t.Fatalf("unexpected order: %+v", rows)
	}
}

func TestParseHumanBytes(t *testing.T) {
	if got := parseHumanBytes("20G"); got != 20*1024*1024*1024 {
		t.Fatalf("20G = %d", got)
	}
	if got := parseHumanBytes("512M"); got != 512*1024*1024 {
		t.Fatalf("512M = %d", got)
	}
}
