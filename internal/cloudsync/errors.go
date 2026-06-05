package cloudsync

import (
	"errors"
	"fmt"
	"strings"
)

const maxLoggedBody = 8192

// ResponseError carries HTTP status and raw API body for logging.
type ResponseError struct {
	StatusCode int
	Body       string
	Err        error
}

func (e *ResponseError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "cloud sync request failed"
}

func (e *ResponseError) Unwrap() error {
	return e.Err
}

func newResponseError(statusCode int, body []byte, err error) error {
	return &ResponseError{
		StatusCode: statusCode,
		Body:       truncateBody(body, maxLoggedBody),
		Err:        err,
	}
}

func truncateBody(body []byte, max int) string {
	if len(body) > max {
		body = body[:max]
	}
	return string(body)
}

// APIErrorDetail returns HTTP status and response body when present.
func APIErrorDetail(err error) string {
	var re *ResponseError
	if !errors.As(err, &re) {
		return ""
	}
	var b strings.Builder
	if re.StatusCode > 0 {
		fmt.Fprintf(&b, "HTTP %d\n", re.StatusCode)
	}
	if re.Body != "" {
		b.WriteString(re.Body)
	}
	return strings.TrimSpace(b.String())
}
