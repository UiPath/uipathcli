package directories

import (
	"os"
	"path/filepath"
)

const cacheDirectoryVarName = "UIPATH_CACHE_PATH"
const offlineModulesDirectoryVarName = "UIPATH_OFFLINE_MODULES_PATH"
const directoryPermissions = 0700

func Temp() (string, error) {
	return userDirectory("tmp")
}

func Cache() (string, error) {
	return userDirectory("cache")
}

func Modules() (string, error) {
	return userDirectory("modules")
}

func OfflineModules() (string, error) {
	directory := os.Getenv(offlineModulesDirectoryVarName)
	if directory == "" {
		executable, err := os.Executable()
		if err != nil {
			return "", err
		}
		directory = filepath.Join(filepath.Dir(executable), "modules")
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
	cacheDirectory := os.Getenv(cacheDirectoryVarName)
	if cacheDirectory == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		cacheDirectory = filepath.Join(userCacheDir, "uipath", "uipathcli")
	}
	_ = os.MkdirAll(cacheDirectory, directoryPermissions)
	return cacheDirectory, nil
}
