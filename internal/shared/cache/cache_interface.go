package cache

import (
	"time"
)

// CacheInterface defines the interface for cache operations
type CacheInterface interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dest interface{}) error
	Delete(key string) error
	Exists(key string) (bool, error)
}