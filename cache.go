package distributedcache

import (
	"distributed_cache/evict_strategy"
	"sync"
)

type EvictStrategy interface {
	// Define the methods that your strategies will implement
	Add(key string, value evict_strategy.Value)
	Get(key string) (evict_strategy.Value, bool)
}

type Cache struct {
	lock       sync.RWMutex
	CacheBytes int64
	strategy   string
	Strategy   EvictStrategy // TODO: instantiate the strategy when Add calling
}

func NewCache(chosenStrategy string, cacheBytes int64) *Cache {
	var strategy string

	switch chosenStrategy {
	case "lru":
		strategy = chosenStrategy
	default:
		panic("Invalid strategy: " + chosenStrategy)
	}

	return &Cache{
		CacheBytes: cacheBytes,
		strategy:   strategy,
	}
}

// Confusion: why 2 level of locks in cache?
func (c *Cache) Add(key string, value ByteView) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.Strategy == nil {
		switch c.strategy {
		case "lru":
			c.Strategy = evict_strategy.NewLRU(c.CacheBytes, nil)
		}
	}
	c.Strategy.Add(key, value)
}

func (c *Cache) Get(key string) (ByteView, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.Strategy == nil {
		return ByteView{}, false
	}
	if v, ok := c.Strategy.Get(key); ok {
		return v.(ByteView), ok
	}

	return ByteView{}, false
}
