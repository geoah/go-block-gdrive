// Implement a LRU cache
package lru

import (
	"container/list"
	"sync"
)

const (
	defaultCapacity = 16
)

// The LRU object
type LRU struct {
	// Read/Write mutex for map and queue
	mu *sync.Mutex
	// Our queue
	q *list.List
	// A quick lookup table
	m map[interface{}]*list.Element
	// The cache capacity
	c int
}

// An entry in the cache
type lruCacheEntry struct {
	key   interface{}
	value interface{}
}

// Put a an entry in the cache with an indexing key.
// If key already exists just update the value and move to front.
// If we exceed the cache extract the last item and return its
// value.
func (l *LRU) Put(key, value interface{}) (rv interface{}, rb bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := &lruCacheEntry{
		key:   key,
		value: value,
	}

	rv, rb = value, false
	// Item already exists
	if e, ok := l.m[key]; ok {
		e.Value = entry
		l.q.MoveToFront(e)
		return
	}

	// We have reached our capacity
	if l.q.Len() < l.c {
		l.m[key] = l.q.PushFront(entry)
		return
	}

	e := l.q.Back()
	delete(l.m, e.Value.(*lruCacheEntry).key)
	rv = e.Value.(*lruCacheEntry).value
	rb = true
	e.Value = entry
	l.q.MoveToFront(e)
	return
}

// Look for the key. Return its value if found and move object up front.
func (l *LRU) Get(key interface{}) (interface{}, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if e, ok := l.m[key]; ok {
		l.q.MoveToFront(e)
		return e.Value.(*lruCacheEntry).value, true
	}

	return nil, false
}
