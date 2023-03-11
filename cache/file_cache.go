package cache

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const cacheFilePermissions = 0600
const cacheDirectoryPermissions = 0700
const cacheDirectory = "uipath"
const separator = "|"

// The FileCache stores data on disk in order to preserve them across
// multiple CLI invocations.
type FileCache struct{}

func (c FileCache) Get(key string) (string, float32) {
	expiry, value, err := c.readValue(key)
	if err != nil {
		return "", 0
	}
	if expiry < time.Now().Unix()+30 {
		return "", 0
	}
	return value, float32(expiry)
}

func (c FileCache) Set(key string, value string, expiresIn float32) {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return
	}
	expires := time.Now().Unix() + int64(expiresIn)
	data := []byte(fmt.Sprintf("%d%s%s", expires, separator, value))
	os.WriteFile(path, data, cacheFilePermissions)
}

func (c FileCache) readValue(key string) (int64, string, error) {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return 0, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, "", err
	}
	split := strings.Split(string(data), separator)
	if len(split) != 2 {
		return 0, "", errors.New("Could not split cache data")
	}
	expiry, err := strconv.ParseInt(split[0], 10, 64)
	if err != nil {
		return 0, "", err
	}
	value := split[1]
	return expiry, value, nil
}

func (c FileCache) cacheFilePath(key string) (string, error) {
	userCacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cacheDirectory := filepath.Join(userCacheDirectory, cacheDirectory)
	os.MkdirAll(cacheDirectory, cacheDirectoryPermissions)

	hash := sha256.Sum256([]byte(key))
	fileName := fmt.Sprintf("%x.cache", hash)
	return filepath.Join(cacheDirectory, fileName), nil
}

func NewFileCache() *FileCache {
	return &FileCache{}
}
