// Package handler contains the shared HTTP handlers exposed by the Base
// module. The Base module is a cross-cutting module, so the only route
// it serves is the shared lookup proxy used by formui/lookup widgets.
package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	apperrors "github.com/subhasundardas/gofar/framework/errors"
	"github.com/subhasundardas/gofar/framework/response"
	"github.com/subhasundardas/gofar/modules/base/lookup"
	"github.com/subhasundardas/gofar/modules/base/service"
)

// Handlers bundles all Base handler sets.
type Handlers struct {
	Lookup *LookupHandlers
}

// NewHandlers constructs all handler sets. Takes services only — never fiber.App.
func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		Lookup: &LookupHandlers{svc: svc.Lookup},
	}
}

// LookupHandlers exposes the shared /base/lookup/:name endpoint.
type LookupHandlers struct {
	svc *service.LookupService
}

// Search handles GET /base/lookup/:name
//
//	Query params:
//	  q     — search term
//	  limit — max hits (defaults to 20, hard cap 100)
//
//	200 OK → { success: true, data: [ {id,label,extra}, ... ] }
func (h *LookupHandlers) Search(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "lookup name is required")
	}

	term := c.Query("q")
	limit := 20
	if v := c.QueryInt("limit", 0); v > 0 {
		limit = v
	}
	if limit > 100 {
		limit = 100
	}

	hits, err := h.svc.Search(name, lookup.Query{
		Term:  term,
		Limit: limit,
	})
	if err != nil {
		// Unknown provider → 404, anything else → 500 envelope.
		return response.Error(c, apperrors.NewNotFound("lookup:"+name, err.Error()))
	}
	return response.OK(c, hits)
}

// Names handles GET /base/lookup (no path param)
//
//	200 OK → { success: true, data: ["customer", "supplier", ...] }
//
// Useful for diagnostics and admin tooling.
func (h *LookupHandlers) Names(c *fiber.Ctx) error {
	return response.OK(c, h.svc.Names())
}

// ── Helper exposed for tests ─────────────────────────────────────────────────

// ParseLimit is exposed so tests can exercise the limit-parsing rule
// without spinning up a real Fiber context.
func ParseLimit(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
