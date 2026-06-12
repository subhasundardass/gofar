package gen

import (
	"fmt"
	"os"
	"strings"
)

// MakeHandler generates a new handler file inside an existing module.
//
//	go run ./cmd/gofar make:handler name=Refund module=Billing
//
// Creates: modules/<module_snake>/handler/<handler_snake>.go
func MakeHandler(args map[string]string) error {
	handlerName, ok := args["name"]
	if !ok || handlerName == "" {
		return fmt.Errorf("name is required: make:handler name=<Name> module=<Module>")
	}
	moduleName, ok := args["module"]
	if !ok || moduleName == "" {
		return fmt.Errorf("module is required: make:handler name=<Name> module=<Module>")
	}

	h := newNames(handlerName)
	mod := newNames(moduleName)
	modPath, err := readModulePath()
	if err != nil {
		return err
	}

	fmt.Printf("\n🔧 Adding handler \"%s\" to module \"%s\"\n\n", h.Pascal, mod.Pascal)

	type handlerTemplateData struct {
		Handler Names
		Module  Names
		ModPath string
	}

	tmpl := `// {{.Handler.Pascal}}Handlers holds HTTP handlers for the {{.Handler.Pascal}} resource.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"{{.ModPath}}/framework/request"
	"{{.ModPath}}/framework/response"
	"{{.ModPath}}/modules/{{.Module.Snake}}/service"
)

// {{.Handler.Pascal}}Handlers handles {{.Handler.Pascal}} HTTP endpoints.
type {{.Handler.Pascal}}Handlers struct {
	svc *service.{{.Module.Pascal}}Service
}

// List handles GET /{{.Module.Kebab}}/{{.Handler.Kebab}}
func (h *{{.Handler.Pascal}}Handlers) List(c *fiber.Ctx) error {
	return response.OK(c, []any{})
}

// Create handles POST /{{.Module.Kebab}}/{{.Handler.Kebab}}
func (h *{{.Handler.Pascal}}Handlers) Create(c *fiber.Ctx) error {
	type body struct {
		Name string ` + "`json:\"name\"`" + `
	}
	dto, err := request.Parse[body](c)
	if err != nil {
		return response.Error(c, err)
	}
	_ = dto
	return response.Created(c, fiber.Map{"status": "created"})
}

// Get handles GET /{{.Module.Kebab}}/{{.Handler.Kebab}}/:id
func (h *{{.Handler.Pascal}}Handlers) Get(c *fiber.Ctx) error {
	id, err := request.ParamInt(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, fiber.Map{"id": id})
}

// Update handles PUT /{{.Module.Kebab}}/{{.Handler.Kebab}}/:id
func (h *{{.Handler.Pascal}}Handlers) Update(c *fiber.Ctx) error {
	id, err := request.ParamInt(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, fiber.Map{"id": id, "updated": true})
}

// Delete handles DELETE /{{.Module.Kebab}}/{{.Handler.Kebab}}/:id
func (h *{{.Handler.Pascal}}Handlers) Delete(c *fiber.Ctx) error {
	if _, err := request.ParamInt(c, "id"); err != nil {
		return response.Error(c, err)
	}
	return response.NoContent(c)
}
`
	content, err := renderTemplate("handler", tmpl, handlerTemplateData{
		Handler: h, Module: mod, ModPath: modPath,
	})
	if err != nil {
		return err
	}

	path := fmt.Sprintf("modules/%s/handler/%s.go", mod.Snake, h.Snake)
	if err := writeFile(path, content); err != nil {
		return err
	}

	fmt.Printf("\n✅ Handler \"%s\" created.\n", h.Pascal)
	fmt.Printf("   Add it to Handlers struct in modules/%s/handler/handler.go\n", mod.Snake)
	fmt.Printf("   Mount routes in modules/%s/routes.go\n\n", mod.Snake)
	return nil
}

// MakeService generates a new service file inside an existing module.
//
//	go run ./cmd/gofar make:service name=TaxCalculator module=Billing
//
// Creates: modules/<module_snake>/service/<service_snake>.go
func MakeService(args map[string]string) error {
	serviceName, ok := args["name"]
	if !ok || serviceName == "" {
		return fmt.Errorf("name is required: make:service name=<Name> module=<Module>")
	}
	moduleName, ok := args["module"]
	if !ok || moduleName == "" {
		return fmt.Errorf("module is required: make:service name=<Name> module=<Module>")
	}

	svc := newNames(serviceName)
	mod := newNames(moduleName)
	modPath, err := readModulePath()
	if err != nil {
		return err
	}

	fmt.Printf("\n🔧 Adding service \"%s\" to module \"%s\"\n\n", svc.Pascal, mod.Pascal)

	type svcTemplateData struct {
		Service Names
		Module  Names
		ModPath string
	}

	tmpl := `// {{.Service.Pascal}}Service implements {{.Service.Pascal}} use-cases for the {{.Module.Pascal}} module.
package service

import (
	"{{.ModPath}}/modules/{{.Module.Snake}}/repository"
)

// {{.Service.Pascal}}Service holds the business logic for {{.Service.Pascal}}.
type {{.Service.Pascal}}Service struct {
	repos *repository.Repositories
}

// New{{.Service.Pascal}}Service constructs a {{.Service.Pascal}}Service.
func New{{.Service.Pascal}}Service(repos *repository.Repositories) *{{.Service.Pascal}}Service {
	return &{{.Service.Pascal}}Service{repos: repos}
}

// TODO: add use-case methods here.
`
	content, err := renderTemplate("service", tmpl, svcTemplateData{
		Service: svc, Module: mod, ModPath: modPath,
	})
	if err != nil {
		return err
	}

	path := fmt.Sprintf("modules/%s/service/%s.go", mod.Snake, svc.Snake)
	if err := writeFile(path, content); err != nil {
		return err
	}

	fmt.Printf("\n✅ Service \"%s\" created.\n", svc.Pascal)
	fmt.Printf("   Add it to Services struct in modules/%s/service/service.go\n\n", mod.Snake)
	return nil
}

