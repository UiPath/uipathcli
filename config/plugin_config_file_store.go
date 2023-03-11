package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// PluginConfigFileStore reads the plugin configuration file from disk
//
// The store searches in the $HOME/.uipath for a file called 'plugin'
// to read the available authenticator plugins.
type PluginConfigFileStore struct {
	filePath string
}

func (s PluginConfigFileStore) Read() ([]byte, error) {
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

func (s PluginConfigFileStore) pluginsFilePath() (string, error) {
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

func NewPluginConfigFileStore(filePath string) *PluginConfigFileStore {
	return &PluginConfigFileStore{filePath}
}
