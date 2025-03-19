package cache

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestGetReturnsNoDataWhenNotCached(t *testing.T) {
	cache := NewFileCache()

	value, expiry := cache.Get("UNKNOWN")

	if value != "" {
		t.Errorf("Should not return any data from cache, but got: %v", value)
	}
	if expiry != 0 {
		t.Errorf("Should not return expiry value, but got: %v", expiry)
	}
}

func TestGetReturnsDataWhenSet(t *testing.T) {
	cache := NewFileCache()

	key := randomKey()
	before := time.Now().Unix() + int64(60)

	cache.Set(key, "my-value", 60)
	value, expiry := cache.Get(key)

	if value != "my-value" {
		t.Errorf("Should return data from cache, but got: %v", value)
	}
	if expiry < float32(before) {
		t.Errorf("Should return expiry value which is after %v, but got: %v", before, expiry)
	}
}

func TestGetDoesNotReturnExpiredData(t *testing.T) {
	cache := NewFileCache()

	key := randomKey()
	cache.Set(key, "my-value", -1)
	value, expiry := cache.Get(key)

	if value != "" {
		t.Errorf("Should not return any data from cache, but got: %v", value)
	}
	if expiry != 0 {
		t.Errorf("Should not return expiry value, but got: %v", expiry)
	}
}

func TestGetDoesNotReturnDataWhichExpiresSoon(t *testing.T) {
	cache := NewFileCache()

	key := randomKey()
	cache.Set(key, "my-value", 10)
	value, expiry := cache.Get(key)

	if value != "" {
		t.Errorf("Should not return any data from cache, but got: %v", value)
	}
	if expiry != 0 {
		t.Errorf("Should not return expiry value, but got: %v", expiry)
	}
}

func randomKey() string {
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(fmt.Errorf("Error generating random cache key: %w", err))
	}
	return hex.EncodeToString(randBytes)
}
