package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 24 * time.Hour

// CacheInfo stores cache metadata
type CacheInfo struct {
	LastFetch time.Time `json:"last_fetch"`
}

// Cache manages the template cache
type Cache struct {
	cacheDir string
}

// NewCache creates a new cache manager
func NewCache(cacheDir string) *Cache {
	return &Cache{cacheDir: cacheDir}
}

// CacheInfoPath returns the path to cache.json
func (c *Cache) CacheInfoPath() string {
	return filepath.Join(c.cacheDir, "cache.json")
}

// LoadInfo loads the cache info from disk
func (c *Cache) LoadInfo() (*CacheInfo, error) {
	data, err := os.ReadFile(c.CacheInfoPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &CacheInfo{}, nil
		}
		return nil, fmt.Errorf("reading cache info: %w", err)
	}

	var info CacheInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parsing cache info: %w", err)
	}

	return &info, nil
}

// SaveInfo saves the cache info to disk
func (c *Cache) SaveInfo(info *CacheInfo) error {
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache info: %w", err)
	}

	return os.WriteFile(c.CacheInfoPath(), data, 0644)
}

// IsStale returns true if the cache is older than the TTL
func (c *Cache) IsStale() bool {
	info, err := c.LoadInfo()
	if err != nil {
		return true
	}

	if info.LastFetch.IsZero() {
		return true
	}

	return time.Since(info.LastFetch) > cacheTTL
}

// MarkFetched updates the last fetch time to now
func (c *Cache) MarkFetched() error {
	return c.SaveInfo(&CacheInfo{
		LastFetch: time.Now(),
	})
}

// NeedsInitialFetch returns true if there's no cached templates at all
func (c *Cache) NeedsInitialFetch() bool {
	templatesDir := filepath.Join(c.cacheDir, "templates")
	_, err := os.Stat(templatesDir)
	return os.IsNotExist(err)
}
