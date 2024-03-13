package distributedcache

import (
	"log"
	"sync"
)

type Group struct {
	name   string
	getter Getter
	cache  *Cache
}

// callback function to load data as user wish: type or function as the input
// limited: only 1 function for the interface
type Getter interface {
	Get(key string) ([]byte, error)
}

// defined by user with specific function name
type GetterFunc func(key string) ([]byte, error)

// GetterFuc.Get() to call the function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	lock   sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter, chosenStrategy ...string) *Group {
	lock.Lock()
	defer lock.Unlock()
	strategy := "lru"
	if len(chosenStrategy) > 0 {
		strategy = chosenStrategy[0]
	}

	if getter == nil {
		panic("nil Getter")
	}

	g := &Group{
		name:   name,
		getter: getter,
		cache:  NewCache(strategy, cacheBytes),
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	lock.RLock()
	defer lock.RUnlock()

	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	// TODO: implement the Get method

	if key == "" {
		log.Printf("key is required")
		return ByteView{}, nil
	}
	lock.Lock()
	defer lock.Unlock()
	if v, ok := g.cache.Get(key); ok {
		log.Printf("DistributedCache: hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	// TODO: get cache from other nodes
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.cache.Add(key, value)
}
