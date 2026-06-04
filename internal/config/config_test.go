package config

import "testing"

func TestHeartbeatInterval(t *testing.T) {
	if got := (Server{HeartbeatSec: 30}).HeartbeatInterval(); got.Seconds() != 30 {
		t.Fatalf("expected 30s, got %v", got)
	}
	if got := (Server{HeartbeatSec: 0}).HeartbeatInterval(); got != 0 {
		t.Fatalf("expected disabled, got %v", got)
	}
}
