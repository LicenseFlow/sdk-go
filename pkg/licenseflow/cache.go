// Package licenseflow provides a local entitlement cache for zero-network
// license verification. Thread-safe via sync.RWMutex.
//
// Usage:
//
//	cache := licenseflow.NewEntitlementCache(300, 72*time.Hour)
//	if entry, ok := cache.Get("LF-XXXX"); ok {
//	    // Use cached entitlements — zero network I/O
//	}
package licenseflow

import (
	"sync"
	"time"
)

// CachedEntry stores a single cached entitlement decision.
type CachedEntry struct {
	Data      map[string]interface{} `json:"data"`
	CachedAt  time.Time              `json:"cached_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	Source    string                  `json:"source"` // "network", "cache", "offline"
}

// EntitlementCache provides thread-safe, TTL-based in-memory caching
// with an offline grace period for degraded-mode operation.
type EntitlementCache struct {
	mu          sync.RWMutex
	entries     map[string]*CachedEntry
	ttl         time.Duration
	gracePeriod time.Duration
	strategy    string // "cache-first", "stale-while-revalidate", "network-first"
}

// NewEntitlementCache creates a new cache with the given TTL and offline grace period.
func NewEntitlementCache(ttlSeconds int, gracePeriod time.Duration) *EntitlementCache {
	return &EntitlementCache{
		entries:     make(map[string]*CachedEntry),
		ttl:         time.Duration(ttlSeconds) * time.Second,
		gracePeriod: gracePeriod,
		strategy:    "stale-while-revalidate",
	}
}

// Get retrieves a cached entry. Returns nil and false on miss or full expiry.
func (c *EntitlementCache) Get(key string) (*CachedEntry, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	now := time.Now()

	// Within normal TTL
	if now.Before(entry.ExpiresAt) {
		return &CachedEntry{
			Data:      entry.Data,
			CachedAt:  entry.CachedAt,
			ExpiresAt: entry.ExpiresAt,
			Source:    "cache",
		}, true
	}

	// Within offline grace period
	if now.Before(entry.CachedAt.Add(c.gracePeriod)) {
		return &CachedEntry{
			Data:      entry.Data,
			CachedAt:  entry.CachedAt,
			ExpiresAt: entry.ExpiresAt,
			Source:    "offline",
		}, true
	}

	// Fully expired — remove
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
	return nil, false
}

// Set stores an entitlement decision in the cache.
func (c *EntitlementCache) Set(key string, data map[string]interface{}) {
	now := time.Now()
	c.mu.Lock()
	c.entries[key] = &CachedEntry{
		Data:      data,
		CachedAt:  now,
		ExpiresAt: now.Add(c.ttl),
		Source:    "network",
	}
	c.mu.Unlock()
}

// Invalidate removes a specific cached entry.
func (c *EntitlementCache) Invalidate(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Flush clears all cached entries.
func (c *EntitlementCache) Flush() {
	c.mu.Lock()
	c.entries = make(map[string]*CachedEntry)
	c.mu.Unlock()
}

// GetStrategy determines the cache action for a given key:
// "use_cache", "use_cache_revalidate", or "use_network".
func (c *EntitlementCache) GetStrategy(key string) string {
	entry, ok := c.Get(key)
	if !ok {
		return "use_network"
	}

	switch c.strategy {
	case "cache-first":
		if entry.Source == "offline" {
			return "use_cache_revalidate"
		}
		return "use_cache"
	case "stale-while-revalidate":
		if entry.Source == "cache" {
			return "use_cache"
		}
		return "use_cache_revalidate"
	default:
		return "use_network"
	}
}

// Size returns the number of cached entries.
func (c *EntitlementCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
