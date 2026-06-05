package config

import (
	"encoding/json"
	"testing"
)

func TestServerLocalRootJSON(t *testing.T) {
	store := Store{
		Servers: []Server{{
			ID:        "srv-1",
			Name:      "test",
			LocalRoot: `C:\Users\me\Projects`,
		}},
	}
	data, err := json.Marshal(store)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Store
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Servers) != 1 || decoded.Servers[0].LocalRoot != `C:\Users\me\Projects` {
		t.Fatalf("local_root not preserved: %+v", decoded.Servers)
	}
}

func TestHeartbeatInterval(t *testing.T) {
	if got := (Server{HeartbeatSec: 30}).HeartbeatInterval(); got.Seconds() != 30 {
		t.Fatalf("expected 30s, got %v", got)
	}
	if got := (Server{HeartbeatSec: 0}).HeartbeatInterval(); got != 0 {
		t.Fatalf("expected disabled, got %v", got)
	}
}
