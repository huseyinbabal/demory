package cache

import "github.com/huseyinbabal/lru-cache"

const DefaultCacheCapacity = 1000

type Cache struct {
	store map[string]*lru.Cache
}

func New() *Cache {
	return &Cache{
		store: make(map[string]*lru.Cache),
	}
}

// Put Puts value at a key location under a specified cache. It initializes an empty cache if name does not exist.
// It returns zero if there is already a value at specified key, returns 1 otherwise.
func (c *Cache) Put(name, key string, value []byte) {
	if !c.exists(name) {
		c.store[name] = lru.New(DefaultCacheCapacity)
	}

	c.store[name].Put([]byte(key), value)
}

// Get returns the value associated with key within specific cache.
func (c *Cache) Get(name, key string) []byte {
	if !c.exists(name) {
		return nil
	}

	return c.store[name].Get([]byte(key))
}

// Remove removes value specified by key from a cache. It ignores if key is not in the cache.
func (c *Cache) Remove(name, key string) {
	if c.exists(name) {
		return
	}

	c.store[name].Remove([]byte(key))
}

// Clear removes all the element within cache.
func (c *Cache) Clear(name string) {
	if !c.exists(name) {
		return
	}
	c.store[name].Clear()
}

func (c *Cache) exists(key string) bool {
	if _, ok := c.store[key]; ok {
		return true
	}
	return false
}
