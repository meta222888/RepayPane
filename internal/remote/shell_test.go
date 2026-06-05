package remote

import "testing"

func TestIsInteractiveCommand(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"vim /var/test.txt", true},
		{"sudo vim /etc/hosts", true},
		{"ls -la", false},
		{"cat /var/test.txt", false},
		{"head -n 20 /var/log/syslog", false},
		{"top", true},
		{"env TERM=xterm vim", true},
	}
	for _, tc := range cases {
		if got := IsInteractiveCommand(tc.cmd); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.cmd, got, tc.want)
		}
	}
}

func TestCommandName(t *testing.T) {
	if got := commandName("sudo vim foo"); got != "vim" {
		t.Fatalf("got %q", got)
	}
}
