package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// PluginConfigStore reads the plugin configuration file
type PluginConfigStore struct {
	filePath string
}

func (s PluginConfigStore) Read() ([]byte, error) {
	filename, err := s.pluginsFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("Error reading plugins file '%s': %v", filename, err)
	}
	return data, nil
}

func (s PluginConfigStore) pluginsFilePath() (string, error) {
	if s.filePath != "" {
		return s.filePath, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error reading plugins file: %v", err)
	}
	filename := filepath.Join(homeDir, ".uipath", "plugins")
	return filename, nil
}

func NewPluginConfigStore(filePath string) *PluginConfigStore {
	return &PluginConfigStore{filePath}
}
