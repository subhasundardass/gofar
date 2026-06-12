# GoFar — A Modular Go Web Framework Built for Clarity

> **GoFar** is a modular, opinionated Go web framework built on top of [Fiber](https://gofiber.io), [Ent ORM](https://entgo.io), [Templ](https://templ.guide), and [Datastar](https://data-star.dev). It gives you a structured way to build maintainable web applications by organizing code into self-contained modules — each with its own repositories, services, handlers, and routes.

---

## Why GoFar?

Go is fast, simple, and powerful — but large Go web projects can quickly become a mess of scattered files with no clear structure. GoFar solves this by:

- Enforcing a **module-based architecture** — every feature lives in its own bounded context
- Providing a **single access point** (`svc`) for all framework services inside any module
- Including a **CLI generator** so you never write boilerplate by hand
- Keeping **stateless utilities** (forms, errors, rendering) as plain imports — no magic
- Being **fully testable** — every layer can be mocked without spinning up the framework

---

## Table of Contents

1. [Project Structure](#1-project-structure)
2. [Bootstrap & Lifecycle](#2-bootstrap--lifecycle)
3. [Modules](#3-modules)
4. [The gofar Facade](#4-the-gofar-facade)
5. [Services Registry (svc)](#5-services-registry-svc)
6. [Repositories & Data Layer](#6-repositories--data-layer)
7. [Services Layer](#7-services-layer)
8. [Handlers](#8-handlers)
9. [Routing](#9-routing)
10. [Forms](#10-forms)
11. [Request & Response](#11-request--response)
12. [Rendering with Templ](#12-rendering-with-templ)
13. [Datastar & SSE](#13-datastar--sse)
14. [Toast Notifications](#14-toast-notifications)
15. [Error Handling](#15-error-handling)
16. [Events](#16-events)
17. [Configuration](#17-configuration)
18. [CLI Generator](#18-cli-generator)
19. [Testing](#19-testing)

---

## 1. Project Structure

```
gofar/
├── cmd/
│   ├── server/main.go          ← application entry point
│   └── gofar/                  ← CLI code generator
│       ├── main.go
│       └── gen/
├── framework/                  ← framework core (don't edit)
│   ├── app/                    ← App struct, bootstrap
│   ├── config/                 ← environment config
│   ├── container/              ← IoC service container
│   ├── data/                   ← Repository interface, pagination
│   ├── datastar/               ← SSE, signals, fragments
│   ├── errors/                 ← typed HTTP errors
│   ├── event/                  ← in-process event bus
│   ├── form/                   ← form schema & validation
│   ├── gofar/                  ← module facade (fw.Use)
│   ├── logger/                 ← structured logger
│   ├── middleware/              ← global middleware
│   ├── module/                 ← module lifecycle
│   ├── render/                 ← templ rendering helpers
│   ├── request/                ← request parsing helpers
│   ├── response/               ← JSON response helpers
│   ├── scheduler/              ← cron/task scheduler
│   ├── services/               ← Registry + BaseService
│   └── store/                  ← key/value store
├── modules/                    ← your application modules
│   ├── all/all.go              ← module registration
│   └── base/                   ← example module
│       ├── module.go
│       ├── routes.go
│       ├── events.go
│       ├── manifest.json
│       ├── repository/
│       ├── service/
│       └── handler/
├── views/                      ← templ components
│   ├── layouts/
│   └── components/
├── static/                     ← CSS, JS, images
├── ent/                        ← Ent ORM generated code
└── .env
```

---

## 2. Bootstrap & Lifecycle

GoFar bootstraps in a fixed, documented order so every dependency is available when it is needed:

```
Config → Store → Events → Logger → Fiber → Middleware → Database → Modules → Migration
```

The application entry point is minimal:

```go
// cmd/server/main.go
func main() {
    app, err := app.New()
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(app.Run(":3000"))
}
```

`app.New()` calls `Bootstrap()` internally, which wires everything in the correct order. You never call setup functions manually.

### Module Lifecycle

Every module goes through three phases:

```
Register (all modules) → Boot (all modules) → Shutdown (reverse order)
```

| Phase      | Purpose                                                                             |
| ---------- | ----------------------------------------------------------------------------------- |
| `Register` | Bind repos, services, handlers into the container. No cross-module calls yet.       |
| `Boot`     | Mount routes, subscribe to events, start workers. Cross-module calls are safe here. |
| `Shutdown` | Release resources, flush buffers, close connections.                                |

---

## 3. Modules

A module is a self-contained bounded context. Every feature in your application is a module.

### Creating a Module

```bash
go run ./cmd/gofar make:module name=Billing
```

This generates the full scaffold:

```
modules/billing/
├── manifest.json       ← module metadata
├── module.go           ← lifecycle hooks
├── routes.go           ← HTTP route registration
├── events.go           ← event subscriptions
├── repository/
│   └── repository.go   ← Repositories struct
├── service/
│   └── service.go      ← Services struct
└── handler/
    └── handler.go      ← Handlers struct
```

### Module Structure

```go
// modules/billing/module.go
package billing

import (
    "context"

    fw "github.com/subhasundardas/gofar/framework/gofar"
    "github.com/subhasundardas/gofar/framework/module"
    "github.com/subhasundardas/gofar/modules/billing/handler"
    "github.com/subhasundardas/gofar/modules/billing/repository"
    "github.com/subhasundardas/gofar/modules/billing/service"
)

type Module struct {
    *module.BaseModule
    repos    *repository.Repositories
    services *service.Services
    handlers *handler.Handlers
}

var _ module.Module = (*Module)(nil)

func New() module.Module {
    return &Module{
        BaseModule: module.MustNewBase("./modules/billing/manifest.json"),
    }
}

func init() {
    module.Register(New)
}

func (m *Module) Register(mgr *module.Manager) error {
    svc := fw.Use(mgr).Services

    m.repos    = repository.NewRepositories(svc.DB())
    m.services = service.NewServices(m.repos, svc.BaseService)
    m.handlers = handler.NewHandlers(m.services)
    return nil
}

func (m *Module) Boot(mgr *module.Manager) error {
    svc := fw.Use(mgr).Services

    RegisterRoutes(mgr.Context.Fiber, m.handlers)
    RegisterEvents(svc.Events(), m.services, svc.Logger())
    return nil
}

func (m *Module) Shutdown(ctx context.Context) error {
    return m.BaseModule.Shutdown(ctx)
}
```

### manifest.json

```json
{
  "name": "billing",
  "version": "1.0.0",
  "description": "Manages invoices and payments",
  "author": "Your Name",
  "enabled": true
}
```

Set `"enabled": false` to disable a module at bootstrap without removing code.

### Registering a Module

Modules self-register via `init()`. Import the package in `modules/all/all.go`:

```go
// modules/all/all.go
package all

import (
    _ "github.com/subhasundardas/gofar/modules/billing"
    _ "github.com/subhasundardas/gofar/modules/auth"
)
```

---

## 4. The gofar Facade

Every module imports `framework/gofar` as `fw` — the single entry point to the entire framework:

```go
import fw "github.com/subhasundardas/gofar/framework/gofar"

func (m *Module) Register(mgr *module.Manager) error {
    gofar := fw.Use(mgr)
    svc   := gofar.Services
    // ...
}
```

`fw.Use(mgr)` returns a `*Framework` with one field: `Services` — your access point to every framework dependency.

---

## 5. Services Registry (svc)

`svc` is a `*services.Registry` — the single object that gives you everything the framework provides. Get it via `fw.Use(mgr).Services`.

```go
svc := fw.Use(mgr).Services

// ── Infrastructure dependencies ────────────────────────
svc.DB()       // *ent.Client        — Ent ORM database client
svc.Logger()   // *logger.Logger     — structured logger
svc.Events()   // *event.Bus         — in-process event bus
svc.Config()   // *config.Config     — environment configuration
svc.Store()    // *store.Store       — application key/value store
svc.Fiber()    // *fiber.App         — root Fiber app (use in Boot)

// ── BaseService helpers (embedded, available directly on svc) ──
svc.WrapError("op", err)
svc.ValidateRequired(map[string]string{"name": name})
svc.ValidatePositive("amount", 10.0)
svc.ValidateMinLength("password", val, 8)
svc.ValidateMaxLength("bio", val, 500)
svc.Paginate(ctx, params, repo.List)

// ── Inject into your module services ──────────────────
svc.BaseService   // *services.BaseService — pass to NewServices()
```

### When to use what

| Use `svc.*`                         | Import directly        |
| ----------------------------------- | ---------------------- |
| `svc.DB()` — needs container        | `form` — stateless     |
| `svc.Logger()` — needs container    | `errors` — stateless   |
| `svc.Events()` — needs container    | `render` — stateless   |
| `svc.Config()` — needs container    | `request` — stateless  |
| `svc.BaseService` — needs container | `datastar` — stateless |
|                                     | `data` — stateless     |

**Rule:** if it needs the container to resolve a live dependency, it's on `svc`. If it's pure functions with no state, import it directly.

---

## 6. Repositories & Data Layer

Repositories are the only layer that touches the database. Each module owns its own repositories.

### Base Repository Interface

```go
// framework/data
type Repository[T Entity] interface {
    List(ctx context.Context) ([]T, error)
    FindByID(ctx context.Context, id int) (T, error)
    Delete(ctx context.Context, id int) error
}
```

### Creating a Repository

```bash
go run ./cmd/gofar make:repository name=Invoice module=Billing
```

```go
// modules/billing/repository/invoice.go
package repository

import (
    "context"

    "github.com/subhasundardas/gofar/ent"
    "github.com/subhasundardas/gofar/framework/data"
)

type InvoiceRepository struct {
    db *ent.Client
}

// compile-time interface check
var _ data.Repository[*ent.Invoice] = (*InvoiceRepository)(nil)

func NewInvoiceRepository(db *ent.Client) *InvoiceRepository {
    return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) List(ctx context.Context) ([]*ent.Invoice, error) {
    return r.db.Invoice.Query().All(ctx)
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id int) (*ent.Invoice, error) {
    return r.db.Invoice.Get(ctx, id)
}

func (r *InvoiceRepository) Delete(ctx context.Context, id int) error {
    return r.db.Invoice.DeleteOneID(id).Exec(ctx)
}

// Module-specific queries beyond the base interface
func (r *InvoiceRepository) ListPaginated(ctx context.Context, offset, limit int) ([]*ent.Invoice, int, error) {
    total, _ := r.db.Invoice.Query().Count(ctx)
    items, err := r.db.Invoice.Query().Offset(offset).Limit(limit).All(ctx)
    return items, total, err
}
```

### Repositories Bundle

```go
// modules/billing/repository/repository.go
package repository

import "github.com/subhasundardas/gofar/ent"

type Repositories struct {
    Invoice *InvoiceRepository
    Payment *PaymentRepository
}

func NewRepositories(db *ent.Client) *Repositories {
    return &Repositories{
        Invoice: NewInvoiceRepository(db),
        Payment: NewPaymentRepository(db),
    }
}
```

### Pagination

```go
import "github.com/subhasundardas/gofar/framework/data"

// in your service
func (s *InvoiceService) List(ctx context.Context, params data.PaginationParams) (*data.PaginatedResult[*ent.Invoice], error) {
    return s.Paginate(ctx, params, s.repo.ListPaginated)
}

// in your handler
params := data.PaginationParams{
    Page:    request.QueryInt(c, "page", 1),
    PerPage: request.QueryInt(c, "per_page", 20),
}
result, err := h.svc.List(c.Context(), params)
return response.Paginated(c, result.Items, response.NewPager(result.Page, result.PerPage, result.Total))
```

---

## 7. Services Layer

Services hold your business logic. They sit between handlers and repositories.

### BaseService

Every module service embeds `*services.BaseService` — injected via `svc.BaseService`. This gives you logging, validation, error wrapping, and pagination for free.

```go
// modules/billing/service/invoice_service.go
package service

import (
    "context"

    "github.com/subhasundardas/gofar/framework/data"
    "github.com/subhasundardas/gofar/framework/services"
    "github.com/subhasundardas/gofar/modules/billing/repository"
)

type InvoiceService struct {
    *services.BaseService          // ← Logger, WrapError, Validate*, Paginate
    repo *repository.InvoiceRepository
}

func NewInvoiceService(repo *repository.InvoiceRepository, base *services.BaseService) *InvoiceService {
    return &InvoiceService{BaseService: base, repo: repo}
}

func (s *InvoiceService) Create(ctx context.Context, amount float64, clientName string) (*ent.Invoice, error) {
    if err := s.ValidateRequired(map[string]string{"client_name": clientName}); err != nil {
        return nil, err
    }
    if err := s.ValidatePositive("amount", amount); err != nil {
        return nil, err
    }

    invoice, err := s.repo.Create(ctx, amount, clientName)
    return invoice, s.WrapError("InvoiceService.Create", err)
}

func (s *InvoiceService) List(ctx context.Context, params data.PaginationParams) (*data.PaginatedResult[*ent.Invoice], error) {
    return s.Paginate(ctx, params, s.repo.ListPaginated)
}
```

### Services Bundle

```go
// modules/billing/service/service.go
package service

import (
    "github.com/subhasundardas/gofar/framework/services"
    "github.com/subhasundardas/gofar/modules/billing/repository"
)

type Services struct {
    Invoice *InvoiceService
    Payment *PaymentService
}

func NewServices(repos *repository.Repositories, base *services.BaseService) *Services {
    return &Services{
        Invoice: NewInvoiceService(repos.Invoice, base),
        Payment: NewPaymentService(repos.Payment, base),
    }
}
```

---

## 8. Handlers

Handlers handle HTTP concerns only — parsing requests, calling services, sending responses.

```bash
go run ./cmd/gofar make:handler name=Invoice module=Billing
```

```go
// modules/billing/handler/invoice.go
package handler

import (
    "github.com/gofiber/fiber/v2"
    "github.com/subhasundardas/gofar/framework/data"
    "github.com/subhasundardas/gofar/framework/errors"
    "github.com/subhasundardas/gofar/framework/request"
    "github.com/subhasundardas/gofar/framework/response"
    "github.com/subhasundardas/gofar/modules/billing/service"
)

type InvoiceHandlers struct {
    svc *service.InvoiceService
}

func NewInvoiceHandlers(svc *service.InvoiceService) *InvoiceHandlers {
    return &InvoiceHandlers{svc: svc}
}

func (h *InvoiceHandlers) List(c *fiber.Ctx) error {
    params := data.PaginationParams{
        Page:    request.QueryInt(c, "page", 1),
        PerPage: request.QueryInt(c, "per_page", 20),
    }
    result, err := h.svc.List(c.Context(), params)
    if err != nil {
        return response.Error(c, err)
    }
    return response.Paginated(c, result.Items, response.NewPager(result.Page, result.PerPage, result.Total))
}

func (h *InvoiceHandlers) Create(c *fiber.Ctx) error {
    type body struct {
        Amount     float64 `json:"amount"`
        ClientName string  `json:"client_name"`
    }
    dto, err := request.Parse[body](c)
    if err != nil {
        return response.Error(c, err)
    }
    invoice, err := h.svc.Create(c.Context(), dto.Amount, dto.ClientName)
    if err != nil {
        return response.Error(c, err)
    }
    return response.Created(c, invoice)
}

func (h *InvoiceHandlers) Get(c *fiber.Ctx) error {
    id, err := request.ParamInt(c, "id")
    if err != nil {
        return response.Error(c, err)
    }
    invoice, err := h.svc.FindByID(c.Context(), id)
    if err != nil {
        return response.Error(c, errors.NewNotFound("invoice", id))
    }
    return response.OK(c, invoice)
}

func (h *InvoiceHandlers) Delete(c *fiber.Ctx) error {
    id, err := request.ParamInt(c, "id")
    if err != nil {
        return response.Error(c, err)
    }
    if err := h.svc.Delete(c.Context(), id); err != nil {
        return response.Error(c, err)
    }
    return response.NoContent(c)
}
```

---

## 9. Routing

Routes are registered in `Boot` via a dedicated `routes.go` file per module.

```go
// modules/billing/routes.go
package billing

import (
    "github.com/gofiber/fiber/v2"
    "github.com/subhasundardas/gofar/modules/billing/handler"
)

func RegisterRoutes(app *fiber.App, h *handler.Handlers) {
    billing := app.Group("/billing")

    invoices := billing.Group("/invoices")
    invoices.Get("/",     h.Invoice.List)
    invoices.Post("/",    h.Invoice.Create)
    invoices.Get("/:id",  h.Invoice.Get)
    invoices.Delete("/:id", h.Invoice.Delete)
}
```

---

## 10. Forms

GoFar includes a type-safe form package with field definitions, validators, and per-request instances.

```go
import "github.com/subhasundardas/gofar/framework/form"
import "github.com/subhasundardas/gofar/framework/form/inputs"

// Build schema once — in NewHandlers
invoiceForm := form.New("invoice-create").Fields(
    inputs.Text("client_name", form.Label("Client Name"), form.Required()),
    inputs.Number("amount",    form.Label("Amount"),      form.Required()),
)

// Per request — in a handler
inst := invoiceForm.NewInstance()
inst.Set("client_name", c.FormValue("client_name"))
inst.Set("amount", c.FormValue("amount"))
inst.Validate()

if !inst.Valid() {
    return c.Status(422).JSON(inst.Errors)
}

// Extract clean values and pass to service
name := inst.Get("client_name").(string)
```

Forms live in `Handlers`, not in services — they are a UI concern.

---

## 11. Request & Response

### Request Helpers

```go
import "github.com/subhasundardas/gofar/framework/request"

// Parse JSON body into a struct
dto, err := request.Parse[CreateInvoiceDTO](c)

// Parse query string
query, err := request.ParseQuery[ListQuery](c)

// Route parameters
id, err  := request.ParamInt(c, "id")
slug, err := request.ParamString(c, "slug")

// Query parameters with defaults
page    := request.QueryInt(c, "page", 1)
search  := request.QueryString(c, "search", "")

// Auth token
token := request.BearerToken(c)
```

### Response Helpers

```go
import "github.com/subhasundardas/gofar/framework/response"

response.OK(c, data)                          // 200 { success: true, data: ... }
response.Created(c, data)                     // 201 { success: true, data: ... }
response.NoContent(c)                         // 204
response.Paginated(c, items, pager)           // 200 with pagination metadata
response.Error(c, err)                        // auto-detects status from AppError
response.BadRequest(c, "invalid input")       // 400
response.Unauthorized(c, "login required")    // 401
response.Forbidden(c, "access denied")        // 403
response.NotFound(c, "invoice")               // 404
response.ValidationFailed(c, fields)          // 422 with field errors
```

---

## 12. Rendering with Templ

GoFar uses [a-h/templ](https://templ.guide) for type-safe server-side HTML rendering.

```go
import "github.com/subhasundardas/gofar/framework/render"

// Render a full page
func (h *InvoiceHandlers) Index(c *fiber.Ctx) error {
    return render.Component(c, layouts.Base("Invoices"){
        views.InvoiceList(invoices)
    })
}

// Render component to string (for email, SSE fragments)
html, err := render.ComponentToString(views.InvoiceCard(invoice))

// Raw HTML (edge cases only)
render.HTML(c, "<p>Hello</p>")

// Plain text
render.Text(c, "pong")
```

### Base Layout

```templ
// views/layouts/base.templ
templ Base(title string) {
    <!DOCTYPE html>
    <html>
    <head><title>{ title }</title></head>
    <body data-signals='{"toast":{"message":"","level":"success","visible":false}}'>
        { children... }
        @components.Toast()
    </body>
    </html>
}
```

Usage:

```templ
@layouts.Base("Invoices") {
    @views.InvoiceList(invoices)
}
```

---

## 13. Datastar & SSE

GoFar integrates with [Datastar](https://data-star.dev) for reactive UI over SSE — no JavaScript framework needed.

```go
import "github.com/subhasundardas/gofar/framework/datastar"

// Push signals (reactive state) to client
datastar.MarshalAndMergeSignals(c, map[string]any{
    "count": 42,
    "loading": false,
})

// Push a templ component fragment
datastar.MergeFragmentTempl(c, views.InvoiceRow(invoice))

// Push a raw HTML fragment
datastar.MergeFragments(c, `<div id="status">Saved</div>`)

// Open an SSE stream and send multiple events
datastar.SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
    sse.PatchElementTempl(views.InvoiceRow(invoice))
    sse.MarshalAndPatchSignals(map[string]any{"loading": false})
    return nil
})

// Read signals sent from the client
var signals struct {
    Search string `json:"search"`
    Page   int    `json:"page"`
}
datastar.ReadSignals(c, &signals)
```

### Form State with Datastar

```go
// Read incoming form state
patch, err := datastar.ReadFormState(c, "invoice-create")

// Apply to form instance and validate
inst := form.NewInstance(invoiceForm).Apply(patch)
inst.Validate()

// Push updated state back to client (shows errors, disables submit, etc.)
return datastar.PushFormState(c, form.NewFormState("invoice-create", inst))
```

---

## 14. Toast Notifications

Toast notifications are driven by Datastar signals — the backend pushes a signal, the frontend reacts.

### Backend

```go
import "github.com/subhasundardas/gofar/framework/datastar"

// Success toast
datastar.Toast(c, datastar.ToastSuccess, "Invoice saved successfully")

// Error toast
datastar.Toast(c, datastar.ToastError, "Failed to save invoice")

// Warning and info
datastar.Toast(c, datastar.ToastWarning, "Invoice is overdue")
datastar.Toast(c, datastar.ToastInfo, "Syncing with payment gateway...")
```

### Frontend Component

```templ
// views/components/toast.templ
templ Toast() {
    <div
        id="toast"
        data-show="$toast.visible"
        data-on-load="setTimeout(() => $toast.visible = false, 3000)"
        class="fixed bottom-4 right-4 px-4 py-2 rounded text-white shadow-lg"
    >
        <span data-text="$toast.message"></span>
        <button data-on-click="$toast.visible = false">✕</button>
    </div>
}
```

Include `@components.Toast()` once in your base layout — it is always in the DOM, hidden until a signal fires.

---

## 15. Error Handling

GoFar uses typed HTTP-aware errors that carry a status code, machine-readable code, and human message.

```go
import "github.com/subhasundardas/gofar/framework/errors"

// Constructors
errors.NewNotFound("invoice", id)               // 404
errors.NewValidation(map[string]string{...})    // 422 with field errors
errors.NewUnauthorized("token expired")         // 401
errors.NewForbidden("insufficient permissions") // 403
errors.NewBadRequest("invalid UUID")            // 400
errors.NewConflict("invoice")                   // 409
errors.NewInternal(err)                         // 500 (hides cause from client)

// Wrapping
errors.Wrap(422, "customCode", "message", originalErr)

// Unwrapping (standard library compatible)
var appErr *errors.AppError
if errors.As(err, &appErr) {
    fmt.Println(appErr.Status, appErr.Code, appErr.Message)
}
```

`response.Error(c, err)` automatically detects `*AppError` and uses the correct HTTP status — no manual status codes in handlers.

---

## 16. Events

The in-process event bus lets modules communicate without direct dependencies.

```go
import "github.com/subhasundardas/gofar/framework/event"

// Subscribe (in Boot)
svc.Events().Subscribe("invoice.created", func(payload any) {
    invoice := payload.(*ent.Invoice)
    // send email, update stats, etc.
})

// Publish (from a service or handler)
svc.Events().Publish("invoice.created", invoice)
```

Modules subscribe in `Boot` — all modules are registered by that point so there are no missing subscribers.

---

## 17. Configuration

GoFar reads configuration from environment variables and an optional `.env` file.

```go
// Get a value (empty string if missing)
svc.Config().Get("STRIPE_KEY")

// Get with a default
svc.Config().GetWithDefault("APP_PORT", "3000")

// Set programmatically
svc.Config().Set("FEATURE_X", "true")
```

`.env` file:

```
APP_NAME=MyApp
APP_ENV=production
DB_DRIVER=postgres
DB_DSN=postgres://user:pass@localhost/mydb
STRIPE_KEY=sk_live_...
```

Framework-known defaults (always set unless overridden by environment):

| Key         | Default            |
| ----------- | ------------------ |
| `APP_NAME`  | `GoFar`            |
| `APP_ENV`   | `development`      |
| `DB_DRIVER` | `sqlite`           |
| `DB_DSN`    | `./gofar.db?_fk=1` |

---

## 18. CLI Generator

GoFar ships a CLI tool to generate module scaffolds and add components to existing modules.

```bash
# Generate a complete module
go run ./cmd/gofar make:module name=Billing

# Add a handler to an existing module
go run ./cmd/gofar make:handler name=Invoice module=Billing

# Add a service to an existing module
go run ./cmd/gofar make:service name=TaxCalculator module=Billing

# Add a repository to an existing module
go run ./cmd/gofar make:repository name=Invoice module=Billing

# List all registered modules
go run ./cmd/gofar list:modules
```

Add a Makefile shortcut:

```makefile
module:
    go run ./cmd/gofar make:module name=$(name)

handler:
    go run ./cmd/gofar make:handler name=$(name) module=$(module)
```

```bash
make module name=Billing
make handler name=Invoice module=Billing
```

---

## 19. Testing

GoFar is designed to be testable at every layer without spinning up the full framework.

### Testing a Service

```go
func TestInvoiceService_Create(t *testing.T) {
    // mock logger — satisfies services.Logger interface
    type mockLogger struct{}
    func (m *mockLogger) Info(msg string, args ...any)  {}
    func (m *mockLogger) Warn(msg string, args ...any)  {}
    func (m *mockLogger) Error(msg string, args ...any) {}

    base := &services.BaseService{Logger: &mockLogger{}}
    repo := &mockInvoiceRepository{}  // implement data.Repository[*ent.Invoice]
    svc  := NewInvoiceService(repo, base)

    invoice, err := svc.Create(context.Background(), 100.0, "Acme Corp")
    assert.NoError(t, err)
    assert.Equal(t, "Acme Corp", invoice.ClientName)
}
```

### Testing a Handler

```go
func TestInvoiceHandler_Get(t *testing.T) {
    app := fiber.New()
    svc := &mockInvoiceService{}
    h   := NewInvoiceHandlers(svc)

    app.Get("/invoices/:id", h.Get)

    req := httptest.NewRequest("GET", "/invoices/1", nil)
    resp, _ := app.Test(req)
    assert.Equal(t, 200, resp.StatusCode)
}
```

No framework bootstrap needed — each layer is independently testable.

---

## Summary

GoFar gives you a clear, consistent structure for every feature in your application:

```
fw.Use(mgr).Services    ← one accessor, everything from the framework
      ↓
repository.NewRepositories(svc.DB())    ← your data layer
      ↓
service.NewServices(repos, svc.BaseService)    ← your business logic
      ↓
handler.NewHandlers(services)    ← your HTTP layer
      ↓
RegisterRoutes(fiber, handlers)    ← your routes
```

Every module follows the same pattern. Every developer on your team knows where to find everything. The CLI means you never write boilerplate. And every layer is independently testable.

**GoFar — structured Go, the far way.**

---

_Built with Go, Fiber, Ent ORM, Templ, and Datastar._
