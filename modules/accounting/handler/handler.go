// Package handler contains HTTP handlers for the Accounting module.
// Handlers are thin: parse request → call service → return response.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/render"
	"github.com/subhasundardas/gofar/framework/response"
	"github.com/subhasundardas/gofar/modules/accounting/service"
	"github.com/subhasundardas/gofar/modules/accounting/views"
)

// Handlers bundles all Accounting handler sets.
type Handlers struct {
	Accounting *AccountingHandlers
	Journal    *JournalHandlers
}

// AccountingHandlers holds all Accounting HTTP handlers.
type AccountingHandlers struct {
	svc   *service.AccountingService
	jrSvc *service.JournalServices
}

// NewHandlers constructs all handler sets. Takes services only — never fiber.App.
func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		Accounting: NewAccountingHandlers(svc.Accounting),
		Journal:    NewJournalHandlers(svc.Accounting, svc.Journal),
	}
}

func NewAccountingHandlers(svc *service.AccountingService) *AccountingHandlers {
	return &AccountingHandlers{svc: svc}
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
