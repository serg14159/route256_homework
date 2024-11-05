package cacher

import (
	"container/list"
	"route256/cart/internal/models"
	"sync"
)

// LRUCache represents LRU cache.
type LRUCache struct {
	capacity int
	mu       sync.Mutex
	cache    map[models.SKU]*list.Element
	list     *list.List
}

// entry represents a key-value pair in the cache.
type entry struct {
	key   models.SKU
	value *models.GetProductResponse
}

// NewLRUCache creates a new LRUCache.
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 100
	}
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[models.SKU]*list.Element, capacity),
		list:     list.New(),
	}
}

// Get retrieves a value from the cache.
func (c *LRUCache) Get(key models.SKU) (*models.GetProductResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return nil, false
}

// Set adds a value to the cache.
func (c *LRUCache) Set(key models.SKU, value *models.GetProductResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		elem.Value.(*entry).value = value
	} else {
		if c.list.Len() >= c.capacity {
			back := c.list.Back()
			if back != nil {
				c.list.Remove(back)
				delete(c.cache, back.Value.(*entry).key)
			}
		}
		elem := c.list.PushFront(&entry{key, value})
		c.cache[key] = elem
	}
}