// ListModules prints every module currently registered in modules/all/all.go.
func ListModules() error {
	data, err := os.ReadFile("modules/all/all.go")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No modules registered yet (modules/all/all.go not found).")
			return nil
		}
		return err
	}

	fmt.Println("Registered modules (modules/all/all.go):")
	count := 0
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, `_ "`) {
			trimmed := strings.Trim(line, `_ "`)
			fmt.Printf("  • %s\n", trimmed[:len(trimmed)-1])
			count++
		}
	}
	if count == 0 {
		fmt.Println("  (none yet)")
	}
	fmt.Println()
	return nil
}

// MakeRepository generates a new repository file inside an existing module.
//
//	go run ./cmd/gofar make:repository name=Invoice module=Billing
//
// Creates: modules/<module_snake>/repository/<repository_snake>.go
func MakeRepository(args map[string]string) error {
	repoName, ok := args["name"]
	if !ok || repoName == "" {
		return fmt.Errorf("name is required: make:repository name=<Name> module=<Module>")
	}
	moduleName, ok := args["module"]
	if !ok || moduleName == "" {
		return fmt.Errorf("module is required: make:repository name=<Name> module=<Module>")
	}

	repo := newNames(repoName)
	mod := newNames(moduleName)
	modPath, err := readModulePath()
	if err != nil {
		return err
	}

	fmt.Printf("\n🔧 Adding repository \"%s\" to module \"%s\"\n\n", repo.Pascal, mod.Pascal)

	type repoTemplateData struct {
		Repo    Names
		Module  Names
		ModPath string
	}

	tmpl := `// {{.Repo.Pascal}}Repository handles database access for the {{.Repo.Pascal}} entity.
package repository

import (
	"context"

	"{{.ModPath}}/ent"
	"{{.ModPath}}/framework/data"
)

// {{.Repo.Pascal}}Repository is the data access layer for {{.Repo.Pascal}}.
type {{.Repo.Pascal}}Repository struct {
	db *ent.Client
}

// compile-time check — fails at build if interface is not satisfied.
var _ data.Repository[*ent.{{.Repo.Pascal}}] = (*{{.Repo.Pascal}}Repository)(nil)

// New{{.Repo.Pascal}}Repository constructs a {{.Repo.Pascal}}Repository.
func New{{.Repo.Pascal}}Repository(db *ent.Client) *{{.Repo.Pascal}}Repository {
	return &{{.Repo.Pascal}}Repository{db: db}
}

// List returns all {{.Repo.Pascal}} records.
func (r *{{.Repo.Pascal}}Repository) List(ctx context.Context) ([]*ent.{{.Repo.Pascal}}, error) {
	return r.db.{{.Repo.Pascal}}.Query().All(ctx)
}

// FindByID returns a single {{.Repo.Pascal}} by ID.
func (r *{{.Repo.Pascal}}Repository) FindByID(ctx context.Context, id int) (*ent.{{.Repo.Pascal}}, error) {
	return r.db.{{.Repo.Pascal}}.Get(ctx, id)
}

// Delete removes a {{.Repo.Pascal}} by ID.
func (r *{{.Repo.Pascal}}Repository) Delete(ctx context.Context, id int) error {
	return r.db.{{.Repo.Pascal}}.DeleteOneID(id).Exec(ctx)
}

// TODO: add query methods here.
`
	content, err := renderTemplate("repository", tmpl, repoTemplateData{
		Repo: repo, Module: mod, ModPath: modPath,
	})
	if err != nil {
		return err
	}

	path := fmt.Sprintf("modules/%s/repository/%s.go", mod.Snake, repo.Snake)
	if err := writeFile(path, content); err != nil {
		return err
	}

	fmt.Printf("\n✅ Repository \"%s\" created.\n", repo.Pascal)
	fmt.Printf("   Add it to Repositories struct in modules/%s/repository/repository.go\n\n", mod.Snake)
	return nil
}
