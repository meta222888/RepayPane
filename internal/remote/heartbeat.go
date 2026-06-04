package remote

import (
	"fmt"
	"time"
)

func (c *Client) Ping() error {
	if c.sftp == nil {
		return fmt.Errorf("sftp not connected")
	}
	_, err := c.sftp.Getwd()
	return err
}

// StartHeartbeat sends periodic keepalive checks. Call StopHeartbeat before Close.
func (c *Client) StartHeartbeat(interval time.Duration, onFailure func(error)) {
	if interval <= 0 {
		return
	}
	c.heartbeatMu.Lock()
	c.stopHeartbeatLocked()
	stop := make(chan struct{})
	c.heartbeatStop = stop
	c.heartbeatMu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if err := c.Ping(); err != nil {
					if onFailure != nil {
						onFailure(err)
					}
					return
				}
			}
		}
	}()
}

func (c *Client) StopHeartbeat() {
	c.heartbeatMu.Lock()
	defer c.heartbeatMu.Unlock()
	c.stopHeartbeatLocked()
}

func (c *Client) stopHeartbeatLocked() {
	if c.heartbeatStop != nil {
		close(c.heartbeatStop)
		c.heartbeatStop = nil
	}
}
