// Package example is the GoFar module for Example.
//
// Lifecycle:
//   - Register: bind repositories, services, handlers into the container.
//   - Boot:     mount HTTP routes and subscribe to cross-module events.
package example

import (
	"context"

	"github.com/subhasundardas/gofar/framework/module"
	"github.com/subhasundardas/gofar/modules/example/handler"
	"github.com/subhasundardas/gofar/modules/example/repository"
	"github.com/subhasundardas/gofar/modules/example/service"
)

// Module is the Example bounded context.
type Module struct {
	*module.BaseModule

	repos    *repository.Repositories
	services *service.Services
	handlers *handler.Handlers
}

// compile-time interface check — fails at build time if Module drifts from the interface.
var _ module.Module = (*Module)(nil)

// New constructs the Example module. Panics if manifest.json is missing or malformed.
func New() module.Module {
	return &Module{
		BaseModule: module.MustNewBase("./modules/example/manifest.json"),
	}
}

// init self-registers the module factory with the global registry.
// Triggered once when the binary imports this package (via modules/all).
func init() {
	module.Register(New)
}

// Register binds infrastructure into the shared container.
// Called before Boot — do NOT resolve cross-module services here.
func (m *Module) Register(mgr *module.Manager) error {
	m.Logger.Info("registering")

	// Wire: repo → service → handler (each layer depends on the one below).
	// Uncomment and replace the db arg once your Ent client is wired:
	//   db := container.MustResolve[*ent.Client](mgr.Context.Container)
	m.repos = repository.NewRepositories( /* db */ )
	mgr.Context.Container.Register(m.repos)

	m.services = service.NewServices(m.repos)
	mgr.Context.Container.Register(m.services)

	m.handlers = handler.NewHandlers(m.services)
	mgr.Context.Container.Register(m.handlers)

	return nil
}

// Boot mounts HTTP routes and subscribes to events.
// All modules are Registered before Boot runs — cross-module Resolve is safe here.
func (m *Module) Boot(mgr *module.Manager) error {
	m.Logger.Info("booting")

	RegisterRoutes(mgr.Context.Fiber, m.handlers)
	RegisterEvents(mgr.Context.Events, m.services, m.Logger)
	return nil
}

// Shutdown releases module resources. Called in reverse registration order.
func (m *Module) Shutdown(ctx context.Context) error {
	m.Logger.Info("shutting down")
	return m.BaseModule.Shutdown(ctx)
}
