package cloudsync

import (
	"encoding/json"
	"testing"
)

func TestAPIResponseNumericExpiredAt(t *testing.T) {
	body := `{"code":200,"message":"ok","data":{"key":"relaypane-servers","value":"enc","expired_at":0,"updated_at":1749168391}}`
	var out apiResponse
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatal(err)
	}
	if out.Data == nil {
		t.Fatal("expected data")
	}
	if out.Data.ExpiredAt.String() != "0" {
		t.Fatalf("expired_at = %q", out.Data.ExpiredAt)
	}
	if got := formatAPITime(out.Data.UpdatedAt.String()); got == "1749168391" {
		t.Fatalf("expected formatted updated_at, got raw %q", got)
	}
}

func TestAPIResponseStringTimes(t *testing.T) {
	body := `{"code":200,"data":{"key":"k","value":"v","expired_at":"","updated_at":"2026-06-06 01:26:31"}}`
	var out apiResponse
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatal(err)
	}
	if formatAPITime(out.Data.UpdatedAt.String()) != "2026-06-06 01:26:31" {
		t.Fatalf("updated_at = %q", out.Data.UpdatedAt)
	}
}

func TestFlexStringNull(t *testing.T) {
	var s flexString
	if err := json.Unmarshal([]byte("null"), &s); err != nil {
		t.Fatal(err)
	}
	if s.String() != "" {
		t.Fatalf("got %q", s)
	}
}
