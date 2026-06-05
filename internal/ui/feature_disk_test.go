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
