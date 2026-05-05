package ratelimiter

// Storage is the interface for rate limiter state storage.
// Implementations must be safe for concurrent use.
// This allows for pluggable storage backends (memory, Redis, etc.)
// for distributed rate limiting.
type Storage interface {
	// Get retrieves the value associated with the given key.
	// It returns the value and a boolean indicating whether the key exists.
	Get(key string) (interface{}, bool)

	// Set sets the value associated with the given key.
	Set(key string, value interface{})

	// Delete deletes the value associated with the given key.
	Delete(key string)

	// Clear removes all keys and values.
	Clear()
}
