package module

import (
	"context"
	"fmt"

	"github.com/subhasundardas/gofar/framework/services"
)

// Manager orchestrates the lifecycle (Register → Boot) of all registered
// modules. It exposes the shared Context so modules can access the container,
// event bus, Fiber instance, and config.
type Manager struct {
	// Context is the set of shared services available to every module.
	Context *Context

	modules []Module
	names   map[string]struct{} // guard against duplicate module names
}

// NewManager creates a Manager with the provided shared context.
func NewManager(ctx *Context) *Manager {
	return &Manager{
		Context: ctx,
		names:   make(map[string]struct{}),
	}
}

// Add registers a module with the manager. Returns an error if a module with
// the same name has already been added, preventing accidental double-loading.
func (m *Manager) Add(mod Module) error {
	name := mod.Name()
	if _, exists := m.names[name]; exists {
		return fmt.Errorf("module: %q already registered", name)
	}
	m.names[name] = struct{}{}
	m.modules = append(m.modules, mod)
	return nil
}

// MustAdd is like Add but panics on error. Convenient in bootstrap code where
// a duplicate module name is always a programming error.
func (m *Manager) MustAdd(mod Module) {
	if err := m.Add(mod); err != nil {
		panic(err)
	}
}

// RegisterAll calls Register on every module in registration order.
// If any module's Register returns an error, registration stops immediately
// and the error is wrapped with the module name for easy diagnosis.
func (m *Manager) RegisterAll() error {
	for _, mod := range m.modules {
		if err := mod.Register(m); err != nil {
			return fmt.Errorf("module %q register: %w", mod.Name(), err)
		}
	}
	return nil
}

// BootAll calls Boot on every module in registration order.
// All modules have already been Registered when BootAll runs, so it is safe
// to Resolve services registered by other modules.
func (m *Manager) BootAll() error {
	for _, mod := range m.modules {
		if err := mod.Boot(m); err != nil {
			return fmt.Errorf("module %q boot: %w", mod.Name(), err)
		}
	}
	return nil
}

// Names returns the names of all registered modules in registration order.
// Useful for debugging and health-check endpoints.
func (m *Manager) Names() []string {
	names := make([]string, len(m.modules))
	for i, mod := range m.modules {
		names[i] = mod.Name()
	}
	return names
}

// Count returns the total number of registered modules.
func (m *Manager) Count() int { return len(m.modules) }

// ShutdownAll calls Shutdown on every module in reverse registration order
// (last registered is shut down first). This mirrors typical teardown order
// where high-level modules depending on low-level ones stop first.
//
// All modules receive the shutdown signal even if one returns an error;
// errors are collected and returned as a combined message.
func (m *Manager) ShutdownAll(ctx context.Context) error {
	var errs []string
	for i := len(m.modules) - 1; i >= 0; i-- {
		mod := m.modules[i]
		if err := mod.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("module %q shutdown: %v", mod.Name(), err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// Services returns the gofar service registry — the single access point for
// all framework-provided services (DB, Logger, Events, Config, Store, Fiber).
//
// Call this inside Register or Boot:
//
//	db  := mgr.Services().DB()
//	log := mgr.Services().Logger()
func (m *Manager) Services() *services.Registry {
	return services.New(m.Context.Container)
}
