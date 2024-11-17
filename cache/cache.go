package cache

import (
	"fmt"
	"sync"
)

// Cache Implementation
// A Circular Doubly Linked List,

// CacheNode: In order to o
type CacheNode struct {
	prev  *CacheNode
	next  *CacheNode
	value string
	key   string
}

type LruCache struct {
	mutex    *sync.Mutex
	head     *CacheNode
	capacity int
	store    map[string]*CacheNode
}

// 	INTERNAL FUNCTIONS
// 	______________________
//
// 	WARNING: 		These function are not supposed to be used outside of this package, they suppose that
// 					they are being used in a synchronized execution using mutexes
//
// 	Operations: 	Add To Head, Remove from List, Add to Tail

func (cache *LruCache) addToHead(key, value string) *CacheNode {
	var node CacheNode
	node.value = value
	node.key = key

	// handle empty cache case
	if cache.head == nil {
		node.next = &node
		node.prev = &node
	} else {
		// this code handles the single node case and multinode case correctly
		// the single node case can be verified by tracking memory changes by hand for each instructions
		node.next = cache.head
		node.prev = cache.head.prev

		cache.head.prev.next = &node
		cache.head.prev = &node
	}
	cache.head = &node
	return &node
}

func (cache *LruCache) addToTail(key, value string) *CacheNode {
	node := cache.addToHead(key, value)
	cache.head = cache.head.next
	return node
}

func (cache *LruCache) removeFromList(node *CacheNode) {
	// Handle single node case
	if node.next == node {
		cache.head = nil
		return
	}

	// Handle head case
	if node == cache.head {
		cache.head = node.next
	}
	// this code handles the 2 node case and multinode case correctly
	// the 2 node case can be verified by tracking memory changes by hand for each instructions
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (cache *LruCache) Get(key string) (value string, ok bool) {
	// protect DS
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// Get the node
	node, ok := cache.store[key]

	// get the value
	if !ok {
		return "", ok
	}

	// update the internals
	cache.removeFromList(node)
	newNode := cache.addToHead(key, value)
	cache.store[key] = newNode

	return node.value, ok
}

func (cache *LruCache) Set(key, value string) (updated bool) {
	// protect DS
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// check if this an update
	existing, ok := cache.store[key]
	if ok {
		existing.value = value
		return true
	}

	// check for evicition
	if len(cache.store) == cache.capacity {
		tail := cache.head.prev
		cache.removeFromList(tail)
		delete(cache.store, tail.key)
	}

	// add new node
	node := cache.addToTail(key, value)
	cache.store[key] = node
	return false
}

func (cache *LruCache) Delete(key string) (value string, ok bool) {
	// protect DS
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// check if it exists
	existing, ok := cache.store[key]
	if !ok {
		return "", false
	}

	cache.removeFromList(existing)
	delete(cache.store, existing.key)
	return existing.value, true
}

func (cache *LruCache) Capacity() int {
	return cache.capacity
}

func NewCache(capacity int) (*LruCache, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}

	var mutex sync.Mutex
	store := make(map[string]*CacheNode)
	var cache LruCache = LruCache{
		mutex:    &mutex,
		store:    store,
		head:     nil,
		capacity: capacity,
	}
	return &cache, nil
}
