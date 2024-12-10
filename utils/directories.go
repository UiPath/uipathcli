package utils

import (
	"os"
	"path/filepath"
)

type Directories struct {
}

func (d Directories) Cache() (string, error) {
	userCacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cacheDirectory := filepath.Join(userCacheDirectory, "uipath", "uipathcli")
	return cacheDirectory, nil
}
