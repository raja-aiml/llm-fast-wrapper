package cache

import (
	"sync"
)

// Cache provides thread-safe in-memory caching for embeddings
type Cache struct {
	cache map[string][]float32
	mutex sync.RWMutex
}

// NewCache creates a new embedding cache
func NewCache() *Cache {
	return &Cache{
		cache: make(map[string][]float32),
	}
}

// Get retrieves an embedding from the cache
func (c *Cache) Get(text string) ([]float32, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	embedding, found := c.cache[text]
	return embedding, found
}

// Set stores an embedding in the cache
func (c *Cache) Set(text string, embedding []float32) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[text] = embedding
}

// Clear empties the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string][]float32)
}

// Size returns the number of entries in the cache
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}
