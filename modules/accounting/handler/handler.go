// Package handler contains HTTP handlers for the Accounting module.
// Handlers are thin: parse request → call service → return response.
package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/datastar"
	"github.com/subhasundardas/gofar/framework/form"
	"github.com/subhasundardas/gofar/framework/render"
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
		Accounting: &AccountingHandlers{
			svc:   svc.Accounting,
			jrSvc: svc.Journal,
		},
	}
}

// AccountingHandlers holds all Accounting HTTP handlers.
type AccountingHandlers struct {
	svc   *service.AccountingService
	jrSvc *service.JournalServices
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

// -- Journal
func (h *AccountingHandlers) ListJournal(c *fiber.Ctx) error {
	params := data.PaginationParams{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 20),
		Search:  c.Query("q"),
	}
	result, err := h.jrSvc.ListJournal(c.Context(), params)
	if err != nil {
		return response.Error(c, err)
	}

	return render.Component(c, views.JournalView("Journal Master", result))
}

func (h *AccountingHandlers) NewJournal(c *fiber.Ctx) error {
	jr := form.New("Journal")
	jr.Fields(
		form.DateField{BaseField: form.BaseField{
			Key:      "jr_date",
			Type:     "date",
			Label:    "Entry Date",
			Required: true,
			DefaultValue: func() time.Time {
				return time.Now()
			},
			Validators: []form.Validator{
				form.RequiredValidator{Field: "jr_date"},
				form.BetweenLengthValidator{Field: "jr_date"},
			},
		}},
		form.TextField{BaseField: form.BaseField{
			Key:          "jr_no",
			Type:         "text",
			Label:        "Journal No",
			Required:     true,
			Placeholder:  "30",
			HelpText:     "Enter Journal no",
			DefaultValue: 25,
			Validators: []form.Validator{
				form.RequiredValidator{Field: "jr_no"},
				form.BetweenLengthValidator{Field: "jr_no"},
			},
		}},
		form.TextField{BaseField: form.BaseField{
			Key:          "ref_no",
			Type:         "text",
			Label:        "Refference No",
			Required:     true,
			Placeholder:  "30",
			HelpText:     "Enter Refference no",
			DefaultValue: "SALE-25",
			Validators: []form.Validator{
				form.RequiredValidator{Field: "ref_no"},
				form.BetweenLengthValidator{Field: "ref_no"},
			},
		}},
	)

	inst := form.NewInstance(jr)
	fields := form.BuildUIFields(jr, inst)
	fieldMap := form.BuildUIFieldMap(fields)

	// Build the ledger option list for the entry rows' <select> widgets.
	ledgers, err := h.svc.ListAllLedgers(c.Context())
	if err != nil {
		return response.Error(c, err)
	}
	ledgerOptions := make([]form.Option, 0, len(ledgers))
	for _, l := range ledgers {
		ledgerOptions = append(ledgerOptions, form.Option{
			Value: strconv.Itoa(l.ID),
			Label: l.Name,
		})
	}

	// initialRows := 2

	// type Props struct{
	// 	title string
	// 	fields []form.UIField
	// 	fieldMap map[string]form.UIField
	// 	initialRows int8
	// 	ledgers []form.Option
	// }

	// props := Props{
	// 	title: "New Journal",
	// 	fields: fields,
	// 	fieldMap: fieldMap,
	// 	initialRows: 2,
	// 	ledgers:  []form.Option,

	// }
	props := views.JournalProps{
		Title:       "New Journal",
		Fields:      fields,
		FieldMap:    fieldMap,
		InitialRows: 2,
		Ledgers:     ledgerOptions,
	}

	// fmt.Printf("%s", props)

	return render.Component(c, views.JournalNew(props))
	// return render.Component(c, views.JournalEntryForm(ledgerOptions, initialRows))
}

func (h *AccountingHandlers) AddRow(c *fiber.Ctx) error {
	index := c.QueryInt("index", 1)
	ledgers, err := h.svc.ListAllLedgers(c.Context())
	if err != nil {
		return response.Error(c, err)
	}
	ledgerOptions := make([]form.Option, 0, len(ledgers))
	for _, l := range ledgers {
		ledgerOptions = append(ledgerOptions, form.Option{
			Value: strconv.Itoa(l.ID),
			Label: l.Name,
		})
	}

	return datastar.MergeFragmentTempl(
		c,
		views.JournalEntryRow(index, ledgerOptions),

		datastar.WithSelector("#entry-body"),
		datastar.WithModeAppend(),
	)

}
