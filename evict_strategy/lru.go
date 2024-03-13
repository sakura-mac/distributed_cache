package evict_strategy

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	maxBytes int64
	nbytes   int64
	lock     sync.Mutex
	ll       *list.List
	cache    map[string]*list.Element
	// callback function
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func NewLRU(maxBytes int64, onEvicted func(string, Value)) *LRUCache {
	return &LRUCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *LRUCache) Add(key string, value Value) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Remove the tail element
func (c *LRUCache) RemoveOldest() {
	// c.lock.Lock()
	// defer c.lock.Unlock()
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *LRUCache) Get(key string) (value Value, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// Len the number of cache entries
func (c *LRUCache) Len() int {
	return c.ll.Len()
}
