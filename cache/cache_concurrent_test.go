package cache

import (
	"fmt"
	"sync"
	"testing"
)

// If all sequential tests pass, then only race conditions need to be checked
// Operations on the cache are protected by a mutex, if a stress test doesn't
// affect the integrity of the cache it should pass the test

func TestCacheConcurrency(t *testing.T) {
	cache, _ := NewCache(5)
	var wg sync.WaitGroup

	// Define test parameters as constants for better maintainability
	const (
		numGoroutines   = 100
		opsPerGoroutine = 100
		// If we increase the keys we increase cache eviction
		numUniqueKeys = 20
	)

	for j := 0; j < numGoroutines; j++ {
		wg.Add(1)
		go func(routineNum int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("key%d", i%numUniqueKeys)
				if (i+routineNum)%2 == 0 {
					cache.Set(key, fmt.Sprintf("value for key%d", i%numUniqueKeys))
				} else {
					cache.Get(key)
				}
			}
		}(j)
	}

	wg.Wait()
	if err := verifyIntegrity(cache); err != nil {
		t.Fatalf("integrity check failed after concurrent access: %v", err)
	}
}
