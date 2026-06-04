package remote

import (
	"bytes"
	"fmt"
)

func (c *Client) Run(cmd string) (string, error) {
	if c.ssh == nil {
		return "", fmt.Errorf("ssh not connected")
	}
	session, err := c.ssh.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var buf bytes.Buffer
	session.Stdout = &buf
	session.Stderr = &buf
	err = session.Run(cmd)
	out := buf.String()
	if err != nil {
		if out != "" {
			return out, fmt.Errorf("%w: %s", err, out)
		}
		return "", err
	}
	return out, nil
}

func (c *Client) RunCombined(cmd string) (string, error) {
	if c.ssh == nil {
		return "", fmt.Errorf("ssh not connected")
	}
	session, err := c.ssh.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	out, err := session.CombinedOutput(cmd)
	return string(out), err
}
