package cache

import (
	"testing"
)

// Unit tests to test the functioning of the cache in a sequential manner
func TestNewCache(t *testing.T) {
	testsTable := []struct {
		name     string
		capacity int
		wantErr  bool
	}{
		{"valid capacity", 5, false},
		{"zero capacity", 0, true},
		{"negative capacity", -1, true},
	}

	for _, tt := range testsTable {
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
			if err := verifyIntegrity(cache); err != nil {
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
			if err := verifyIntegrity(cache); err != nil {
				t.Errorf("integrity check failed after Get: %v", err)
			}
		})
	}
}

// Caches with size 1 will test the limit of the Circular-DL implementation because it will
// trigger node movements in the DLL extensively, this can be seen as a stress test
func TestSingleElement(t *testing.T) {
	cache, _ := NewCache(1)

	t.Run("single element operations", func(t *testing.T) {
		// Add element
		cache.Set("key1", "value1")
		if err := verifyIntegrity(cache); err != nil {
			t.Errorf("integrity check failed after Set: %v", err)
		}

		// Get element
		if val, ok := cache.Get("key1"); !ok || val != "value1" {
			t.Error("failed to get single element")
		}
		if err := verifyIntegrity(cache); err != nil {
			t.Errorf("integrity check failed after Get: %v", err)
		}

		// Update element
		cache.Set("key1", "newvalue")
		if val, ok := cache.Get("key1"); !ok || val != "newvalue" {
			t.Error("failed to update single element")
		}
		if err := verifyIntegrity(cache); err != nil {
			t.Errorf("integrity check failed after update: %v", err)
		}

		// Delete element
		if ok := cache.Delete("key1"); !ok {
			t.Error("failed to delete single element")
		}
		if err := verifyIntegrity(cache); err != nil {
			t.Errorf("integrity check failed after Delete: %v", err)
		}

		// Verify empty
		if _, ok := cache.Get("key1"); ok {
			t.Error("element should not exist after deletion")
		}

		// Add 2 elements
		cache.Set("key1", "val1")
		if err := verifyIntegrity(cache); err != nil {
			t.Errorf("integrity check failed after inserting an element in an empty but previously full cache: %v", err)
		}
		cache.Set("key2", "val2")
		if _, ok := cache.Get("key1"); ok {
			t.Error("key1 should be evicted")
		}

		if val, ok := cache.Get("key2"); !ok || val != "val2" {
			t.Error("key2 should be set correctly")
		}
		if err := verifyIntegrity(cache); err != nil {
			t.Errorf("Integrety check failed after a complex operation: %v", err)
		}

	})
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
		{"evict first", "set", "key3", "value3", "key1", false},                  // key1 should be evicted
		{"verify second", "get", "key2", "", "key2", true},                       // key2 should still exist
		{"verify third", "get", "key3", "", "key3", true},                        // key3 should exist
		{"change priority by update", "set", "key2", "hi", "nonexistent", false}, // key3 has been updated
		{"evict second", "set", "key4", "value", "key3", false},
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

			if err := verifyIntegrity(cache); err != nil {
				t.Errorf("integrity check failed after %s: %v", step.name, err)
			}
		})
	}
}

func TestEmptyCache(t *testing.T) {
	cache, _ := NewCache(1)

	t.Run("empty cache operations", func(t *testing.T) {
		// Test Get
		if _, ok := cache.Get("nonexistent"); ok {
			t.Error("Get on empty cache should return false")
		}
		if err := verifyIntegrity(cache); err != nil {
			t.Error("integrity check failed after Get on empty cache")
		}

		// Test Delete
		if ok := cache.Delete("nonexistent"); ok {
			t.Error("Delete on empty cache should return false")
		}
		if err := verifyIntegrity(cache); err != nil {
			t.Error("integrity check failed after Delete on empty cache")
		}
	})
}
