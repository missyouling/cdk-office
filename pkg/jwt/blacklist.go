package jwt

import (
	"time"

	"cdk-office/internal/shared/cache"
)

// TokenBlacklist manages blacklisted JWT tokens
type TokenBlacklist struct {
	cache cache.CacheInterface
}

// NewTokenBlacklist creates a new TokenBlacklist
func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		cache: cache.NewRedisCache(),
	}
}

// AddToBlacklist adds a token to the blacklist with its expiration time
func (tb *TokenBlacklist) AddToBlacklist(token string, exp time.Time) error {
	// Calculate the remaining time until expiration
	now := time.Now()
	if exp.After(now) {
		duration := exp.Sub(now)
		return tb.cache.Set(token, "blacklisted", duration)
	}
	return nil
}

// IsBlacklisted checks if a token is in the blacklist
func (tb *TokenBlacklist) IsBlacklisted(token string) (bool, error) {
	var value string
	err := tb.cache.Get(token, &value)
	if err != nil {
		// If key doesn't exist, token is not blacklisted
		return false, nil
	}
	return value == "blacklisted", nil
}