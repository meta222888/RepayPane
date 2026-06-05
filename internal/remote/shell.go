package remote

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const shellCommandTimeout = 90 * time.Second

var (
	ErrInteractiveCommand = errors.New("interactive command not supported")
	ErrCommandTimeout     = errors.New("command timed out")
)

var interactiveCommands = map[string]struct{}{
	"vim": {}, "vi": {}, "nvim": {}, "nano": {}, "emacs": {}, "micro": {},
	"less": {}, "more": {}, "top": {}, "htop": {}, "btop": {}, "man": {},
	"watch": {}, "crontab": {}, "passwd": {}, "visudo": {}, "ssh": {},
	"telnet": {}, "ftp": {}, "sftp": {}, "mysql": {}, "psql": {},
	"redis-cli": {}, "mongo": {}, "mongosh": {}, "systemctl": {},
	"journalctl": {}, "su": {}, "login": {},
}

func IsInteractiveCommand(cmd string) bool {
	first := commandName(cmd)
	if first == "" {
		return false
	}
	_, blocked := interactiveCommands[first]
	return blocked
}

func commandName(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	for cmd != "" {
		if strings.HasPrefix(cmd, "sudo ") {
			cmd = strings.TrimSpace(cmd[5:])
			continue
		}
		if strings.HasPrefix(cmd, "env ") {
			parts := strings.Fields(cmd)
			for i := 1; i < len(parts); i++ {
				if !strings.Contains(parts[i], "=") {
					cmd = strings.Join(parts[i:], " ")
					break
				}
				if i == len(parts)-1 {
					return ""
				}
			}
			continue
		}
		fields := strings.Fields(cmd)
		if len(fields) == 0 {
			return ""
		}
		return strings.ToLower(filepath.Base(fields[0]))
	}
	return ""
}

func (c *Client) Run(cmd string) (string, error) {
	out, res := c.runSession(cmd)
	if res.connErr != nil {
		return out, res.connErr
	}
	if res.runErr != nil {
		if out != "" {
			return out, fmt.Errorf("%w: %s", res.runErr, out)
		}
		return "", res.runErr
	}
	return out, nil
}

type shellResult struct {
	connErr  error
	runErr   error
	exitCode int
	timedOut bool
}

func (c *Client) RunCombined(cmd string) (string, error) {
	out, res := c.runSession(cmd)
	if res.connErr != nil {
		return out, res.connErr
	}
	if res.timedOut {
		if out != "" {
			return out, fmt.Errorf("%w\n%s", ErrCommandTimeout, out)
		}
		return "", ErrCommandTimeout
	}
	if res.runErr != nil {
		return out, res.runErr
	}
	return out, nil
}

func (c *Client) runSession(cmd string) (string, shellResult) {
	if c.ssh == nil {
		return "", shellResult{connErr: fmt.Errorf("ssh not connected")}
	}
	session, err := c.ssh.NewSession()
	if err != nil {
		return "", shellResult{connErr: err}
	}

	session.Stdin = bytes.NewReader(nil)
	wrapped := "export TERM=dumb; " + cmd

	type cmdResult struct {
		out []byte
		err error
	}
	done := make(chan cmdResult, 1)
	go func() {
		defer session.Close()
		out, runErr := session.CombinedOutput(wrapped)
		done <- cmdResult{out: out, err: runErr}
	}()

	select {
	case r := <-done:
		return strings.TrimRight(string(r.out), "\n"), classifyRunErr(r.err)
	case <-time.After(shellCommandTimeout):
		_ = session.Signal(ssh.SIGTERM)
		time.Sleep(200 * time.Millisecond)
		_ = session.Signal(ssh.SIGKILL)
		r := <-done
		out := strings.TrimRight(string(r.out), "\n")
		return out, shellResult{runErr: ErrCommandTimeout, timedOut: true}
	}
}

func classifyRunErr(err error) shellResult {
	if err == nil {
		return shellResult{}
	}
	var exitErr *ssh.ExitError
	if errors.As(err, &exitErr) {
		return shellResult{runErr: err, exitCode: exitErr.ExitStatus()}
	}
	return shellResult{connErr: err}
}

// ExitStatus extracts remote exit code from a command error.
func ExitStatus(err error) (int, bool) {
	var exitErr *ssh.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitStatus(), true
	}
	return 0, false
}
