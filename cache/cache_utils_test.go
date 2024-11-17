package cache

import (
	"fmt"
)

// The circular-doubly linked list has many moving parts, pointers and links
// this methods attempts to check all aspects of the implementation internals
// for incositencies

// verifyIntegrity is a test only function used to check for the following:
//   - cohesion between the LruCache instance and the underlying data structure
//   - cohesion between linked list and hashmap
//   - edge cases: empty cache, single node,
func verifyIntegrity(cache *LruCache) error {
	// check empty cache
	if len(cache.store) == 0 {
		if cache.head != nil {
			return fmt.Errorf("empty cache should have nil head, got non-nil")
		}
		return nil
	}

	// check size constraints
	if len(cache.store) > cache.capacity {
		return fmt.Errorf("cache size %d exceeds capacity %d", len(cache.store), cache.capacity)
	}

	// verify circular list integrity
	nodeCount := 0
	visited := make(map[*cacheNode]bool)
	current := cache.head

	// circualr dll traversal
	for {
		// shoduldn't have nil nodes
		if current == nil {
			return fmt.Errorf("Unexpected nil, cache is not empty")
		}

		// Check for cycles (should complete exactly one cycle)
		if visited[current] {
			if nodeCount != len(cache.store) {
				return fmt.Errorf("circular list size (%d) doesn't match store size (%d)", nodeCount, len(cache.store))
			}
			break
		}

		// shouldn't have nil pointers in non-empty cache
		if current.next == nil || current.prev == nil {
			return fmt.Errorf("node has nil pointer: next=%p, prev=%p", current.next, current.prev)
		}

		// checking order and pointer correctness
		if current.next.prev != current {
			return fmt.Errorf("broken bidirectional link: node.next.prev != node")
		}
		if current.prev.next != current {
			return fmt.Errorf("broken bidirectional link: node.prev.next != node")
		}

		// verify node exists in store
		storeNode, exists := cache.store[current.key]
		if !exists {
			return fmt.Errorf("node with key %s exists in list but not in store", current.key)
		}
		if storeNode != current {
			return fmt.Errorf("store points to different node for key %s", current.key)
		}

		visited[current] = true
		nodeCount++
		current = current.next

		// Safety check for infinite loops
		if nodeCount > len(cache.store) {
			return fmt.Errorf("circular list appears to have more nodes than store")
		}
	}

	// Verify all store entries are in the list
	for key, node := range cache.store {
		if !visited[node] {
			return fmt.Errorf("node for key %s exists in store but not in list", key)
		}
	}

	return nil
}
