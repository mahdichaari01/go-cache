package cache

import (
	"fmt"
	"sync"
)

// The implementation uses two main data structures:
// 1. A circular doubly linked list (circular DLL) for maintaining access order
// 2. A hashmap for O(1) node lookups
//
// The circular DLL design simplifies the code by:
// - Avoiding explicit tail tracking
// - Making edge cases (empty list, single node) behave like normal cases
// - Simplifying head/tail operations
// Note on circularity: It may affect the readability of some code parts, obscure parts are well commented and documented
//
// All operations are O(1) time complexity. Thread safety is ensured through a cache-wide mutex due to operations affecting the overall DS

// A node in the Circular-DLL
type cacheNode struct {
	prev  *cacheNode
	next  *cacheNode
	value string
	key   string
}

type LruCache struct {
	mutex    *sync.Mutex
	head     *cacheNode
	capacity int
	store    map[string]*cacheNode
}

// 	INTERNAL FUNCTIONS
// 	WARNING: 		These function are not supposed to be used outside of this package,
// 					they suppose that they are being used in a synchronized execution using mutexes

// addToHead creates a new node and makes it the head of the DLL
func (cache *LruCache) addToHead(key, value string) *cacheNode {
	var node cacheNode
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

// addToTail adds a new node to the end of the DLL
// It makes use of the circularity of the DLL, it adds the new node to the tail and shifts the head
func (cache *LruCache) addToTail(key, value string) *cacheNode {
	node := cache.addToHead(key, value)
	cache.head = cache.head.next
	return node
}

// removeFromList removes a node from the DLL
func (cache *LruCache) removeFromList(node *cacheNode) {
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

// Public Functions
// 	______________________

// Get retrieves a value from the cache by its key.
// It behaves just like map access eg: value,ok:=m[key]
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
	newNode := cache.addToHead(key, node.value)
	cache.store[key] = newNode

	return node.value, ok
}

// Set adds or updates a key-value pair in the cache.
// An assumption has been made: new elements are added to the tail
// updated elements don't change eviction time
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

// Delete removes the item associated to key, it returns true if element exists, false otherwise
func (cache *LruCache) Delete(key string) (ok bool) {
	// protect DS
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// check if it exists
	existing, ok := cache.store[key]
	if !ok {
		return false
	}

	cache.removeFromList(existing)
	delete(cache.store, existing.key)
	return true
}

// Getter for cache.capacity
func (cache *LruCache) Capacity() int {
	return cache.capacity
}

// Returns current cache size
func (cache *LruCache) Len() int {
	return len(cache.store)
}

// NewCache creates and returns a new LRU cache with the specified capacity.
// Returns an error if capacity is less than or equal to zero.
func NewCache(capacity int) (*LruCache, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}

	var mutex sync.Mutex
	store := make(map[string]*cacheNode)
	var cache LruCache = LruCache{
		mutex:    &mutex,
		store:    store,
		head:     nil,
		capacity: capacity,
	}
	return &cache, nil
}
