package app

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/config"
	"github.com/subhasundardas/gofar/framework/container"
	"github.com/subhasundardas/gofar/framework/event"
	"github.com/subhasundardas/gofar/framework/logger"
	"github.com/subhasundardas/gofar/framework/module"
	"github.com/subhasundardas/gofar/framework/store"
)

type App struct {
	Container *container.Container
	Ctx       context.Context
	Cancel    context.CancelFunc
}

func New() (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		Container: container.New(),
		Ctx:       ctx,
		Cancel:    cancel,
	}

	if err := app.Bootstrap(); err != nil {
		cancel()
		return nil, err
	}

	return app, nil
}

func (a *App) Run(addr string) error {
	return a.Fiber().Listen(addr)
}

// Shutdown gracefully tears the application down. Order matters:
//
//  1. Stop accepting new connections (Fiber.Shutdown) — drains in-flight
//     requests and returns.
//  2. Shut every module down in reverse registration order so high-level
//     modules that depend on lower-level ones stop first.
//  3. Cancel the root context as a last resort so anything still blocked
//     on it unblocks.
//
// Module.Shutdown is the right place for a module to flush buffers, close
// DB connections, stop background workers, etc.
func (a *App) Shutdown() error {
	var firstErr error

	if fiber := a.Fiber(); fiber != nil {
		if err := fiber.Shutdown(); err != nil {
			firstErr = fmt.Errorf("app: fiber shutdown: %w", err)
		}
	}

	if mgr := a.Modules(); mgr != nil {
		// Manager.ShutdownAll collects every module error and returns them
		// as a single combined error; we surface the first non-nil one.
		if err := mgr.ShutdownAll(a.Ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	a.Cancel()
	return firstErr
}

func (a *App) Fiber() *fiber.App {
	app, _ := container.Resolve[*fiber.App](a.Container)
	return app
}

func (a *App) Config() *config.Config {
	cfg, _ := container.Resolve[*config.Config](a.Container)
	return cfg
}

func (a *App) Events() *event.Bus {
	bus, _ := container.Resolve[*event.Bus](a.Container)
	return bus
}

func (a *App) Store() *store.Store {
	store, _ := container.Resolve[*store.Store](a.Container)
	return store
}

func (a *App) Modules() *module.Manager {
	manager, _ := container.Resolve[*module.Manager](a.Container)
	return manager
}

func (a *App) Context() context.Context {
	return a.Ctx
}

func (a *App) Logger() *logger.Logger {
	l, err := container.Resolve[*logger.Logger](a.Container)
	if err != nil {
		panic(err)
	}
	return l
}
