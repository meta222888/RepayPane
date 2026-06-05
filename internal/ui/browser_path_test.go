package ui

import "testing"

func TestCleanRemotePath(t *testing.T) {
	tests := map[string]string{
		"":           "/",
		"/":          "/",
		"/var/www/":  "/var/www",
		`\\var\\www`: "/var/www",
	}
	for in, want := range tests {
		if got := cleanRemotePath(in); got != want {
			t.Fatalf("cleanRemotePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRemoteParentPath(t *testing.T) {
	tests := map[string]string{
		"/":              "/",
		"/var/www":       "/var",
		"/var/www/html":  "/var/www",
	}
	for in, want := range tests {
		if got := remoteParentPath(in); got != want {
			t.Fatalf("remoteParentPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestJoinPathRemote(t *testing.T) {
	p := &FilePane{kind: PaneRemote, path: "/var/www"}
	if got := p.joinPath("html"); got != "/var/www/html" {
		t.Fatalf("joinPath(html) = %q, want /var/www/html", got)
	}
}

func TestValidRenameName(t *testing.T) {
	valid := []string{"file.txt", "My Folder", "a"}
	for _, name := range valid {
		if !validRenameName(name) {
			t.Fatalf("validRenameName(%q) = false, want true", name)
		}
	}
	invalid := []string{"", " ", "..", "bad/name", `a\b`, "a:b"}
	for _, name := range invalid {
		if validRenameName(name) {
			t.Fatalf("validRenameName(%q) = true, want false", name)
		}
	}
}
