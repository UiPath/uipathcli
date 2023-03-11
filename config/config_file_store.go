package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigFileStore reads and writes the configuration file
//
// The config file is
type ConfigFileStore struct {
	data     []byte
	filePath string
}

const configFilePermissions = 0600
const configDirectoryPermissions = 0700

func (s ConfigFileStore) Write(data []byte) error {
	filename, err := s.configurationFilePath()
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(filename), configDirectoryPermissions)
	if err != nil {
		return fmt.Errorf("Error creating configuration folder: %v", err)
	}
	err = os.WriteFile(filename, data, configFilePermissions)
	if err != nil {
		return fmt.Errorf("Error updating configuration file: %v", err)
	}
	return nil
}

func (s ConfigFileStore) Read() ([]byte, error) {
	if s.data != nil {
		return s.data, nil
	}
	filename, err := s.configurationFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filename)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return []byte{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration file '%s': %v", filename, err)
	}
	return data, nil
}

func (s ConfigFileStore) configurationFilePath() (string, error) {
	if s.filePath != "" {
		return s.filePath, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error reading configuration file: %v", err)
	}
	filename := filepath.Join(homeDir, ".uipath", "config")
	return filename, nil
}

func NewConfigFileStore(filePath string) *ConfigFileStore {
	return &ConfigFileStore{
		filePath: filePath,
	}
}

func NewConfigFileStoreWithData(filePath string, data []byte) *ConfigFileStore {
	return &ConfigFileStore{
		filePath: filePath,
		data:     data,
	}
}
