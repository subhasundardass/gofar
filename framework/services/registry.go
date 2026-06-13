// Package services provides the gofar.Services accessor — a single entry
// point for every framework-provided service a module may need.
//
// Usage inside any module's Register or Boot:
//
//	db     := mgr.Services().DB()
//	log    := mgr.Services().Logger()
//	events := mgr.Services().Events()
//	cfg    := mgr.Services().Config()
//	store  := mgr.Services().Store()
package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/framework/config"
	"github.com/subhasundardas/gofar/framework/container"
	"github.com/subhasundardas/gofar/framework/event"
	"github.com/subhasundardas/gofar/framework/logger"
	"github.com/subhasundardas/gofar/framework/store"
)

// Registry exposes all framework-provided services to modules.
// Obtain one via mgr.Services() inside Register or Boot.
type Registry struct {
	*BaseService
	c *container.Container
}

// New creates a Registry backed by the given container.
// Normally called only by module.Manager.Services().
func New(c *container.Container) *Registry {
	return &Registry{
		c: c}
}

// DB returns the ent ORM client. Panics if not registered (always available
// after bootstrap).
func (r *Registry) DB() *ent.Client {
	return container.MustResolve[*ent.Client](r.c)
}

// Logger returns the application-level structured logger.
func (r *Registry) Logger() *logger.Logger {
	return container.MustResolve[*logger.Logger](r.c)
}

// Events returns the in-process event bus.
func (r *Registry) Events() *event.Bus {
	return container.MustResolve[*event.Bus](r.c)
}

// Config returns the application configuration.
func (r *Registry) Config() *config.Config {
	return container.MustResolve[*config.Config](r.c)
}

// Store returns the application-wide key/value store.
func (r *Registry) Store() *store.Store {
	return container.MustResolve[*store.Store](r.c)
}

// Fiber returns the root Fiber app. Use in Boot to register routes or
// middleware when you need the raw instance; prefer mgr.Context.Fiber for
// route groups.
func (r *Registry) Fiber() *fiber.App {
	return container.MustResolve[*fiber.App](r.c)
}
