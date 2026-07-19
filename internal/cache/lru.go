// Package cache provides an in-memory, TTL-aware LRU cache for rendered
// SVG cards.
package cache

import (
	"container/list"
	"sync"
	"time"
)

type entry struct {
	key      string
	value    []byte
	storedAt time.Time
}

// LRU is a fixed-capacity, thread-safe cache with TTL-on-read expiry.
// The least recently used entry is evicted when a new entry would exceed
// capacity.
type LRU struct {
	mu       sync.Mutex
	capacity int
	ttl      time.Duration
	now      func() time.Time
	ll       *list.List
	items    map[string]*list.Element
}

// New returns an LRU cache holding at most capacity entries, each valid
// for ttl after being set.
func New(capacity int, ttl time.Duration) *LRU {
	return &LRU{
		capacity: capacity,
		ttl:      ttl,
		now:      time.Now,
		ll:       list.New(),
		items:    make(map[string]*list.Element, capacity),
	}
}

// Get returns the cached value for key, if present and not expired.
func (c *LRU) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		return nil, false
	}

	e := el.Value.(*entry) //nolint:forcetypeassert // container/list.Element.Value is always *entry, set exclusively by this type
	if c.now().Sub(e.storedAt) >= c.ttl {
		c.removeElement(el)
		return nil, false
	}

	c.ll.MoveToFront(el)
	return e.value, true
}

// Set stores value under key, evicting the least recently used entry if
// the cache is at capacity.
func (c *LRU) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		e := el.Value.(*entry) //nolint:forcetypeassert // see Get
		e.value = value
		e.storedAt = c.now()
		c.ll.MoveToFront(el)
		return
	}

	el := c.ll.PushFront(&entry{key: key, value: value, storedAt: c.now()})
	c.items[key] = el

	if c.ll.Len() > c.capacity {
		c.removeOldest()
	}
}

// Len returns the number of entries currently stored, including any not
// yet expired but that Get has not observed.
func (c *LRU) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ll.Len()
}

func (c *LRU) removeOldest() {
	el := c.ll.Back()
	if el != nil {
		c.removeElement(el)
	}
}

func (c *LRU) removeElement(el *list.Element) {
	c.ll.Remove(el)
	e := el.Value.(*entry) //nolint:forcetypeassert // see Get
	delete(c.items, e.key)
}
