package cloudsync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const logFileName = "cloudsync.log"

func logPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".relaypane")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, logFileName), nil
}

// LogPath returns the cloud sync log file path (~/.relaypane/cloudsync.log).
func LogPath() (string, error) {
	return logPath()
}

// AppendLog appends an operation error to the log file and returns the log path.
func AppendLog(operation string, err error, detail string) (string, error) {
	if err == nil {
		return logPath()
	}
	path, errPath := logPath()
	if errPath != nil {
		return "", errPath
	}
	f, errOpen := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if errOpen != nil {
		return path, errOpen
	}
	defer f.Close()

	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] %s\n", ts, operation)
	fmt.Fprintf(f, "Error: %s\n", err.Error())
	if detail != "" {
		fmt.Fprintf(f, "Response:\n%s\n", detail)
	}
	fmt.Fprintln(f, "---")
	return path, nil
}

// LogUploadError logs a failed cloud upload and returns the log file path.
func LogUploadError(err error) (string, error) {
	return AppendLog("upload", err, APIErrorDetail(err))
}
