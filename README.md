# GoCache - Thread-Safe LRU Cache Implementation

GoCache is a high-performance, thread-safe in-memory cache implementation in Go that uses the Least Recently Used (LRU) eviction strategy. It provides concurrent access to cached data while maintaining data integrity in multi-threaded environments.

## Features

- Thread-safe operations using Go's synchronization primitives
- LRU (Least Recently Used) eviction strategy
- Configurable cache capacity
- Simple key-value string interface
- Support for concurrent reads and writes
- Interactive demo included

## Implementation Details
### Core Components

- **Cache Structure**: Thread-safe implementation using mutex for synchronization
- **LRU Implementation**: Circular Doubly-linked list and hash map for O(1) operations
- **Concurrency Control**: Using sync.Mutex for thread safety

### Operations

- Set(key string, value string) bool
- Get(key string) (string, bool) 
- Delete(key string) bool 

### Thread Safety
The implementation ensures thread safety through:
- Mutex protection for all cache operations
- Atomic updates for LRU management
- Safe concurrent access patterns

## Design Decisions

1. **String Keys and Values**: Chosen for simplicity and common use cases
2. **Mutex vs RWMutex**: Single mutex chosen for simplicity and to prevent potential write starvation
3. **LRU Implementation**: Custom doubly-linked list for efficient node removal and updates

## Installation

Clone the repository:
```bash
git clone https://github.com/yourusername/gocache
cd gocache
```
Interactive Demo:
```bash 
go run main.go
```

## Code Example
```go 
// Create a new cache with capacity of 5
cache, err := NewCache(5)
if err != nil {
    log.Fatal(err)
}

// Set a value
cache.Set("user1", "John Doe")

// Get a value
value, exists := cache.Get("user1")
if exists {
    fmt.Printf("Value: %s\n", value)
}

// Delete a value
deleted := cache.Delete("user1")
```

## Testing
The implementation includes comprehensive test coverage for both functionality and concurrency scenarios. To run the tests:
```bash
go test ./...
```
To run tests with race condition detection:
```bash
go test -race ./...
```

