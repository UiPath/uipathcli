//go:build !windows

package commandline

import (
	"os"
	"path/filepath"
)

func PowershellProfilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "powershell", "profile.ps1"), nil
}

func BashrcPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".bashrc"), nil
}
