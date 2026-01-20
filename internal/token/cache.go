// Package token provides access token caching for data plane operations.
//
// Access tokens are required for data plane operations (code execution, file operations, etc.)
// but acquiring them has different behaviors depending on the control plane backend:
//   - Cloud backend: Can call AcquireSandboxInstanceToken API to get a token at any time
//   - E2B backend: Token is only returned during instance creation
//
// This cache provides a persistent file-based storage to save instance ID to access token
// mappings, allowing CLI commands to retrieve tokens across invocations.
package token

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// CacheDir is the directory name under user home for storing cache files
	CacheDir = ".ags"
	// CacheFile is the filename for token cache
	CacheFile = "tokens.json"
	// CacheVersion is the current version of cache file format
	CacheVersion = 1
)

// TokenEntry represents a cached access token
type TokenEntry struct {
	AccessToken string    `json:"access_token"`
	CreatedAt   time.Time `json:"created_at"`
}

// CacheData represents the structure of the cache file
type CacheData struct {
	Version int                    `json:"version"`
	Tokens  map[string]*TokenEntry `json:"tokens"`
}

// Cache manages instance access tokens with file-based persistence.
// It is safe for concurrent use.
type Cache struct {
	path string
	mu   sync.RWMutex
}

// NewCache creates a new token cache.
// The cache file is stored at ~/.ags/tokens.json
func NewCache() (*Cache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, CacheDir)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		path: filepath.Join(cacheDir, CacheFile),
	}, nil
}

// load reads the cache file and returns the cache data.
// If the file doesn't exist, returns an empty cache data.
func (c *Cache) load() (*CacheData, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CacheData{
				Version: CacheVersion,
				Tokens:  make(map[string]*TokenEntry),
			}, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache CacheData
	if err := json.Unmarshal(data, &cache); err != nil {
		// If the file is corrupted, start fresh
		return &CacheData{
			Version: CacheVersion,
			Tokens:  make(map[string]*TokenEntry),
		}, nil
	}

	// Handle version migration if needed
	if cache.Version < CacheVersion {
		cache.Version = CacheVersion
	}

	if cache.Tokens == nil {
		cache.Tokens = make(map[string]*TokenEntry)
	}

	return &cache, nil
}

// save writes the cache data to file
func (c *Cache) save(cache *CacheData) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Get retrieves the access token for an instance.
// Returns the token and true if found, empty string and false otherwise.
func (c *Cache) Get(instanceID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cache, err := c.load()
	if err != nil {
		return "", false
	}

	entry, ok := cache.Tokens[instanceID]
	if !ok || entry == nil {
		return "", false
	}

	return entry.AccessToken, true
}

// Set stores the access token for an instance.
func (c *Cache) Set(instanceID, accessToken string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cache, err := c.load()
	if err != nil {
		return err
	}

	cache.Tokens[instanceID] = &TokenEntry{
		AccessToken: accessToken,
		CreatedAt:   time.Now(),
	}

	return c.save(cache)
}

// Delete removes the access token for an instance.
func (c *Cache) Delete(instanceID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cache, err := c.load()
	if err != nil {
		return err
	}

	delete(cache.Tokens, instanceID)
	return c.save(cache)
}

// Clear removes all cached tokens.
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cache := &CacheData{
		Version: CacheVersion,
		Tokens:  make(map[string]*TokenEntry),
	}

	return c.save(cache)
}

// List returns all cached instance IDs.
func (c *Cache) List() ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cache, err := c.load()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(cache.Tokens))
	for id := range cache.Tokens {
		ids = append(ids, id)
	}

	return ids, nil
}
