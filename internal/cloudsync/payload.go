package cloudsync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relaypane/relaypane/internal/config"
)

const PayloadVersion = 1

const StorageKey = "relaypane_servers"

type cloudServer struct {
	config.Server
	PrivateKeyPEM string `json:"private_key_pem,omitempty"`
}

type Payload struct {
	Version int           `json:"version"`
	Servers []cloudServer `json:"servers"`
}

func PackStore(store *config.Store) ([]byte, error) {
	if store == nil {
		return nil, fmt.Errorf("no server data")
	}
	out := Payload{Version: PayloadVersion, Servers: make([]cloudServer, 0, len(store.Servers))}
	for _, s := range store.Servers {
		cs := cloudServer{Server: s}
		if s.PrivateKey != "" {
			data, err := os.ReadFile(s.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("read private key %s: %w", s.PrivateKey, err)
			}
			cs.PrivateKeyPEM = string(data)
		}
		out.Servers = append(out.Servers, cs)
	}
	return json.Marshal(out)
}

func ApplyPayload(data []byte, store *config.Store) error {
	var payload Payload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	if payload.Version != PayloadVersion {
		return fmt.Errorf("unsupported payload version %d", payload.Version)
	}
	keysDir, err := configKeysDir()
	if err != nil {
		return err
	}
	servers := make([]config.Server, 0, len(payload.Servers))
	for _, cs := range payload.Servers {
		s := cs.Server
		if cs.PrivateKeyPEM != "" {
			id := s.ID
			if id == "" {
				id = fmt.Sprintf("srv-import-%d", len(servers))
				s.ID = id
			}
			keyPath := filepath.Join(keysDir, id+".pem")
			if err := os.WriteFile(keyPath, []byte(cs.PrivateKeyPEM), 0o600); err != nil {
				return fmt.Errorf("write private key: %w", err)
			}
			s.PrivateKey = keyPath
			s.AutoSSHKey = false
		}
		servers = append(servers, s)
	}
	store.Servers = servers
	return nil
}

func configKeysDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, config.AppDirName, "keys")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}
