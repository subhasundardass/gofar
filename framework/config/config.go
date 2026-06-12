// Package config provides a simple, flexible configuration system for GoFar.
//
// Config values are resolved in this priority order (highest → lowest):
//  1. In-memory overrides set via Set().
//  2. OS environment variables.
//  3. Values loaded from a .env file (via LoadEnvFile).
//  4. Fallback defaults supplied at the call site.
//
// Example – bootstrapping:
//
//	cfg := config.New()
//	cfg.LoadEnvFile(".env")       // optional – silently ignored if file absent
//	cfg.Load("APP_ENV", "local")  // pre-load with a default
//
//	env := cfg.Get("APP_ENV")                   // "local" unless overridden
//	port := cfg.GetIntWithDefault("PORT", 3000) // typed getter with fallback
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config is a thread-safe key/value configuration store.
// Always create via New().
type Config struct {
	mu     sync.RWMutex
	values map[string]string
}

// New returns an empty Config ready for use.
func New() *Config {
	return &Config{
		values: make(map[string]string),
	}
}

// LoadEnvFile reads a .env-style file and populates the config.
// Lines beginning with '#' are treated as comments and skipped.
// If the file does not exist the call is silently ignored so that
// development .env files are truly optional.
//
// Example .env file:
//
//	APP_ENV=production
//	DB_DSN=postgres://user:pass@host/db
func (c *Config) LoadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // file is optional
		}
		return fmt.Errorf("config: open %s: %w", path, err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			continue // malformed line – skip silently
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Strip optional surrounding quotes
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		c.values[key] = val
	}
	return nil
}

// Load reads key from the OS environment and stores it. If the environment
// variable is not set, fallback is stored instead.
// This is the primary way to declare expected config keys with their defaults.
func (c *Config) Load(key, fallback string) {
	if v := os.Getenv(key); v != "" {
		c.mu.Lock()
		c.values[key] = v
		c.mu.Unlock()
		return
	}
	c.mu.Lock()
	c.values[key] = fallback
	c.mu.Unlock()
}

// Set stores a value directly. It overrides any previously loaded value for
// the same key and is useful in tests or runtime config changes.
func (c *Config) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// Get returns the stored value for key, or an empty string if not found.
func (c *Config) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

// GetWithDefault returns the stored value for key. If the key is absent or
// empty, fallback is returned instead.
func (c *Config) GetWithDefault(key, fallback string) string {
	c.mu.RLock()
	v := c.values[key]
	c.mu.RUnlock()
	if v == "" {
		return fallback
	}
	return v
}

// GetInt returns the value for key as an integer.
// Returns 0 and an error if the key is absent or cannot be parsed.
func (c *Config) GetInt(key string) (int, error) {
	v := c.Get(key)
	if v == "" {
		return 0, fmt.Errorf("config: key %q not set", key)
	}
	return strconv.Atoi(v)
}

// GetIntWithDefault returns the integer value for key, falling back to def
// if the key is not set or cannot be parsed.
func (c *Config) GetIntWithDefault(key string, def int) int {
	n, err := c.GetInt(key)
	if err != nil {
		return def
	}
	return n
}

// GetBool returns the boolean value for key.
// Truthy strings: "1", "true", "yes", "on" (case-insensitive).
func (c *Config) GetBool(key string) bool {
	v := strings.ToLower(c.Get(key))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// IsProduction reports whether APP_ENV is set to "production".
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Get("APP_ENV")) == "production"
}

// IsDevelopment reports whether APP_ENV is set to "development" or "dev" or
// "local" (common local dev names).
func (c *Config) IsDevelopment() bool {
	v := strings.ToLower(c.Get("APP_ENV"))
	return v == "development" || v == "dev" || v == "local" || v == ""
}

// All returns a copy of every stored key/value pair. The returned map is safe
// to read and modify without affecting the Config.
func (c *Config) All() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]string, len(c.values))
	for k, v := range c.values {
		out[k] = v
	}
	return out
}
