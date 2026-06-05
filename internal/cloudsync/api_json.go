package cloudsync

import (
	"encoding/json"
	"strconv"
	"time"
)

// flexString accepts JSON string or number (API may return expired_at as 0).
type flexString string

func (s *flexString) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*s = ""
		return nil
	}
	if b[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		*s = flexString(str)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(b, &num); err != nil {
		return err
	}
	*s = flexString(num.String())
	return nil
}

func (s flexString) String() string {
	return string(s)
}

func formatAPITime(raw string) string {
	if raw == "" {
		return raw
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 1_000_000_000 {
		return time.Unix(n, 0).Local().Format("2006-01-02 15:04:05")
	}
	return raw
}
