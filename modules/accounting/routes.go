// routes.go declares all HTTP routes for the Accounting module.
// Called from Module.Boot — do not call directly.
package accounting

import (
	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/modules/accounting/handler"
)

// RegisterRoutes mounts Accounting endpoints under /accounting.
//
// Route table:
//
//	GET    /accounting          → list
//	POST   /accounting          → create
//	GET    /accounting/:id      → get by ID
//	PUT    /accounting/:id      → update
//	DELETE /accounting/:id      → delete
func RegisterRoutes(app *fiber.App, h *handler.Handlers) {
	grp := app.Group("/accounting")
	grp.Get("/", h.Accounting.ChartOfAccount)
	grp.Get("/ledger", h.Accounting.ListLedger)

	// Journal
	grp.Get("/journal", h.Accounting.ListJournal)
	grp.Get("/journal/new", h.Accounting.NewJournal)
	grp.Get("/journal/new/newrow", h.Accounting.AddRow)

}
