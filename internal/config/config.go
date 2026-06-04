package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	AppDirName     = ".relaypane"
	ServersFile    = "servers.json"
	MaxEditBytes   = 2 * 1024 * 1024 // 2 MB
	DefaultSFTPPort = 22
)

type Server struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"`
	PrivateKey   string `json:"private_key_path,omitempty"`
	RemoteRoot   string `json:"remote_root,omitempty"`
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
