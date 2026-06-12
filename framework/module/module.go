// Package module defines GoFar's modular application architecture.
//
// A Module is a self-contained unit of functionality (e.g. "accounting",
// "auth", "notifications"). Every module goes through two lifecycle phases:
//
//  1. Register – bind services, repositories, and event listeners into the
//     container. At this point other modules' services may not be available
//     yet, so do not call Resolve on foreign types here.
//
//  2. Boot – wire routes, run migrations, start background workers, etc.
//     By the time Boot runs every module has already been Registered, so it
//     is safe to Resolve services from other modules.
//
// Implementing a module:
//
//	type AccountingModule struct{}
//
//	func (m *AccountingModule) Name() string { return "accounting" }
//
//	func (m *AccountingModule) Register(mgr *module.Manager) error {
//	    mgr.Container.Register(accounting.NewRepository(mgr.Container))
//	    return nil
//	}
//
//	func (m *AccountingModule) Boot(mgr *module.Manager) error {
//	    router := mgr.Fiber.Group("/accounting")
//	    accounting.RegisterRoutes(router, mgr.Container)
//	    return nil
//	}
package module

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/config"
	"github.com/subhasundardas/gofar/framework/container"
	"github.com/subhasundardas/gofar/framework/event"
	"github.com/subhasundardas/gofar/framework/logger"
	"github.com/subhasundardas/gofar/framework/store"
)

// Context carries the shared framework services that every module has access
// to during its Register and Boot phases.
type Context struct {
	// Container is the application-wide service registry.
	Container *container.Container
	// Events is the in-process event bus.
	Events *event.Bus
	// Fiber is the root Fiber application, used in Boot to register routes.
	Fiber *fiber.App
	// Config exposes application configuration.
	Config *config.Config
	// Logger is the application-level logger. Modules typically use their own
	// scoped logger from BaseModule.Logger instead.
	Logger *logger.Logger

	// App Store
	Store *store.Store
}

// Module is the interface every GoFar module must satisfy.
//
// Lifecycle order guaranteed by the Manager:
//
//	Register (all modules) → Boot (all modules) → [running] → Shutdown (all modules, reverse order)
type Module interface {
	// Name returns a unique, stable identifier for the module (e.g. "auth").
	// Used for logging, dependency ordering, and health-check output.
	Name() string

	// Register is called during the first boot pass. Use it to bind services,
	// repositories, and infrastructure adapters into the container.
	// At this point other modules' Boot has NOT run, so do not resolve
	// cross-module services here.
	Register(mgr *Manager) error

	// Boot is called after ALL modules have been Registered. Use it to wire
	// HTTP routes, subscribe to events, and start background workers.
	// Cross-module Resolve calls are safe here.
	Boot(mgr *Manager) error

	// Shutdown is called when the application is gracefully stopping.
	// Release goroutines, flush buffers, close connections.
	// Modules are shut down in reverse registration order.
	Shutdown(ctx context.Context) error
}
