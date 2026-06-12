// Package pagination is the cross-cutting wrapper around the framework's
// pagination primitives.
//
// It exposes a single one-liner that every list handler can call so
// pagination is applied uniformly across the application:
//
//	func ListInvoices(c *fiber.Ctx) error {
//	    res, err := svc.List(c.Context(), pagination.FromQuery(c))
//	    if err != nil { return response.Error(c, err) }
//	    return pagination.Send(c, res)
//	}
//
// The package does not own the request/response shape — it delegates to
// framework/repository (data) and framework/response (HTTP) so behaviour
// stays in one place.
package pagination

import (
	"github.com/gofiber/fiber/v2"

	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/response"
)

// MaxPerPage is the hard upper bound on items per page. Requests asking
// for more are silently capped. Override in tests if needed.
const MaxPerPage = 200

// FromQuery parses page/per_page/search/order_by/desc from the request
// query string and returns a normalised repository.PaginationParams.
// Invalid or missing values fall back to the framework defaults, with
// per_page clamped to MaxPerPage.
func FromQuery(c *fiber.Ctx) data.PaginationParams {
	p := data.DefaultPagination()

	if v := c.QueryInt("page", 0); v > 0 {
		p.Page = v
	}
	if v := c.QueryInt("per_page", 0); v > 0 {
		p.PerPage = v
	}
	if p.PerPage > MaxPerPage {
		p.PerPage = MaxPerPage
	}
	if s := c.Query("search"); s != "" {
		p.Search = s
	}
	if o := c.Query("order_by"); o != "" {
		p.OrderBy = o
	}
	p.Desc = c.Query("desc") == "true"
	return p
}

// Send emits a paginated response envelope. It is the only thing a
// handler needs to return — the items land under "data" and the
// pagination metadata under "pagination".
//
//	return pagination.Send(c, res)
func Send(c *fiber.Ctx, res *data.PaginatedResult[any]) error {
	return response.Paginated(c, res.Items, response.NewPager(
		res.Page, res.PerPage, res.Total,
	))
}
