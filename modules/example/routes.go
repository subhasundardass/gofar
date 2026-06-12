// routes.go declares all HTTP routes for the Example module.
// Called from Module.Boot — do not call directly.
package example

import (
	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/modules/example/handler"
)

// RegisterRoutes mounts Example endpoints under /example.
//
// Route table:
//
//	GET    /example          → list
//	POST   /example          → create
//	GET    /example/:id      → get by ID
//	PUT    /example/:id      → update
//	DELETE /example/:id      → delete
//
// Plus the datastar demo routes:
//
//	GET    /example/         → HTML home page (templ)
//	POST   /example/counter  → SSE fragment that re-renders #counter
func RegisterRoutes(app *fiber.App, h *handler.Handlers) {
	grp := app.Group("/example")

	// Datastar demo
	grp.Get("/", h.Example.Home)
	grp.Post("/counter", h.Example.Increment)
	grp.Post("/counter/reset", h.Example.Reset)

	// Form
	grp.Get("/form", h.Example.HandleForm)

}
