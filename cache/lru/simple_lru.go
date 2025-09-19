package lru

import (
	"errors"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback[K comparable, V any] func(key K, value V)

// SimpleLRU implements a non-thread safe fixed size LRU cache
type SimpleLRU[K comparable, V any] struct {
	size      int
	evictList *lruList[K, V]
	items     map[K]*entry[K, V]
	onEvict   EvictCallback[K, V]
}

// NewSimpleLRU constructs an LRU of the given size
func NewSimpleLRU[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*SimpleLRU[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}

	c := &SimpleLRU[K, V]{
		size:      size,
		evictList: newLRUList[K, V](),
		items:     make(map[K]*entry[K, V]),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *SimpleLRU[K, V]) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *SimpleLRU[K, V]) Add(key K, value V) (evicted bool) {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value = value
		return false
	}

	// Add new item
	ent := c.evictList.PushFront(key, value)
	c.items[key] = ent

	evict := c.evictList.Length() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return evict
}

// Get looks up a key's value from the cache.
func (c *SimpleLRU[K, V]) Get(key K) (value V, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		return ent.Value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *SimpleLRU[K, V]) Contains(key K) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *SimpleLRU[K, V]) Peek(key K) (value V, ok bool) {
	var ent *entry[K, V]
	if ent, ok = c.items[key]; ok {
		return ent.Value, true
	}
	return
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *SimpleLRU[K, V]) Remove(key K) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (c *SimpleLRU[K, V]) RemoveOldest() (key K, value V, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		c.removeElement(ent)
		return ent.Key, ent.Value, true
	}
	return
}

// GetOldest returns the oldest entry
func (c *SimpleLRU[K, V]) GetOldest() (key K, value V, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		return ent.Key, ent.Value, true
	}
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *SimpleLRU[K, V]) Keys() []K {
	keys := make([]K, c.evictList.Length())
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.PrevEntry() {
		keys[i] = ent.Key
		i++
	}
	return keys
}

// Values returns a slice of the values in the cache, from oldest to newest.
func (c *SimpleLRU[K, V]) Values() []V {
	values := make([]V, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.PrevEntry() {
		values[i] = ent.Value
		i++
	}
	return values
}

// Len returns the number of items in the cache.
func (c *SimpleLRU[K, V]) Len() int {
	return c.evictList.Length()
}

// Resize changes the cache size.
func (c *SimpleLRU[K, V]) Resize(size int) (evicted int) {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff
}

// removeOldest removes the oldest item from the cache.
func (c *SimpleLRU[K, V]) removeOldest() {
	if ent := c.evictList.Back(); ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *SimpleLRU[K, V]) removeElement(e *entry[K, V]) {
	c.evictList.Remove(e)
	delete(c.items, e.Key)
	if c.onEvict != nil {
		c.onEvict(e.Key, e.Value)
	}
}
