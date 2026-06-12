package module

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/subhasundardas/gofar/framework/logger"
)

// ── Meta ─────────────────────────────────────────────────────────────────────

// Meta holds the static metadata declared in a module's manifest.json.
// Every field is optional except Name.
//
// Example manifest.json:
//
//	{
//	  "name":        "accounting",
//	  "version":     "1.0.0",
//	  "description": "Manages invoices, payments and ledger entries",
//	  "author":      "Subha Sundar Das",
//	  "enabled":     true
//	}
type Meta struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	// Enabled gates the module at bootstrap. If false the module is skipped
	// entirely during Register and Boot. Defaults to true when omitted.
	Enabled   *bool    `json:"enabled"`
	DependsOn []string `json:"depends_on"`
}

// IsEnabled returns true unless the manifest explicitly sets "enabled": false.
func (m Meta) IsEnabled() bool {
	return m.Enabled == nil || *m.Enabled
}

// ── BaseModule ────────────────────────────────────────────────────────────────

// BaseModule is an embeddable struct that provides:
//   - Module metadata loaded from a manifest.json
//   - A named logger scoped to the module
//   - A graceful-shutdown context that is cancelled when Shutdown() is called
//   - Default no-op implementations of Register, Boot and Shutdown so a
//     concrete module only needs to override the methods it actually uses
//
// Embed it in your module struct:
//
//	type AccountingModule struct {
//	    *module.BaseModule
//	}
//
// Then call LoadMeta in New() to populate metadata and create the logger:
//
//	func New() module.Module {
//	    b, err := module.NewBase("./modules/example/manifest.json")
//	    if err != nil { panic(err) }
//	    return &AccountingModule{BaseModule: b}
//	}
type BaseModule struct {
	Meta   Meta
	Logger *logger.Logger

	shutdownOnce sync.Once
	shutdownCh   chan struct{}
}

// NewBase creates a BaseModule by loading metadata from the given manifest path.
// It initialises the module logger with the module's name as context.
// Returns an error if the file cannot be read or contains invalid JSON.
func NewBase(manifestPath string) (*BaseModule, error) {
	b := &BaseModule{
		shutdownCh: make(chan struct{}),
	}
	if err := b.LoadMeta(manifestPath); err != nil {
		return nil, err
	}
	return b, nil
}

// MustNewBase is like NewBase but panics on error. Use in init() or New()
// where a bad manifest is always a programming error.
func MustNewBase(manifestPath string) *BaseModule {
	b, err := NewBase(manifestPath)
	if err != nil {
		panic(fmt.Sprintf("module: failed to load manifest %q: %v", manifestPath, err))
	}
	return b
}

// LoadMeta reads the manifest JSON at path and populates b.Meta.
// It also (re)initialises the scoped logger once the name is known.
// Returns an error if the file is missing, unreadable, or malformed.
func (b *BaseModule) LoadMeta(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("module: read manifest %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &b.Meta); err != nil {
		return fmt.Errorf("module: parse manifest %q: %w", path, err)
	}
	if b.Meta.Name == "" {
		return fmt.Errorf("module: manifest %q is missing required field \"name\"", path)
	}
	// Initialise a logger scoped to this module's name.
	b.Logger = logger.New(logger.LevelDebug, false).With("module", b.Meta.Name)
	return nil
}

// Name returns the module's name as declared in manifest.json.
// Satisfies the Module interface.
func (b *BaseModule) Name() string { return b.Meta.Name }

// Register is a no-op default. Override in your concrete module when you need
// to bind services into the container.
func (b *BaseModule) Register(_ *Manager) error { return nil }

// Boot is a no-op default. Override in your concrete module when you need to
// wire routes, subscribe to events, or start background workers.
func (b *BaseModule) Boot(_ *Manager) error { return nil }

// Shutdown signals the module to release resources. It is safe to call
// multiple times – only the first call has any effect.
// Override in your concrete module to add teardown logic, then call
// b.BaseModule.Shutdown(ctx) at the end.
func (b *BaseModule) Shutdown(_ context.Context) error {
	b.shutdownOnce.Do(func() {
		close(b.shutdownCh)
		if b.Logger != nil {
			b.Logger.Infof("%s module shutdown", b.Meta.Name)
		}
	})
	return nil
}

// Done returns a channel that is closed when Shutdown() has been called.
// Use this to signal background goroutines to stop:
//
//	select {
//	case <-m.Done():
//	    return
//	case job := <-jobs:
//	    process(job)
//	}
func (b *BaseModule) Done() <-chan struct{} { return b.shutdownCh }

// ── Global registry ───────────────────────────────────────────────────────────

// Factory is a function that constructs and returns a Module instance.
// The registry stores factories rather than instances so each bootstrap
// creates fresh module objects.
type Factory func() Module

var (
	registryMu sync.RWMutex
	registry   []Factory
)

// Register adds a module factory to the global registry.
// Call this from each module package's init() function:
//
//	func init() {
//	    module.Register(New)
//	}
//
// Modules are booted in the order they were registered.
func Register(factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(registry, factory)
}

// LoadFromRegistry instantiates every registered module factory and adds
// the resulting modules to mgr. Returns the first error encountered.
// Call this once during application bootstrap, after building the Manager:
//
//	mgr := module.NewManager(ctx)
//	if err := mgr.LoadFromRegistry(); err != nil {
//	    log.Fatal(err)
//	}
func (mgr *Manager) LoadFromRegistry() error {
	registryMu.RLock()
	factories := make([]Factory, len(registry))
	copy(factories, registry)
	registryMu.RUnlock()

	for _, factory := range factories {
		mod := factory()
		// Honour the manifest "enabled" flag – skip disabled modules silently.
		if bm, ok := mod.(interface{ GetMeta() Meta }); ok {
			if !bm.GetMeta().IsEnabled() {
				continue
			}
		}
		if err := mgr.Add(mod); err != nil {
			return err
		}
	}

	fmt.Println("Module Count:", mgr.Count())
	fmt.Println("Modules:", mgr.Names())
	return nil
}

// GetMeta exposes the embedded Meta so LoadFromRegistry can inspect it.
// Concrete modules that embed *BaseModule get this for free.
func (b *BaseModule) GetMeta() Meta { return b.Meta }
