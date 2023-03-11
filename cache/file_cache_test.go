package cache

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestGetReturnsNoDataWhenNotCached(t *testing.T) {
	cache := FileCache{}

	value, expiry := cache.Get("UNKNOWN")

	if value != "" {
		t.Errorf("Should not return any data from cache, but got: %v", value)
	}
	if expiry != 0 {
		t.Errorf("Should not return expiry value, but got: %v", expiry)
	}
}

func TestGetReturnsDataWhenSet(t *testing.T) {
	cache := FileCache{}

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
	cache := FileCache{}

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
	cache := FileCache{}

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
	rand.Read(randBytes)
	return hex.EncodeToString(randBytes)
}
