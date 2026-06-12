// Package lookup provides a shared, in-memory registry of "lookup providers"
// that back the formui/lookup autocomplete widget.
//
// Any module can register a provider keyed by a stable name. The base
// module exposes a single HTTP endpoint (GET /base/lookup/:name) that
// proxies the active provider so client-side form widgets can fetch
// suggestions without knowing which module owns the data.
//
// Example registration in a domain module's Register phase:
//
//	if err := baseRepos.Lookup.Register("customer", &CustomerProvider{
//	    Repo: customerRepo,
//	}); err != nil {
//	    return err
//	}
package lookup

import (
	"fmt"
	"strings"
	"sync"
)

// Hit is a single autocomplete suggestion returned to the client.
type Hit struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// Extra is an optional bag of fields the client may want to render
	// (e.g. subtitle, icon, badge). Free-form on purpose.
	Extra map[string]any `json:"extra,omitempty"`
}

// Query carries the parsed request from the formui/lookup widget.
type Query struct {
	// Term is the user's current input.
	Term string
	// Limit caps the number of hits. 0 means "use default".
	Limit int
}

// Provider is implemented by every module that wants to expose a lookup.
type Provider interface {
	// Name returns the registry key. Must be unique.
	Name() string
	// Search returns up to limit hits matching term.
	Search(q Query) ([]Hit, error)
}

// Registry is a thread-safe collection of named lookup providers.
// One instance is created at app boot and shared via the base module's
// Repositories struct.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry constructs an empty registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register adds a provider. Returns an error if the name is empty,
// the provider is nil, or the name is already taken.
func (r *Registry) Register(p Provider) error {
	if p == nil {
		return fmt.Errorf("lookup: nil provider")
	}
	name := strings.TrimSpace(p.Name())
	if name == "" {
		return fmt.Errorf("lookup: provider name is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[name]; ok {
		return fmt.Errorf("lookup: provider %q already registered", name)
	}
	r.providers[name] = p
	return nil
}

// Get returns the provider with the given name, or nil if not found.
func (r *Registry) Get(name string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// Names returns all registered provider names. Useful for diagnostics.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.providers))
	for k := range r.providers {
		out = append(out, k)
	}
	return out
}
