package screen

import (
	"sync"
	"time"
)

// LookupCacheTTL is how long a weather, sports or news result is reused. The tablet
// re-fetches the screen every few minutes; without this, one weather widget alone
// would make about 860 OpenWeather calls a day (the free tier allows 1,000).
const LookupCacheTTL = 10 * time.Minute

type cacheEntry struct {
	value   any
	expires time.Time
}

// lookupCache keeps successful lookups for LookupCacheTTL; failures aren't kept,
// so the next request tries again
type lookupCache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	now     func() time.Time
}

func newLookupCache(now func() time.Time) *lookupCache {
	return &lookupCache{entries: map[string]cacheEntry{}, now: now}
}

// get returns the cached value for key, or calls fetch and caches its result
func (c *lookupCache) get(key string, fetch func() (any, error)) (any, error) {
	c.mu.Lock()
	if e, ok := c.entries[key]; ok && c.now().Before(e.expires) {
		c.mu.Unlock()
		return e.value, nil
	}
	c.mu.Unlock()

	value, err := fetch()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.entries[key] = cacheEntry{value: value, expires: c.now().Add(LookupCacheTTL)}
	c.mu.Unlock()
	return value, nil
}
