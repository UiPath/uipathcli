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
	zero := time.Time{}
	if expiry != zero {
		t.Errorf("Should not return expiry value, but got: %v", expiry)
	}
}

func TestGetReturnsDataWhenSet(t *testing.T) {
	cache := NewFileCache()

	key := randomKey()
	expiry := time.Now().UTC().Add(time.Second * 30)

	cache.Set(key, "my-value", expiry)
	value, expiresAt := cache.Get(key)

	if value != "my-value" {
		t.Errorf("Should return data from cache, but got: %v", value)
	}
	if expiry.Unix() != expiresAt.Unix() {
		t.Errorf("Should return expiry value %v, but got: %v", expiry.Unix(), expiresAt.Unix())
	}
}

func TestGetDoesNotReturnExpiredData(t *testing.T) {
	cache := NewFileCache()

	key := randomKey()
	cache.Set(key, "my-value", time.Now().UTC().Add(-time.Second))
	value, expiry := cache.Get(key)

	if value != "" {
		t.Errorf("Should not return any data from cache, but got: %v", value)
	}
	zero := time.Time{}
	if expiry != zero {
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
