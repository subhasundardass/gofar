// Package base is the GoFar module for Base.
//
// Lifecycle:
//   - Register: bind repositories, services, handlers into the container.
//   - Boot:     mount HTTP routes and subscribe to cross-module events.
package base

import (
	"context"

	"github.com/subhasundardas/gofar/framework/module"
	"github.com/subhasundardas/gofar/modules/base/handler"
	"github.com/subhasundardas/gofar/modules/base/repository"
	"github.com/subhasundardas/gofar/modules/base/service"
)

// Module is the Base bounded context.
type Module struct {
	*module.BaseModule

	repos    *repository.Repositories
	services *service.Services
	handlers *handler.Handlers
}

// compile-time interface check — fails at build time if Module drifts from the interface.
var _ module.Module = (*Module)(nil)

// New constructs the Base module. Panics if manifest.json is missing or malformed.
func New() module.Module {
	return &Module{
		BaseModule: module.MustNewBase("./modules/base/manifest.json"),
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
	svc := mgr.Services()

	m.repos = repository.NewRepositories(svc.DB())
	// container.Register(mgr.Context.Container, m.repos)

	m.services = service.NewServices(m.repos, svc.Logger())
	// container.Register(mgr.Context.Container, m.services)

	m.handlers = handler.NewHandlers(m.services)
	// container.Register(mgr.Context.Container, m.handlers)

	svc.Logger().Infof("%s: registered", m.Name())
	return nil
}

// Boot mounts HTTP routes and subscribes to events.
// All modules are Registered before Boot runs — cross-module Resolve is safe here.
func (m *Module) Boot(mgr *module.Manager) error {
	svc := mgr.Services()

	RegisterRoutes(mgr.Context.Fiber, m.handlers)
	RegisterEvents(svc.Events(), m.services, svc.Logger())

	svc.Logger().Infof("%s: booted", m.Name())
	return nil
}

// Shutdown releases module resources. Called in reverse registration order.
func (m *Module) Shutdown(ctx context.Context) error {
	m.Logger.Info("shutting down")
	return m.BaseModule.Shutdown(ctx)
}
