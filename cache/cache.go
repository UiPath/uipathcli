// Package cache allows storing temporary cache data which is preserved
// across multiple CLI invocations.
package cache

import "time"

// Cache interface for storing temporary data.
// It is used to persist bearer tokens and other temporary auth tokens
// in order to preserve them across multiple CLI invocations.
type Cache interface {
	Get(key string) (string, time.Time)
	Set(key string, value string, expiresAt time.Time)
}
