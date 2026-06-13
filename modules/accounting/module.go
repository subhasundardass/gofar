// Package accounting is the GoFar module for Accounting.
//
// Lifecycle:
//   - Register: bind repositories, services, handlers into the container.
//   - Boot:     mount HTTP routes, subscribe to events, and run Seed
//     when APP_ENV is not "production".
package accounting

import (
	"context"

	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/framework/container"
	fw "github.com/subhasundardas/gofar/framework/gofar"
	"github.com/subhasundardas/gofar/framework/module"
	"github.com/subhasundardas/gofar/modules/accounting/handler"
	"github.com/subhasundardas/gofar/modules/accounting/repository"
	"github.com/subhasundardas/gofar/modules/accounting/service"
)

// Module is the Accounting bounded context.
type Module struct {
	*module.BaseModule

	repos    *repository.Repositories
	services *service.Services
	handlers *handler.Handlers
}

// compile-time interface check — fails at build time if Module drifts from the interface.
var _ module.Module = (*Module)(nil)

// New constructs the Accounting module. Panics if manifest.json is missing or malformed.
func New() module.Module {
	return &Module{
		BaseModule: module.MustNewBase("./modules/accounting/manifest.json"),
	}
}

// init self-registers the module factory with the global registry.
// Triggered once when the binary imports this package (via modules/all).
func init() {
	module.Register(New)
}

func (m *Module) Register(mgr *module.Manager) error {

	// Init Gofar and its services
	gofar := fw.Use(mgr)
	svc := gofar.Services

	// svc.Logger().Error("adas")
	// svc.Config().Get("aaa")
	// svc.DB()
	// svc.Store().Set("name", "Subha")

	m.repos = repository.NewRepositories(svc.DB())
	m.services = service.NewServices(m.repos)
	m.handlers = handler.NewHandlers(m.services)

	svc.Logger().Infof("%s: registered", m.Name())
	return nil
}

// Boot mounts HTTP routes, subscribes to events, and (in non-production
// environments) seeds default reference data such as countries, states,
// account groups, and ledgers.
//
// All modules are Registered before Boot runs — cross-module Resolve is
// safe here.
func (m *Module) Boot(mgr *module.Manager) error {
	m.Logger.Info("booting")
	RegisterRoutes(mgr.Context.Fiber, m.handlers)
	RegisterEvents(mgr.Context.Events, m.services, m.Logger)

	// Seed default reference data unless we are running in production.
	if !mgr.Context.Config.IsProduction() {
		db := container.MustResolve[*ent.Client](mgr.Context.Container)
		if err := Seed(context.Background(), db); err != nil {
			m.Logger.Errorf("accounting seed: %v", err)
			return err
		}
		m.Logger.Info("accounting seed complete")
	}

	return nil
}

// Shutdown releases module resources. Called in reverse registration order.
func (m *Module) Shutdown(ctx context.Context) error {
	m.Logger.Info("shutting down")
	return m.BaseModule.Shutdown(ctx)
}
