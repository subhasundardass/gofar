package gen

// Every template receives a templateData value as its dot (.) object.
// Fields available in every template:
//   .Module   — Names for the module  (Pascal, Snake, Kebab, Lower, Camel)
//   .ModPath  — Go module path from go.mod  e.g. "github.com/subhasundardas/gofar"

type templateData struct {
	Module  Names
	ModPath string
}

// ── manifest.json ─────────────────────────────────────────────────────────────

var tmplManifest = `{
  "name":        "{{.Module.Snake}}",
  "version":     "1.0.0",
  "description": "{{.Module.Pascal}} module",
  "author":      "",
  "enabled":     true,
  "depends_on": ["base"]
}
`

// ── module.go ─────────────────────────────────────────────────────────────────
//
// Generated modules follow the same pattern as modules/base/module.go and
// modules/accounting/module.go: framework dependencies are obtained from
// mgr.Services() (a *services.Registry) inside Register / Boot, never via
// the raw container. *ent.Client is wired directly into the repository
// layer; no extra Resolve calls are needed.

var tmplModule = `// Package {{.Module.Lower}} is the GoFar module for {{.Module.Pascal}}.
//
// Lifecycle:
//   - Register: bind repositories, services, handlers into the shared container.
//   - Boot:     mount HTTP routes and subscribe to cross-module events.
package {{.Module.Lower}}

import (
	"context"

	fw "github.com/subhasundardas/gofar/framework/gofar"
	"github.com/subhasundardas/gofar/framework/module"
	"github.com/subhasundardas/gofar/modules/{{.Module.Snake}}/handler"
	"github.com/subhasundardas/gofar/modules/{{.Module.Snake}}/repository"
	"github.com/subhasundardas/gofar/modules/{{.Module.Snake}}/service"
)

// Module is the {{.Module.Pascal}} bounded context.
type Module struct {
	*module.BaseModule

	repos    *repository.Repositories
	services *service.Services
	handlers *handler.Handlers
}

// compile-time interface check — fails at build time if Module drifts from the interface.
var _ module.Module = (*Module)(nil)

// New constructs the {{.Module.Pascal}} module. Panics if manifest.json is missing or malformed.
func New() module.Module {
	return &Module{
		BaseModule: module.MustNewBase("./modules/{{.Module.Snake}}/manifest.json"),
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
	m.services = service.NewServices(m.repos, svc.Logger())
	m.handlers = handler.NewHandlers(m.services)

	mgr.Context.Container.Register(m.repos)
	mgr.Context.Container.Register(m.services)
	mgr.Context.Container.Register(m.handlers)

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

// _ keeps the gofar import used in case the user later switches this module
// to the fw.Use(mgr) shorthand; it compiles to nothing at runtime.
var _ = fw.Use
`

// ── routes.go ─────────────────────────────────────────────────────────────────

var tmplRoutes = `// routes.go declares all HTTP routes for the {{.Module.Pascal}} module.
// Called from Module.Boot — do not call directly.
package {{.Module.Lower}}

import (
	"github.com/gofiber/fiber/v2"
	"{{.ModPath}}/modules/{{.Module.Snake}}/handler"
)

// RegisterRoutes mounts {{.Module.Pascal}} endpoints under /{{.Module.Kebab}}.
//
// Route table:
//
//	GET    /{{.Module.Kebab}}          → list
//	POST   /{{.Module.Kebab}}          → create
//	GET    /{{.Module.Kebab}}/:id      → get by ID
//	PUT    /{{.Module.Kebab}}/:id      → update
//	DELETE /{{.Module.Kebab}}/:id      → delete
func RegisterRoutes(app *fiber.App, h *handler.Handlers) {
	grp := app.Group("/{{.Module.Kebab}}")
	grp.Get("/", h.{{.Module.Pascal}}.List)
	grp.Post("/", h.{{.Module.Pascal}}.Create)
	grp.Get("/:id", h.{{.Module.Pascal}}.Get)
	grp.Put("/:id", h.{{.Module.Pascal}}.Update)
	grp.Delete("/:id", h.{{.Module.Pascal}}.Delete)
}
`

// ── events.go ─────────────────────────────────────────────────────────────────

var tmplEvents = `// events.go wires cross-module event subscriptions for the {{.Module.Pascal}} module.
// Called from Module.Boot — do not call directly.
package {{.Module.Lower}}

import (
	"{{.ModPath}}/framework/event"
	"{{.ModPath}}/framework/logger"
	"{{.ModPath}}/modules/{{.Module.Snake}}/service"
)

// RegisterEvents subscribes to events published by other modules.
// Add new subscriptions here as the application grows.
// The event bus dispatches by reflect.Type — use typed structs, never strings.
//
// Example:
//
//	type OrderPlaced struct { OrderID int }
//
//	bus.Subscribe(OrderPlaced{}, func(e any) {
//	    ev := e.(OrderPlaced)
//	    svc.{{.Module.Pascal}}.HandleOrderPlaced(ev.OrderID)
//	})
func RegisterEvents(bus *event.Bus, svc *service.Services, log *logger.Logger) {
	// TODO: subscribe to cross-module events here.
	_ = bus
	_ = svc
	_ = log
}
`

