package remote

import "testing"

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"/var/www/CRMChat", "'/var/www/CRMChat'"},
		{"", "''"},
		{"/tmp/a'b", "'/tmp/a'\\''b'"},
	}
	for _, tc := range tests {
		if got := shellQuote(tc.in); got != tc.want {
			t.Fatalf("shellQuote(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
