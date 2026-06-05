package ui

import (
	"strings"
	"testing"
)

func TestParseNetIfaces(t *testing.T) {
	out := "eth0\t1234567890\t987654321\n" +
		"docker0\t0\t0\n" +
		"ens33\t500\t1200\n"
	stats := parseNetIfaces(out)
	if len(stats) != 3 {
		t.Fatalf("expected 3 stats, got %d", len(stats))
	}
	if stats[0].name != "eth0" || stats[0].rx != 1234567890 || stats[0].tx != 987654321 {
		t.Fatalf("unexpected eth0: %+v", stats[0])
	}
}

func TestFilterNetIfaces(t *testing.T) {
	stats := []netIfaceStat{
		{name: "lo", rx: 100, tx: 100},
		{name: "eth0", rx: 1000, tx: 500},
		{name: "veth123", rx: 0, tx: 0},
		{name: "ens33", rx: 10, tx: 20},
	}
	filtered := filterNetIfaces(stats)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered stats, got %d: %+v", len(filtered), filtered)
	}
	if filtered[0].name != "eth0" || filtered[1].name != "ens33" {
		t.Fatalf("unexpected order: %+v", filtered)
	}
}

func TestParseNetRoutes(t *testing.T) {
	out := "default via 192.168.1.1 dev eth0\n\n192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.10\n"
	routes := parseNetRoutes(out)
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}
}

func TestFormatBytesPerSec(t *testing.T) {
	if got := formatBytesPerSec(512); got != "512 B/s" {
		t.Fatalf("512 = %q", got)
	}
	if got := formatBytesPerSec(1536); got != "1.5 KB/s" {
		t.Fatalf("1536 = %q", got)
	}
}

func TestCompactPortOutput(t *testing.T) {
	in := "=== Listening Ports ===\n\n\nNetid State Recv-Q Send-Q Local Address:Port\n\ntcp   LISTEN 0      128    0.0.0.0:22\n"
	out := compactPortOutput(in)
	if strings.Contains(out, "===") {
		t.Fatalf("header line should be removed: %q", out)
	}
	if strings.Contains(out, "\n\n") {
		t.Fatalf("blank lines should be removed: %q", out)
	}
	if !strings.Contains(out, "tcp   LISTEN") {
		t.Fatalf("missing port line: %q", out)
	}
}
