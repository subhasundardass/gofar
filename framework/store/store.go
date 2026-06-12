package store

import "sync"

// Store is a thread-safe, key-value store for global application data.
//
// It lives on [Container] and is available to every module via deps.Store.
// Use it for app-wide settings that need to be shared across modules —
// things like app name, active theme, feature flags, or runtime config
// that is not known at startup.
//
// # When to use Store vs Config
//
//   - [Config] holds static values read from env at boot (DSN, ports, secrets).
//     It never changes after [NewContainer] returns.
//   - Store holds dynamic values that are set at runtime, often by a module
//     during its [Module.PostRegister] or [Module.Init] lifecycle phase,
//     and may be read by other modules later.
//
// # Lifecycle convention
//
// Write to the Store in [Module.PostRegister] (after all modules are wired).
// Read from the Store anywhere after that phase completes.
// Never write to the Store from a request handler — treat it as
// "set once at startup, read many times at runtime."
//
// # Example
//
//	// In your "settings" module PostRegister:
//	m.deps.Store.Set("app_name", "MyApp")
//	m.deps.Store.Set("theme",    "dark")
//
//	// In any other module or handler:
//	name, ok := m.deps.Store.GetString("app_name")
//	if !ok {
//	    // handle missing key
//	}
type Store struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewStore allocates and returns an empty [Store].
// It is called once by [NewContainer]; you should never need to call it directly.
func NewStore() *Store {
	return &Store{
		data: make(map[string]any),
	}
}

// Set stores value under key, overwriting any previous value.
// It is safe to call from multiple goroutines.
func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get returns the value stored under key and whether the key exists.
// It is safe to call from multiple goroutines.
//
//	v, ok := store.Get("theme")
//	if !ok {
//	    // key was never set
//	}
func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// MustGet returns the value stored under key.
// It panics if the key does not exist, so only use it for keys that are
// guaranteed to be set during module initialisation.
//
//	theme := store.MustGet("theme").(string)
func (s *Store) MustGet(key string) any {
	v, ok := s.Get(key)
	if !ok {
		panic("store: key not found: " + key)
	}
	return v
}

// GetString returns the value for key as a string.
// The second return value is false if the key does not exist or if the
// stored value is not a string.
//
//	name, ok := store.GetString("app_name")
func (s *Store) GetString(key string) (string, bool) {
	v, ok := s.Get(key)
	if !ok {
		return "", false
	}
	str, ok := v.(string)
	return str, ok
}

// GetBool returns the value for key as a bool.
// The second return value is false if the key does not exist or if the
// stored value is not a bool.
//
//	enabled, ok := store.GetBool("feature_x")
func (s *Store) GetBool(key string) (bool, bool) {
	v, ok := s.Get(key)
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// GetInt returns the value for key as an int.
// The second return value is false if the key does not exist or if the
// stored value is not an int.
//
//	limit, ok := store.GetInt("max_upload_size")
func (s *Store) GetInt(key string) (int, bool) {
	v, ok := s.Get(key)
	if !ok {
		return 0, false
	}
	i, ok := v.(int)
	return i, ok
}

// Delete removes key from the store. It is a no-op if the key does not exist.
// It is safe to call from multiple goroutines.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Keys returns a snapshot of all keys currently in the store.
// The order is not guaranteed.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}