// ── repository/repository.go ──────────────────────────────────────────────────
//
// Matches the real pattern in modules/accounting/repository/repository.go:
//   - NewRepositories(db *ent.Client) wires *ent.Client into every repo.
//   - Repositories satisfies data.Repository[*ent.{{.Module.Pascal}}] via a
//     compile-time interface check so drift is caught at build time.
//   - The local {{.Module.Pascal}} struct is a placeholder for the generated
//     *ent.{{.Module.Pascal}} — swap it once the Ent schema is generated.

var tmplRepository = `// Package repository is the data-access layer for the {{.Module.Pascal}} module.
// All DB queries live here. Services must never call the DB directly.
package repository

import (
	"context"

	"{{.ModPath}}/ent"
	"{{.ModPath}}/framework/data"
)

// Repositories bundles all {{.Module.Pascal}} repositories.
// Pass the whole struct to service.NewServices so adding a repo never changes signatures.
type Repositories struct {
	{{.Module.Pascal}} *{{.Module.Pascal}}Repository
}

// NewRepositories constructs all repositories, wiring the shared *ent.Client.
func NewRepositories(db *ent.Client) *Repositories {
	return &Repositories{
		{{.Module.Pascal}}: New{{.Module.Pascal}}Repository(db),
	}
}

// {{.Module.Pascal}} is the read-model returned by repository queries.
// Replace with the generated ent.{{.Module.Pascal}} type once Ent is wired.
type {{.Module.Pascal}} struct {
	ID   int
	Name string
}

// {{.Module.Pascal}}Repository handles all {{.Module.Pascal}} persistence operations.
type {{.Module.Pascal}}Repository struct {
	db *ent.Client
}

// compile-time check — fails at build if interface is not satisfied.
var _ data.Repository[*ent.{{.Module.Pascal}}] = (*{{.Module.Pascal}}Repository)(nil)

// New{{.Module.Pascal}}Repository constructs a {{.Module.Pascal}}Repository.
func New{{.Module.Pascal}}Repository(db *ent.Client) *{{.Module.Pascal}}Repository {
	return &{{.Module.Pascal}}Repository{db: db}
}

// List returns all {{.Module.Pascal}} records.
func (r *{{.Module.Pascal}}Repository) List(ctx context.Context) ([]*ent.{{.Module.Pascal}}, error) {
	return r.db.{{.Module.Pascal}}.Query().All(ctx)
}

// FindByID returns a single {{.Module.Pascal}} by ID.
func (r *{{.Module.Pascal}}Repository) FindByID(ctx context.Context, id int) (*ent.{{.Module.Pascal}}, error) {
	return r.db.{{.Module.Pascal}}.Get(ctx, id)
}

// Delete removes a {{.Module.Pascal}} by ID.
func (r *{{.Module.Pascal}}Repository) Delete(ctx context.Context, id int) error {
	return r.db.{{.Module.Pascal}}.DeleteOneID(id).Exec(ctx)
}

// TODO: add query methods here (Create, Update, ListPaginated, …).
`

// ── service/service.go ────────────────────────────────────────────────────────
//
// Matches the real pattern in modules/base/service/service.go:
//   - NewServices(repos, logger) — logger is injected so services can log
//     without going through the container.
//   - All service methods take ctx as their first argument so they can be
//     composed into request handlers and background jobs alike.

