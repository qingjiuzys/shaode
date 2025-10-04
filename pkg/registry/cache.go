package registry

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache manages local cache for registry data
type Cache struct {
	dir         string
	metadata    map[string]*cacheEntry
	tarballs    map[string]string // package@version -> tarball path
	mu          sync.RWMutex
	maxAge      time.Duration
}

// cacheEntry represents a cached metadata entry
type cacheEntry struct {
	data      *PackageMetadata
	timestamp time.Time
}

// NewCache creates a new cache manager
func NewCache(cacheDir string) *Cache {
	return &Cache{
		dir:      cacheDir,
		metadata: make(map[string]*cacheEntry),
		tarballs: make(map[string]string),
		maxAge:   24 * time.Hour, // Cache metadata for 24 hours
	}
}

// GetPackageMetadata retrieves package metadata from cache
func (c *Cache) GetPackageMetadata(name string) (*PackageMetadata, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.metadata[name]
	if !exists {
		// Try loading from disk
		return c.loadMetadataFromDisk(name)
	}

	// Check if cache is still valid
	if time.Since(entry.timestamp) > c.maxAge {
		delete(c.metadata, name)
		return nil, false
	}

	return entry.data, true
}

// SetPackageMetadata stores package metadata in cache
func (c *Cache) SetPackageMetadata(name string, metadata *PackageMetadata) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metadata[name] = &cacheEntry{
		data:      metadata,
		timestamp: time.Now(),
	}

	// Save to disk
	c.saveMetadataToDisk(name, metadata)
}

// GetTarball retrieves tarball path from cache
func (c *Cache) GetTarball(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path, exists := c.tarballs[key]
	if !exists {
		return "", false
	}

	// Check if file still exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		delete(c.tarballs, key)
		return "", false
	}

	return path, true
}

// SetTarball stores tarball path in cache
func (c *Cache) SetTarball(key, path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tarballs[key] = path
}

// Clear clears all cache entries
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear in-memory cache
	c.metadata = make(map[string]*cacheEntry)
	c.tarballs = make(map[string]string)

	// Clear disk cache
	entries, err := ioutil.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(c.dir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}

// CleanExpired removes expired cache entries
func (c *Cache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for name, entry := range c.metadata {
		if now.Sub(entry.timestamp) > c.maxAge {
			delete(c.metadata, name)
		}
	}
}

// loadMetadataFromDisk loads package metadata from disk cache
func (c *Cache) loadMetadataFromDisk(name string) (*PackageMetadata, bool) {
	metadataPath := filepath.Join(c.dir, "metadata", name+".json")
	
	data, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return nil, false
	}

	var metadata PackageMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, false
	}

	// Check if cache is still valid
	info, err := os.Stat(metadataPath)
	if err != nil {
		return nil, false
	}

	if time.Since(info.ModTime()) > c.maxAge {
		return nil, false
	}

	// Store in memory cache
	c.metadata[name] = &cacheEntry{
		data:      &metadata,
		timestamp: info.ModTime(),
	}

	return &metadata, true
}

// saveMetadataToDisk saves package metadata to disk cache
func (c *Cache) saveMetadataToDisk(name string, metadata *PackageMetadata) error {
	metadataDir := filepath.Join(c.dir, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return err
	}

	metadataPath := filepath.Join(metadataDir, name+".json")
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(metadataPath, data, 0644)
}

// GetCacheDir returns the cache directory path
func (c *Cache) GetCacheDir() string {
	return c.dir
}

// GetCacheStats returns cache statistics
func (c *Cache) GetCacheStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["metadata_count"] = len(c.metadata)
	stats["tarball_count"] = len(c.tarballs)
	stats["cache_dir"] = c.dir
	stats["max_age_hours"] = c.maxAge.Hours()

	// Calculate disk usage
	var diskUsage int64
	filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			diskUsage += info.Size()
		}
		return nil
	})
	stats["disk_usage_mb"] = float64(diskUsage) / (1024 * 1024)

	return stats
}
