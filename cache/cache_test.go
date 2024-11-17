package cache

import (
	"fmt"
	"testing"
)

// verifyIntegrity checks if the cache's internal data structures are valid and consistent.
// It returns an error if any inconsistency is found.
func verifyIntegrity(t *testing.T, cache *LruCache) error {
	// Check empty cache case
	if len(cache.store) == 0 {
		if cache.head != nil {
			return fmt.Errorf("empty cache should have nil head, got non-nil")
		}
		return nil
	}

	// Check size constraints
	if len(cache.store) > cache.capacity {
		return fmt.Errorf("cache size %d exceeds capacity %d", len(cache.store), cache.capacity)
	}

	// Verify circular list integrity
	nodeCount := 0
	visited := make(map[*cacheNode]bool)
	current := cache.head

	// Traverse the list and verify each node
	for {
		if current == nil {
			return fmt.Errorf("unexpected nil node in non-empty cache")
		}

		// Check for cycles (should complete exactly one cycle)
		if visited[current] {
			if nodeCount != len(cache.store) {
				return fmt.Errorf("circular list size (%d) doesn't match store size (%d)", nodeCount, len(cache.store))
			}
			break
		}

		// Basic node validation
		if current.next == nil || current.prev == nil {
			return fmt.Errorf("node has nil pointer: next=%p, prev=%p", current.next, current.prev)
		}

		// Check bidirectional links
		if current.next.prev != current {
			return fmt.Errorf("broken bidirectional link: node.next.prev != node")
		}
		if current.prev.next != current {
			return fmt.Errorf("broken bidirectional link: node.prev.next != node")
		}

		// Verify node exists in store
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

func TestNewCache(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		wantErr  bool
	}{
		{"valid capacity", 5, false},
		{"zero capacity", 0, true},
		{"negative capacity", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(tt.capacity)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCache(%d) error = %v, wantErr %v", tt.capacity, err, tt.wantErr)
				return
			}
			if !tt.wantErr && cache == nil {
				t.Errorf("NewCache(%d) returned nil cache without error", tt.capacity)
			}
		})
	}
}

func TestBasicOperations(t *testing.T) {
	cache, err := NewCache(3)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// Test Set operations
	tests := []struct {
		name   string
		key    string
		value  string
		wantOk bool
	}{
		{"set new item 1", "key1", "value1", false},
		{"set new item 2", "key2", "value2", false},
		{"update existing", "key1", "value1-updated", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cache.Set(tt.key, tt.value)
			if got != tt.wantOk {
				t.Errorf("Set(%s, %s) = %v, want %v", tt.key, tt.value, got, tt.wantOk)
			}
			if err := verifyIntegrity(t, cache); err != nil {
				t.Errorf("integrity check failed after Set: %v", err)
			}
		})
	}

	// Test Get operations
	getTests := []struct {
		name    string
		key     string
		wantVal string
		wantOk  bool
	}{
		{"get existing updated", "key1", "value1-updated", true},
		{"get existing", "key2", "value2", true},
		{"get nonexistent", "nonexistent", "", false},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := cache.Get(tt.key)
			if gotOk != tt.wantOk || gotVal != tt.wantVal {
				t.Errorf("Get(%s) = (%v, %v), want (%v, %v)",
					tt.key, gotVal, gotOk, tt.wantVal, tt.wantOk)
			}
			if err := verifyIntegrity(t, cache); err != nil {
				t.Errorf("integrity check failed after Get: %v", err)
			}
		})
	}
}

func TestEviction(t *testing.T) {
	cache, _ := NewCache(2)

	// Test eviction sequence
	steps := []struct {
		name     string
		op       string
		key      string
		value    string
		checkKey string
		wantOk   bool
	}{
		{"add first", "set", "key1", "value1", "key1", true},
		{"add second", "set", "key2", "value2", "key2", true},
		{"evict first", "set", "key3", "value3", "key1", false}, // key1 should be evicted
		{"verify second", "get", "key2", "", "key2", true},      // key2 should still exist
		{"verify third", "get", "key3", "", "key3", true},       // key3 should exist
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			switch step.op {
			case "set":
				cache.Set(step.key, step.value)
			}

			// Check if the expected key exists
			_, ok := cache.Get(step.checkKey)
			if ok != step.wantOk {
				t.Errorf("after %s: Get(%s) ok = %v, want %v",
					step.name, step.checkKey, ok, step.wantOk)
			}

			if err := verifyIntegrity(t, cache); err != nil {
				t.Errorf("integrity check failed after %s: %v", step.name, err)
			}
		})
	}
}

func TestSingleElement(t *testing.T) {
	cache, _ := NewCache(1)

	t.Run("single element operations", func(t *testing.T) {
		// Add element
		cache.Set("key1", "value1")
		if err := verifyIntegrity(t, cache); err != nil {
			t.Errorf("integrity check failed after Set: %v", err)
		}

		// Get element
		if val, ok := cache.Get("key1"); !ok || val != "value1" {
			t.Error("failed to get single element")
		}
		if err := verifyIntegrity(t, cache); err != nil {
			t.Errorf("integrity check failed after Get: %v", err)
		}

		// Update element
		cache.Set("key1", "newvalue")
		if val, ok := cache.Get("key1"); !ok || val != "newvalue" {
			t.Error("failed to update single element")
		}
		if err := verifyIntegrity(t, cache); err != nil {
			t.Errorf("integrity check failed after update: %v", err)
		}

		// Delete element
		if ok := cache.Delete("key1"); !ok {
			t.Error("failed to delete single element")
		}
		if err := verifyIntegrity(t, cache); err != nil {
			t.Errorf("integrity check failed after Delete: %v", err)
		}

		// Verify empty
		if _, ok := cache.Get("key1"); ok {
			t.Error("element should not exist after deletion")
		}
	})
}

func TestEmptyCache(t *testing.T) {
	cache, _ := NewCache(1)

	t.Run("empty cache operations", func(t *testing.T) {
		// Test Get
		if _, ok := cache.Get("nonexistent"); ok {
			t.Error("Get on empty cache should return false")
		}
		if err := verifyIntegrity(t, cache); err != nil {
			t.Error("integrity check failed after Get on empty cache")
		}

		// Test Delete
		if ok := cache.Delete("nonexistent"); ok {
			t.Error("Delete on empty cache should return false")
		}
		if err := verifyIntegrity(t, cache); err != nil {
			t.Error("integrity check failed after Delete on empty cache")
		}
	})
}