var tmplService = `// Package service contains business logic for the {{.Module.Pascal}} module.
// Services sit between handlers and repositories — enforce domain rules here.
package service

import (
	"context"
	"fmt"

	"{{.ModPath}}/framework/logger"
	"{{.ModPath}}/modules/{{.Module.Snake}}/repository"
)

// Services bundles all {{.Module.Pascal}} domain services.
type Services struct {
	{{.Module.Pascal}} *{{.Module.Pascal}}Service
}

// NewServices constructs all services, injecting the shared repository group
// and the application logger.
func NewServices(repos *repository.Repositories, logger *logger.Logger) *Services {
	return &Services{
		{{.Module.Pascal}}: New{{.Module.Pascal}}Service(repos.{{.Module.Pascal}}, logger),
	}
}

// {{.Module.Pascal}}Service implements all {{.Module.Pascal}} use-cases.
type {{.Module.Pascal}}Service struct {
	repo   *repository.{{.Module.Pascal}}Repository
	logger *logger.Logger
}

// New{{.Module.Pascal}}Service constructs a {{.Module.Pascal}}Service.
func New{{.Module.Pascal}}Service(repo *repository.{{.Module.Pascal}}Repository, logger *logger.Logger) *{{.Module.Pascal}}Service {
	return &{{.Module.Pascal}}Service{repo: repo, logger: logger}
}

// Create validates input and persists a new {{.Module.Pascal}} record.
func (s *{{.Module.Pascal}}Service) Create(ctx context.Context, name string) (int, error) {
	if name == "" {
		return 0, fmt.Errorf("{{.Module.Snake}}: name is required")
	}
	// TODO: implement once the Ent schema is generated.
	_ = ctx
	return 0, nil
}

// List returns all {{.Module.Pascal}} records.
func (s *{{.Module.Pascal}}Service) List(ctx context.Context) ([]*repository.{{.Module.Pascal}}, error) {
	// TODO: return s.repo.List(ctx)
	_ = ctx
	return []*repository.{{.Module.Pascal}}{}, nil
}

// Get returns a single record. Returns an error if not found.
func (s *{{.Module.Pascal}}Service) Get(ctx context.Context, id int) (*repository.{{.Module.Pascal}}, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("{{.Module.Snake}}: get %d: %w", id, err)
	}
	return item, nil
}

// Update validates and persists changes to an existing record.
func (s *{{.Module.Pascal}}Service) Update(ctx context.Context, id int, name string) error {
	if name == "" {
		return fmt.Errorf("{{.Module.Snake}}: name is required")
	}
	// TODO: implement once the Ent schema is generated.
	_ = ctx
	return nil
}

// Delete permanently removes a record.
func (s *{{.Module.Pascal}}Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
`

// ── handler/handler.go ────────────────────────────────────────────────────────
//
// Matches the real pattern in modules/accounting/handler/handler.go:
//   - Handlers is a bundle that owns all handler sets for the module
//     (e.g. h.Accounting.*, h.Journal.*). The MakeHandler extra appends
//     extra sets to this same bundle.
//   - Handler methods take ctx.Context() / pass ctx down to services.

var tmplHandler = `// Package handler contains HTTP handlers for the {{.Module.Pascal}} module.
// Handlers are thin: parse request → call service → return response.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"{{.ModPath}}/framework/request"
	"{{.ModPath}}/framework/response"
	"{{.ModPath}}/modules/{{.Module.Snake}}/service"
)

// Handlers bundles all {{.Module.Pascal}} handler sets.
type Handlers struct {
	{{.Module.Pascal}} *{{.Module.Pascal}}Handlers
}

// NewHandlers constructs all handler sets. Takes services only — never fiber.App.
func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		{{.Module.Pascal}}: &{{.Module.Pascal}}Handlers{svc: svc.{{.Module.Pascal}}},
	}
}

// {{.Module.Pascal}}Handlers holds all {{.Module.Pascal}} HTTP handlers.
type {{.Module.Pascal}}Handlers struct {
	svc *service.{{.Module.Pascal}}Service
}

// List handles GET /{{.Module.Kebab}}
//   200 OK → { success: true, data: [...] }
func (h *{{.Module.Pascal}}Handlers) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.Context())
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, items)
}

// Create handles POST /{{.Module.Kebab}}
//   Body:         { "name": "..." }
//   201 Created → { success: true, data: { id: N } }
func (h *{{.Module.Pascal}}Handlers) Create(c *fiber.Ctx) error {
	type body struct {
		Name string ` + "`json:\"name\"`" + `
	}
	dto, err := request.Parse[body](c)
	if err != nil {
		return response.Error(c, err)
	}
	id, err := h.svc.Create(c.Context(), dto.Name)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Created(c, fiber.Map{"id": id})
}

// Get handles GET /{{.Module.Kebab}}/:id
//   200 OK → { success: true, data: {...} }
func (h *{{.Module.Pascal}}Handlers) Get(c *fiber.Ctx) error {
	id, err := request.ParamInt(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	item, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, item)
}

// Update handles PUT /{{.Module.Kebab}}/:id
//   Body:   { "name": "..." }
//   200 OK → { success: true, data: { updated: true } }
func (h *{{.Module.Pascal}}Handlers) Update(c *fiber.Ctx) error {
	id, err := request.ParamInt(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	type body struct {
		Name string ` + "`json:\"name\"`" + `
	}
	dto, err := request.Parse[body](c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.svc.Update(c.Context(), id, dto.Name); err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, fiber.Map{"updated": true})
}

// Delete handles DELETE /{{.Module.Kebab}}/:id
//   204 No Content on success
func (h *{{.Module.Pascal}}Handlers) Delete(c *fiber.Ctx) error {
	id, err := request.ParamInt(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return response.Error(c, err)
	}
	return response.NoContent(c)
}
`
