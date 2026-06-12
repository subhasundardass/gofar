// routes.go declares all HTTP routes for the Base module.
// Called from Module.Boot — do not call directly.
package base

import (
	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/modules/base/handler"
)

// RegisterRoutes mounts Base endpoints.
//
// The Base module is cross-cutting, so it owns only the shared lookup
// proxy. The pagination, search and form-widget packages are not HTTP
// routes — they are pure Go helpers that other modules import directly.
//
// Route table:
//
//	GET /base/lookup          → list registered provider names
//	GET /base/lookup/:name    → search a single provider (?q=term&limit=N)
func RegisterRoutes(app *fiber.App, h *handler.Handlers) {
	grp := app.Group("/base/lookup")
	grp.Get("/", h.Lookup.Names)
	grp.Get("/:name", h.Lookup.Search)
}
