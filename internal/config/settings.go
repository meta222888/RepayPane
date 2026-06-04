package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const SettingsFile = "settings.json"

type Settings struct {
	Language string `json:"language,omitempty"` // "en" or "zh"
}

func settingsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SettingsFile), nil
}

func LoadSettings() (*Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Settings{}, nil
		}
		return nil, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func SaveSettings(s *Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
