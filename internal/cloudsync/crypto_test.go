package cloudsync

import (
	"testing"

	"github.com/relaypane/relaypane/internal/config"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plain := []byte(`{"version":1,"servers":[]}`)
	enc, err := Encrypt("x", plain)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Decrypt("x", enc)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(plain) {
		t.Fatalf("got %q", out)
	}
}

func TestPackApplyPayloadWithKey(t *testing.T) {
	store := &config.Store{
		Servers: []config.Server{{
			ID:       "srv-1",
			Name:     "test",
			Host:     "example.com",
			Port:     22,
			Username: "root",
		}},
	}
	data, err := PackStore(store)
	if err != nil {
		t.Fatal(err)
	}
	dst := &config.Store{}
	if err := ApplyPayload(data, dst); err != nil {
		t.Fatal(err)
	}
	if len(dst.Servers) != 1 || dst.Servers[0].Host != "example.com" {
		t.Fatalf("unexpected servers: %+v", dst.Servers)
	}
}
