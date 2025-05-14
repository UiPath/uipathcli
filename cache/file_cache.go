package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/utils/directories"
)

const cacheFilePermissions = 0600
const separator = "|"

// The FileCache stores data on disk in order to preserve them across
// multiple CLI invocations.
type FileCache struct{}

func (c FileCache) Get(key string) (string, time.Time) {
	expires, value, err := c.readValue(key)
	if err != nil {
		return "", time.Time{}
	}
	expiresAt := time.Unix(expires, 0).UTC()
	if expiresAt.Before(time.Now().UTC()) {
		return "", time.Time{}
	}
	return value, expiresAt
}

func (c FileCache) Set(key string, value string, expiresAt time.Time) {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return
	}
	expires := int64(expiresAt.Unix())
	data := []byte(fmt.Sprintf("%d%s%s", expires, separator, value))
	_ = os.WriteFile(path, data, cacheFilePermissions)
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
	cacheDirectory, err := directories.Cache()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(key))
	fileName := hex.EncodeToString(hash[:])
	return filepath.Join(cacheDirectory, fileName), nil
}

func NewFileCache() *FileCache {
	return &FileCache{}
}
