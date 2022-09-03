package executor

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const CachePermissions fs.FileMode = 0600
const CacheDirectory string = "uipath-cli"
const Separator string = "|"

type FileCache struct{}

func (c FileCache) Get(key string) string {
	expiry, value, err := c.readValue(key)
	if err != nil {
		return ""
	}
	if expiry < time.Now().Unix() {
		return ""
	}
	return value
}

func (c FileCache) Set(key string, value string, expiresIn float32) {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return
	}
	expires := time.Now().Unix() + int64(expiresIn) - 30
	data := []byte(fmt.Sprintf("%d%s%s", expires, Separator, value))
	os.WriteFile(path, data, CachePermissions)
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
	split := strings.Split(string(data), Separator)
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
	cacheDirectory := filepath.Join(userCacheDirectory, CacheDirectory)
	os.MkdirAll(cacheDirectory, CachePermissions)

	hash := sha256.Sum256([]byte(key))
	fileName := fmt.Sprintf("%x.cache", hash)
	return filepath.Join(cacheDirectory, fileName), nil
}
