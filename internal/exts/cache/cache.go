package cache

import (
	"sync"
	"time"
)

type ItemStruct struct {
	Value      interface{}
	Expiration int64
}

type CacheStruct struct {
	items           map[string]ItemStruct
	mu              sync.RWMutex
	cleanupInterval time.Duration
	maxSize         int
}

func NewCache() *CacheStruct {
	cache := &CacheStruct{
		items:           make(map[string]ItemStruct),
		cleanupInterval: time.Second * 300,
	}
	go cache.startCleanupRoutine()
	return cache
}

func (c *CacheStruct) startCleanupRoutine() {
	for {
		<-time.After(c.cleanupInterval)
		c.cleanupExpiredItems()
	}
}

func (c *CacheStruct) cleanupExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().Unix()
	for key, item := range c.items {
		if item.Expiration > 0 && now > item.Expiration {
			delete(c.items, key)
		}
	}
}

func (c *CacheStruct) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	//if c.maxSize > 0 && len(c.items) >= c.maxSize {
	//	//实现驱逐item需要改写set、get等功能，关键词LRU、链表
	//	//c.removeOldest()
	//}

	expiration := time.Now().Add(duration).Unix()
	if duration == time.Duration(0) {
		expiration = 0
	}

	c.items[key] = ItemStruct{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *CacheStruct) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil
	}

	if item.Expiration > 0 && time.Now().Unix() > item.Expiration {
		delete(c.items, key)
		return nil
	}

	return item.Value
}

func (c *CacheStruct) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *CacheStruct) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]ItemStruct)
}
