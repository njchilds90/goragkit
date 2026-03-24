// Package cache provides lightweight in-memory caches.
package cache

import "sync"

type node[K comparable, V any] struct {
	k          K
	v          V
	prev, next *node[K, V]
}

// LRU is a concurrency-safe fixed-size least-recently-used cache.
type LRU[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	items    map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
}

// NewLRU returns a new cache.
func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	if capacity <= 0 {
		capacity = 1
	}
	return &LRU[K, V]{capacity: capacity, items: make(map[K]*node[K, V], capacity)}
}

// Get returns a value and marks it as recently used.
func (l *LRU[K, V]) Get(k K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n, ok := l.items[k]
	if !ok {
		var zero V
		return zero, false
	}
	l.moveToFront(n)
	return n.v, true
}

// Put stores a value.
func (l *LRU[K, V]) Put(k K, v V) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n, ok := l.items[k]; ok {
		n.v = v
		l.moveToFront(n)
		return
	}
	n := &node[K, V]{k: k, v: v}
	l.items[k] = n
	l.pushFront(n)
	if len(l.items) > l.capacity {
		victim := l.tail
		l.remove(victim)
		delete(l.items, victim.k)
	}
}

func (l *LRU[K, V]) pushFront(n *node[K, V]) {
	n.prev = nil
	n.next = l.head
	if l.head != nil {
		l.head.prev = n
	}
	l.head = n
	if l.tail == nil {
		l.tail = n
	}
}

func (l *LRU[K, V]) moveToFront(n *node[K, V]) {
	if n == l.head {
		return
	}
	l.remove(n)
	l.pushFront(n)
}

func (l *LRU[K, V]) remove(n *node[K, V]) {
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	if l.head == n {
		l.head = n.next
	}
	if l.tail == n {
		l.tail = n.prev
	}
}
