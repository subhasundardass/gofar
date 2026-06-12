package app

import (
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver registered for database/sql

	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/framework/config"
	"github.com/subhasundardas/gofar/framework/container"
	"github.com/subhasundardas/gofar/framework/database"
	"github.com/subhasundardas/gofar/framework/event"
	"github.com/subhasundardas/gofar/framework/logger"
	"github.com/subhasundardas/gofar/framework/middleware"
	"github.com/subhasundardas/gofar/framework/module"
	"github.com/subhasundardas/gofar/framework/store"
)

// Bootstrap wires the application in the order the rest of the framework
// expects.
//
// The order matters and is intentional:
//
//  1. loadConfig          — everything else reads from Config.
//  2. setupStore          — modules may read/write the Store in Register/Boot.
//  3. setupEvents         — modules may subscribe in Register/Boot.
//  4. setupLogger         — middleware + modules need a logger to log to.
//  5. setupFiber          — middleware and routes are attached to the Fiber app.
//  6. setupMiddleware     — attaches global handlers (recovery, CORS, timing).
//  7. setupDatabase       — opens the *ent.Client and registers it in the
//     container. MUST run before setupModules because modules resolve
//     *ent.Client during their Register phase.
//  8. setupModules        — Register + Boot all modules. They see the fully
//     built Container (including *ent.Client), Store, Events, Logger, and Fiber.
//  9. registerCoreServices — runs auto-migration against the ent schema.
//
// setupStore used to run AFTER setupModules, which meant module.Context.Store
// was nil when modules loaded. That bug is fixed here.
func (a *App) Bootstrap() error {
	if err := a.loadConfig(); err != nil {
		return err
	}

	if err := a.setupStore(); err != nil {
		return err
	}

	if err := a.setupEvents(); err != nil {
		return err
	}

	if err := a.setupLogger(); err != nil {
		return err
	}

	if err := a.setupFiber(); err != nil {
		return err
	}

	if err := a.setupMiddleware(); err != nil {
		return err
	}

	// Database must be wired into the container before modules load —
	// their Register methods resolve *ent.Client directly from the
	// container.
	if err := a.setupDatabase(); err != nil {
		return err
	}

	if err := a.setupModules(); err != nil {
		return err
	}

	return a.registerCoreServices()
}

// loadConfig builds a *Config, optionally loads a .env file from the working
// directory, and pre-seeds the framework's well-known keys with their
// defaults via Config.Load (which prefers the OS environment over the
// fallback). This is the single place to add new "framework-known" defaults.
func (a *App) loadConfig() error {
	cfg := config.New()

	// .env is optional; LoadEnvFile silently ignores a missing file.
	if err := cfg.LoadEnvFile(".env"); err != nil {
		return fmt.Errorf("config: load .env: %w", err)
	}

	// Pre-seed well-known keys. Config.Load reads OS env first and only
	// falls back to the supplied default, so real environment variables
	// always win.
	cfg.Load("APP_NAME", "GoFar")
	cfg.Load("APP_ENV", "development")
	cfg.Load("DB_DRIVER", "sqlite")
	cfg.Load("DB_DSN", "./gofar.db?_fk=1")

	a.Container.Register(cfg)
	return nil
}

func (a *App) setupStore() error {
	a.Container.Register(store.NewStore())
	return nil
}

func (a *App) setupEvents() error {
	a.Container.Register(event.New())
	return nil
}

func (a *App) setupLogger() error {
	a.Container.Register(logger.Default())
	return nil
}

func (a *App) setupFiber() error {
	app := fiber.New(fiber.Config{
		AppName:      a.Config().GetWithDefault("APP_NAME", "GoFar"),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// static files
	app.Static("/assets", "./static/assets")
	app.Static("/uploads", "./static/uploads")

	// a no-op global middleware slot for future cross-cutting concerns
	// (request id, auth injection, etc.) — kept last so user middleware in
	// setupMiddleware can run before/after it.
	app.Use(func(c *fiber.Ctx) error {
		return c.Next()
	})

	a.Container.Register(app)
	return nil
}

func (a *App) setupMiddleware() error {
	// middleware.New takes the logger.Logger struct by value (the
	// framework/logger package has no Logger interface, so the parameter
	// is the concrete struct). Dereference the pointer returned by
	// a.Logger() and pass the value through.
	mw := middleware.New(a.Fiber(), a.Config(), *a.Logger())
	a.Container.Register(mw)
	return nil
}

func (a *App) setupModules() error {
	ctx := &module.Context{
		Container: a.Container,
		Config:    a.Config(),
		Events:    a.Events(),
		Fiber:     a.Fiber(),
		Logger:    a.Logger(),
		Store:     a.Store(),
	}

	manager := module.NewManager(ctx)
	if err := manager.LoadFromRegistry(); err != nil {
		return err
	}
	a.Logger().Infof("loaded %d module(s): %v", manager.Count(), manager.Names())

	if err := manager.RegisterAll(); err != nil {
		return err
	}

	if err := manager.BootAll(); err != nil {
		return err
	}

	a.Container.Register(manager)

	return nil
}

// setupDatabase opens the configured database, registers both *database.DB
// and the underlying *ent.Client in the container, and logs the driver in
// use. Modules can resolve the ent client via
// container.MustResolve[*ent.Client](mgr.Context.Container).
//
// To use a different database, set DB_DRIVER (e.g. "postgres") and DB_DSN
// in your environment or .env file. The matching driver must be imported
// somewhere in the binary (see the blank imports above for sqlite).
func (a *App) setupDatabase() error {
	if container.Has[*ent.Client](a.Container) {
		return fmt.Errorf("setupDatabase called twice — database already registered")
	}

	driver := a.Config().GetWithDefault("DB_DRIVER", "sqlite")
	dsn := a.Config().GetWithDefault("DB_DSN", "./gofar.db?_fk=1")

	db, err := database.New(driver, dsn)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	a.Container.Register(db)
	a.Container.Register(db.GetClient())

	a.Logger().Infof("database connected (driver=%s)", driver)
	return nil
}

// registerCoreServices runs ent's auto-migration against the live database.
// Idempotent: ent only creates tables / columns that don't already exist.
func (a *App) registerCoreServices() error {
	client, err := container.Resolve[*ent.Client](a.Container)
	if err != nil {
		return fmt.Errorf("registerCoreServices: resolve ent client: %w", err)
	}

	if err := client.Schema.Create(a.Ctx); err != nil {
		return fmt.Errorf("ent auto-migration failed: %w", err)
	}
	a.Logger().Info("ent auto-migration complete")

	return nil
}
