package update

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.1.0", "1.0.9", 1},
		{"2.0.0", "1.9.9", 1},
		{"v1.0.0", "1.0.1", -1},
	}
	for _, tc := range tests {
		got := compareVersions(normalizeVersion(tc.a), normalizeVersion(tc.b))
		if got != tc.want {
			t.Fatalf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	if !IsNewer("1.0.0", "1.0.1") {
		t.Fatal("expected newer")
	}
	if IsNewer("1.0.1", "1.0.0") {
		t.Fatal("expected not newer")
	}
}
