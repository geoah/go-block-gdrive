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

// Create a new LRU of size n. If n <= 0 use defaultCapacity.
func NewLRU(n int) *LRU {
	if n <= 0 {
		n = defaultCapacity
	}

	return &LRU{
		mu: &sync.Mutex{},
		q:  list.New(),
		m:  make(map[interface{}]*list.Element),
		c:  n,
	}
}

// Put a an entry in the cache with an indexing key.
// If key already exists just update the value and move to front.
// If we exceed the cache extract the last item return its
// value and set rb to true.
func (l *LRU) Put(key, value interface{}) (rv interface{}, rb bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := &lruCacheEntry{
		key:   key,
		value: value,
	}

	rv = value
	// Item already exists
	if e, ok := l.m[key]; ok {
		e.Value = entry
		l.q.MoveToFront(e)
		return
	}

	if l.q.Len() < l.c {
		l.m[key] = l.q.PushFront(entry)
		return
	}

	// We have reached our capacity
	e := l.q.Back()
	rv = e.Value.(*lruCacheEntry).value
	rb = true
	delete(l.m, e.Value.(*lruCacheEntry).key)
	e.Value = entry
	l.m[key] = e
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

// Remove a single item from the cache.
// Return true if the item was removed
func (l *LRU) Remove(key interface{}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if e, ok := l.m[key]; ok {
		l.q.Remove(e)
		delete(l.m, key)
		return true
	}

	return false
}

// Flush the LRU. All items not referenced elsewhere will be
// collected by gc. If f is not nil execute it for
// every element of the queue before flushing.
func (l *LRU) Flush(f func(element interface{})) *LRU {
	l.mu.Lock()
	defer l.mu.Unlock()

	if f != nil {
		for e := l.q.Front(); e != nil; e = e.Next() {
			f(e.Value)
		}
	}

	l.q.Init()
	l.m = make(map[interface{}]*list.Element)
	return l
}
