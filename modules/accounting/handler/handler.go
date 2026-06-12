// Package handler contains HTTP handlers for the Accounting module.
// Handlers are thin: parse request → call service → return response.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/render"
	"github.com/subhasundardas/gofar/framework/request"
	"github.com/subhasundardas/gofar/framework/response"
	"github.com/subhasundardas/gofar/modules/accounting/service"
	"github.com/subhasundardas/gofar/modules/accounting/views"
)

// Handlers bundles all Accounting handler sets.
type Handlers struct {
	Accounting *AccountingHandlers
}

// NewHandlers constructs all handler sets. Takes services only — never fiber.App.
func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		Accounting: &AccountingHandlers{svc: svc.Accounting},
	}
}

// AccountingHandlers holds all Accounting HTTP handlers.
type AccountingHandlers struct {
	svc   *service.AccountingService
	vcSvc *service.VoucherServices
}

// -- Chart of Account
func (h *AccountingHandlers) ChartOfAccount(c *fiber.Ctx) error {
	params := data.PaginationParams{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 20),
		Search:  c.Query("q"),
	}
	result, err := h.svc.ListGroups(c.Context(), params)
	if err != nil {
		return response.Error(c, err)
	}
	return render.Component(c, views.AccountingView("Group Master", result))
}

// -- Ledgers
func (h *AccountingHandlers) ListLedger(c *fiber.Ctx) error {
	params := data.PaginationParams{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 20),
		Search:  c.Query("q"),
	}
	result, err := h.svc.ListLedgers(c.Context(), params)
	if err != nil {
		return response.Error(c, err)
	}
	return render.Component(c, views.LedgerView("Ledger Master", result))
}

// -- Voucher
func (h *AccountingHandlers) VoucherList(c *fiber.Ctx) error {
	params := data.PaginationParams{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 20),
		Search:  c.Query("q"),
	}
	result, err := h.vcSvc.ListVouchers(c.Context(), params)
	if err != nil {
		return response.Error(c, err)
	}

	return render.Component(c, views.VoucherView("Voucher Master", result))
}

// List handles GET /accounting
//
//	200 OK → { success: true, data: [...] }
func (h *AccountingHandlers) List(c *fiber.Ctx) error {
	items, err := h.svc.List()
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, items)
}

// Create handles POST /accounting
//
//	Body:         { "name": "..." }
//	201 Created → { success: true, data: { id: N } }
func (h *AccountingHandlers) Create(c *fiber.Ctx) error {
	type body struct {
		Name string `json:"name"`
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

// Get handles GET /accounting/:id
//
//	200 OK → { success: true, data: {...} }
func (h *AccountingHandlers) Get(c *fiber.Ctx) error {
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

// Update handles PUT /accounting/:id
//
//	Body:   { "name": "..." }
//	200 OK → { success: true, data: { updated: true } }
func (h *AccountingHandlers) Update(c *fiber.Ctx) error {
	id, err := request.ParamInt(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	type body struct {
		Name string `json:"name"`
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

// Delete handles DELETE /accounting/:id
//
//	204 No Content on success
func (h *AccountingHandlers) Delete(c *fiber.Ctx) error {
	id, err := request.ParamInt(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.svc.Delete(id); err != nil {
		return response.Error(c, err)
	}
	return response.NoContent(c)
}
