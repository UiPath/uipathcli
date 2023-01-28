package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigStore struct {
	Config     []byte
	ConfigFile string
}

const configFilePermissions = 0600
const configDirectoryPermissions = 0700

func (s ConfigStore) Write(data []byte) error {
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

func (s ConfigStore) Read() ([]byte, error) {
	if s.Config != nil {
		return s.Config, nil
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

func (s ConfigStore) configurationFilePath() (string, error) {
	if s.ConfigFile != "" {
		return s.ConfigFile, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error reading configuration file: %v", err)
	}
	filename := filepath.Join(homeDir, ".uipathcli", "config")
	return filename, nil
}
