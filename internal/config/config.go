package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	AppDirName          = ".relaypane"
	ServersFile         = "servers.json"
	MaxEditBytes        = 2 * 1024 * 1024 // 2 MB
	DefaultSFTPPort     = 22
	DefaultHeartbeatSec = 30
)

type Server struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password,omitempty"`
	AutoSSHKey    bool   `json:"auto_ssh_key,omitempty"`
	PrivateKey    string `json:"private_key_path,omitempty"`
	RemoteRoot    string `json:"remote_root,omitempty"`
	LocalRoot     string `json:"local_root,omitempty"`
	HeartbeatSec  int    `json:"heartbeat_sec,omitempty"`
}

// HeartbeatInterval returns the configured interval, or 0 when heartbeat is disabled.
func (s Server) HeartbeatInterval() time.Duration {
	if s.HeartbeatSec <= 0 {
		return 0
	}
	return time.Duration(s.HeartbeatSec) * time.Second
}

type Store struct {
	Servers []Server `json:"servers"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, AppDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func serversPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ServersFile), nil
}

func Load() (*Store, error) {
	path, err := serversPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{}, nil
		}
		return nil, err
	}
	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

func Save(store *Store) error {
	path, err := serversPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
