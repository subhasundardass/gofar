package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/form"
	"github.com/subhasundardas/gofar/framework/render"
	"github.com/subhasundardas/gofar/framework/response"
	"github.com/subhasundardas/gofar/modules/accounting/service"
	"github.com/subhasundardas/gofar/modules/accounting/views"
)

type MasterHandler struct {
	acctSvc *service.AccountingService
}

func NewMasterHandler(
	acc *service.AccountingService) *MasterHandler {
	return &MasterHandler{
		acctSvc: acc,
	}
}

// --Ledger -
var LedgerForm = form.New("ledger").Fields(

	// ─────────────────────────────────────────────
	// Identity
	// ─────────────────────────────────────────────

	form.TextInput("code",
		form.Label("Ledger Code"),
		form.Placeholder("LED-001"),
		form.Required(),
	),

	form.TextInput("name",
		form.Label("Ledger Name"),
		form.Placeholder("Cash Account"),
		form.Required(),
	),

	// ─────────────────────────────────────────────
	// Relations
	// ─────────────────────────────────────────────

	form.NumberInput("group_id",
		form.Label("Group"),
		form.Required(),
	),

	// ─────────────────────────────────────────────
	// Optional fields
	// ─────────────────────────────────────────────

	form.TextAreaInput("description",
		form.Label("Description"),
		form.Placeholder("Optional notes about this ledger"),
	),

	// ─────────────────────────────────────────────
	// System / computed fields
	// ─────────────────────────────────────────────

	form.NumberInput("balance",
		form.Label("Opening Balance"),
		form.Default(0.00),
		form.ReadOnly(), // usually controlled by system
	),
)

func (h *MasterHandler) ListLedger(c *fiber.Ctx) error {
	params := data.PaginationParams{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 20),
		Search:  c.Query("q"),
	}
	result, err := h.acctSvc.ListLedgers(c.Context(), params)
	if err != nil {
		return response.Error(c, err)
	}
	return render.Component(c, views.LedgerView("Ledger Master", result))
}

// Ledger New
func (h *MasterHandler) NewLedger(c *fiber.Ctx) error {

	inst := form.NewInstance(LedgerForm)
	state := form.NewFormState("ledger", inst)
	fields := form.BuildUIFields(LedgerForm, inst, state)
	fieldMap := form.BuildUIFieldMap(fields)

	props := views.LedgerViewProps{
		Title:     "New Ledger",
		Form:      LedgerForm,
		Fields:    fields,
		FormState: state,
		FieldMap:  fieldMap,
	}

	return render.Component(c, views.NewLedgerView(props))
}
