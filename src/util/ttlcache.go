package util

import (
	"sync"
	"time"
)

// ExpireCallback is used as a callback on item expiration or when notifying of an item new to the cache
type expireCallback func(key string, value interface{})

// Cache is a synchronized map of items that can auto-expire once stale
type Cache struct {
	mutex          sync.RWMutex
	opt            CacheOption
	items          map[string]*item
	shutdownSignal chan (chan struct{})
	isShutDown     bool
}

// CacheOption is the optional configuration for Cache
type CacheOption struct {
	TTL            time.Duration
	CleanInterval  time.Duration
	ExpireCallback expireCallback
}

// Get gets an object from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	item, exists := c.items[key]
	c.mutex.RUnlock()
	if !exists {
		return nil, false
	}

	if item.expired() {
		c.mutex.Lock()
		c.opt.ExpireCallback(key, item.data)
		delete(c.items, key)
		c.mutex.Unlock()
		return nil, false
	}

	// item has no expiration
	if item.ttl < 0 {
		return item.data, true
	}

	// update expiry time, puts back to the cache
	item.touch()
	c.mutex.Lock()
	c.items[key] = item
	c.mutex.Unlock()

	return item.data, true
}

// eventLoop name is a disguise. I should convert the lock/unlock to an event loop
func (c *Cache) eventLoop() {
	ticker := time.NewTicker(c.opt.CleanInterval)
	for {
		select {
		case <-ticker.C:
			// RLock is faster than Lock, performant improve to get a slice of keys first
			c.mutex.RLock()
			keys := make([]string, 0, len(c.items))
			for k := range c.items {
				keys = append(keys, k)
			}
			c.mutex.RUnlock()

			// Lock on individual item scan to reduce the lock section
			for _, keyValue := range keys {
				c.mutex.Lock()
				if item, ok := c.items[keyValue]; ok && item.expired() {
					c.opt.ExpireCallback(keyValue, item.data)
					delete(c.items, keyValue)
				}
				c.mutex.Unlock()
			}
		}
	}
}

// Close is not implmented yet
func (c *Cache) Close() {}

// Set adds a new item with a gobally set TTL by the cache
func (c *Cache) Set(key string, data interface{}) {
	c.SetWithTTL(key, data, 0)
}

// SetWithTTL adds a new item with individual ttl
func (c *Cache) SetWithTTL(key string, data interface{}, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.opt.TTL
	}
	item := newItem(key, data, ttl)
	c.mutex.Lock()
	c.items[key] = item
	c.mutex.Unlock()
}

// Delete deletes an item with the key specified
func (c *Cache) Delete(key string) {
	// deletion is an atomic operation therefore it must Write-mutex the entire seciton
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, ok := c.items[key]; ok {
		c.opt.ExpireCallback(key, item.data)
		delete(c.items, key)
	}
}

// Count returns the number of items in the cache
func (c *Cache) Count() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// NewCache is a helper to create instance of the Cache struct
func NewCache(option CacheOption) *Cache {

	shutdownChan := make(chan chan struct{})

	cache := &Cache{
		items:          make(map[string]*item),
		opt:            option,
		shutdownSignal: shutdownChan,
		isShutDown:     false,
	}
	go cache.eventLoop()
	return cache
}
