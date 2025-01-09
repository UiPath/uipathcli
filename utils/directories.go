package utils

import (
	"os"
	"path/filepath"
)

const directoryPermissions = 0700

type Directories struct {
}

func (d Directories) Temp() (string, error) {
	return d.userDirectory("tmp")
}

func (d Directories) Cache() (string, error) {
	return d.userDirectory("cache")
}

func (d Directories) Plugin() (string, error) {
	return d.userDirectory("plugins")
}

func (d Directories) userDirectory(name string) (string, error) {
	userDirectory, err := d.baseUserDirectory()
	if err != nil {
		return "", err
	}
	directory := filepath.Join(userDirectory, name)
	_ = os.MkdirAll(directory, directoryPermissions)
	return directory, nil
}

func (d Directories) baseUserDirectory() (string, error) {
	userCacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	userDirectory := filepath.Join(userCacheDirectory, "uipath", "uipathcli")
	_ = os.MkdirAll(userDirectory, directoryPermissions)
	return userDirectory, nil
}
