//go:build !windows

package commandline

import (
	"os"
	"path/filepath"
)

// PowershellProfilePath returns powershell profile path on linux.
func PowershellProfilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "powershell", "profile.ps1"), nil
}

// BashrcPath returns .bashrc path on linux.
func BashrcPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".bashrc"), nil
}
