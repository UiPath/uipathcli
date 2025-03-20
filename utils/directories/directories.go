package directories

import (
	"os"
	"path/filepath"
)

const offlineDirectoryVarName = "UIPATH_OFFLINE_PATH"
const directoryPermissions = 0700

func Temp() (string, error) {
	return userDirectory("tmp")
}

func Cache() (string, error) {
	return userDirectory("cache")
}

func Plugins() (string, error) {
	return userDirectory("plugins")
}

func Offline() (string, error) {
	directory := os.Getenv(offlineDirectoryVarName)
	if directory == "" {
		executable, err := os.Executable()
		if err != nil {
			return "", err
		}
		directory = filepath.Join(filepath.Dir(executable), "offline")
	}
	_ = os.MkdirAll(directory, directoryPermissions)
	return directory, nil
}

func userDirectory(name string) (string, error) {
	userDirectory, err := baseUserDirectory()
	if err != nil {
		return "", err
	}
	directory := filepath.Join(userDirectory, name)
	_ = os.MkdirAll(directory, directoryPermissions)
	return directory, nil
}

func baseUserDirectory() (string, error) {
	userCacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	userDirectory := filepath.Join(userCacheDirectory, "uipath", "uipathcli")
	_ = os.MkdirAll(userDirectory, directoryPermissions)
	return userDirectory, nil
}
