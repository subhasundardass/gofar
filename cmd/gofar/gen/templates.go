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

var tmplModule = `// Package {{.Module.Lower}} is the GoFar module for {{.Module.Pascal}}.
//
// Lifecycle:
//   - Register: bind repositories, services, handlers into the container.
//   - Boot:     mount HTTP routes and subscribe to cross-module events.
package {{.Module.Lower}}

import (
	"context"

	"{{.ModPath}}/framework/module"
	"{{.ModPath}}/modules/{{.Module.Snake}}/handler"
	"{{.ModPath}}/modules/{{.Module.Snake}}/repository"
	"{{.ModPath}}/modules/{{.Module.Snake}}/service"
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

var tmplRepository = `// Package repository is the data-access layer for the {{.Module.Pascal}} module.
// All DB queries live here. Services must never call the DB directly.
package repository

// Repositories bundles all {{.Module.Pascal}} repositories.
// Pass the whole struct to service.NewServices so adding a repo never changes signatures.
type Repositories struct {
	{{.Module.Pascal}} *{{.Module.Pascal}}Repository
}

// NewRepositories constructs all repositories.
// Replace the comment with *ent.Client once your schema is generated:
//
//	func NewRepositories(db *ent.Client) *Repositories {
//	    return &Repositories{ {{.Module.Pascal}}: New{{.Module.Pascal}}Repository(db) }
//	}
func NewRepositories( /* db *ent.Client */ ) *Repositories {
	return &Repositories{
		{{.Module.Pascal}}: New{{.Module.Pascal}}Repository(),
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
	// db *ent.Client
}

// New{{.Module.Pascal}}Repository constructs a {{.Module.Pascal}}Repository.
func New{{.Module.Pascal}}Repository( /* db *ent.Client */ ) *{{.Module.Pascal}}Repository {
	return &{{.Module.Pascal}}Repository{}
}

// Create persists a new record and returns its generated ID.
func (r *{{.Module.Pascal}}Repository) Create(name string) (int, error) {
	// TODO: r.db.{{.Module.Pascal}}.Create().SetName(name).Save(ctx)
	return 0, nil
}

// List returns all records.
func (r *{{.Module.Pascal}}Repository) List() ([]{{.Module.Pascal}}, error) {
	// TODO: r.db.{{.Module.Pascal}}.Query().All(ctx)
	return []{{.Module.Pascal}}{}, nil
}

// FindByID returns one record or nil if not found.
func (r *{{.Module.Pascal}}Repository) FindByID(id int) (*{{.Module.Pascal}}, error) {
	// TODO: r.db.{{.Module.Pascal}}.Get(ctx, id)
	return nil, nil
}

// Update persists changes to an existing record.
func (r *{{.Module.Pascal}}Repository) Update(id int, name string) error {
	// TODO: r.db.{{.Module.Pascal}}.UpdateOneID(id).SetName(name).Exec(ctx)
	return nil
}

// Delete hard-deletes a record.
func (r *{{.Module.Pascal}}Repository) Delete(id int) error {
	// TODO: r.db.{{.Module.Pascal}}.DeleteOneID(id).Exec(ctx)
	return nil
}
`

// ── service/service.go ────────────────────────────────────────────────────────

var tmplService = `// Package service contains business logic for the {{.Module.Pascal}} module.
// Services sit between handlers and repositories — enforce domain rules here.
package service

import (
	"fmt"

	"{{.ModPath}}/modules/{{.Module.Snake}}/repository"
)

// Services bundles all {{.Module.Pascal}} domain services.
type Services struct {
	{{.Module.Pascal}} *{{.Module.Pascal}}Service
}

// NewServices constructs all services, injecting the shared repository group.
func NewServices(repos *repository.Repositories) *Services {
	return &Services{
		{{.Module.Pascal}}: New{{.Module.Pascal}}Service(repos.{{.Module.Pascal}}),
	}
}

// {{.Module.Pascal}}Service implements all {{.Module.Pascal}} use-cases.
type {{.Module.Pascal}}Service struct {
	repo *repository.{{.Module.Pascal}}Repository
}

// New{{.Module.Pascal}}Service constructs a {{.Module.Pascal}}Service.
func New{{.Module.Pascal}}Service(repo *repository.{{.Module.Pascal}}Repository) *{{.Module.Pascal}}Service {
	return &{{.Module.Pascal}}Service{repo: repo}
}

// Create validates input and persists a new {{.Module.Pascal}} record.
func (s *{{.Module.Pascal}}Service) Create(name string) (int, error) {
	if name == "" {
		return 0, fmt.Errorf("{{.Module.Snake}}: name is required")
	}
	return s.repo.Create(name)
}

// List returns all {{.Module.Pascal}} records.
func (s *{{.Module.Pascal}}Service) List() ([]repository.{{.Module.Pascal}}, error) {
	return s.repo.List()
}

// Get returns a single record. Returns an error if not found.
func (s *{{.Module.Pascal}}Service) Get(id int) (*repository.{{.Module.Pascal}}, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("{{.Module.Snake}}: get %d: %w", id, err)
	}
	if item == nil {
		return nil, fmt.Errorf("{{.Module.Snake}}: %d not found", id)
	}
	return item, nil
}

// Update validates and persists changes to an existing record.
func (s *{{.Module.Pascal}}Service) Update(id int, name string) error {
	if name == "" {
		return fmt.Errorf("{{.Module.Snake}}: name is required")
	}
	return s.repo.Update(id, name)
}

// Delete permanently removes a record.
func (s *{{.Module.Pascal}}Service) Delete(id int) error {
	return s.repo.Delete(id)
}
`

// ── handler/handler.go ────────────────────────────────────────────────────────

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
	items, err := h.svc.List()
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
	id, err := h.svc.Create(dto.Name)
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
	item, err := h.svc.Get(id)
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
	if err := h.svc.Update(id, dto.Name); err != nil {
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
	if err := h.svc.Delete(id); err != nil {
		return response.Error(c, err)
	}
	return response.NoContent(c)
}
`
