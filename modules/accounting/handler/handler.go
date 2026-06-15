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
	Accounting *AccountingHandler
	Journal    *JournalHandler
	Master     *MasterHandler
}

// AccountingHandler holds all Accounting HTTP handlers.
type AccountingHandler struct {
	svc *service.AccountingService
}

// NewHandlers constructs all handler sets. Takes services only — never fiber.App.
func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		Accounting: NewAccountingHandler(svc.Accounting),
		Journal:    NewJournalHandler(svc.Accounting, svc.Journal),
		Master:     NewMasterHandler(svc.Accounting),
	}
}

//----------------------------------------------
//--Accounting Handler
//----------------------------------------------

func NewAccountingHandler(svc *service.AccountingService) *AccountingHandler {
	return &AccountingHandler{svc: svc}
}

// -- Chart of Account
func (h *AccountingHandler) ChartOfAccount(c *fiber.Ctx) error {
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
