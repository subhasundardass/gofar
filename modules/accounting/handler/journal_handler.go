package handler

import (
	"fmt"
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

type JournalHandlers struct {
	journalSvc *service.JournalServices
	accSvc     *service.AccountingService
}

func NewJournalHandlers(acc *service.AccountingService, jr *service.JournalServices) *JournalHandlers {
	return &JournalHandlers{
		journalSvc: jr,
		accSvc:     acc,
	}
}

// -- Journal
func (h *JournalHandlers) ListJournal(c *fiber.Ctx) error {
	params := data.PaginationParams{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 20),
		Search:  c.Query("q"),
	}
	result, err := h.journalSvc.ListJournal(c.Context(), params)
	if err != nil {
		return response.Error(c, err)
	}

	return render.Component(c, views.JournalView("Journal Master", result))
}

func (h *JournalHandlers) NewJournal(c *fiber.Ctx) error {
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
	ledgers, err := h.accSvc.ListAllLedgers(c.Context())
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

	props := views.JournalProps{
		Title:       "New Journal",
		Fields:      fields,
		FieldMap:    fieldMap,
		InitialRows: 2,
		Ledgers:     ledgerOptions,
	}

	return render.Component(c, views.JournalNew(props))
}

func (h *JournalHandlers) AddRow(c *fiber.Ctx) error {
	index := c.QueryInt("index", 1)
	ledgers, err := h.accSvc.ListAllLedgers(c.Context())
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

// Create
func (h *JournalHandlers) Create(c *fiber.Ctx) error {
	var signals map[string]interface{}
	if err := c.BodyParser(&signals); err != nil {
		return response.Error(c, err)
	}

	getString := func(key string) string {
		if v, ok := signals[key]; ok {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}

	getFloat := func(key string) float64 {
		if v, ok := signals[key]; ok {
			switch val := v.(type) {
			case float64:
				return val
			case string:
				f, _ := strconv.ParseFloat(val, 64)
				return f
			}
		}
		return 0
	}

	getInt := func(key string) int {
		return int(getFloat(key))
	}

	// --- Parse date ---
	jrDate, err := time.Parse("2006-01-02", getString("jr_date"))
	if err != nil {
		jrDate = time.Now()
	}

	// --- Parse rows ---
	rowCount := getInt("rowCount")
	rows := make([]service.JournalLineInput, 0, rowCount)

	for i := 1; i <= rowCount; i++ {
		rows = append(rows, service.JournalLineInput{
			EntryType: getString(fmt.Sprintf("rows_%d_type", i)),
			LedgerID:  getInt(fmt.Sprintf("rows_%d_account", i)),
			Desc:      getString(fmt.Sprintf("rows_%d_desc", i)),
			Amount:    getFloat(fmt.Sprintf("rows_%d_amount", i)),
		})
	}

	// --- Call service ---
	journal, err := h.journalSvc.CreateJournal(c.Context(), service.CreateJournalInput{
		VoucherNo:   getString("jr_no"),
		VoucherType: "JV",
		Date:        jrDate,
		ReferenceNo: getString("ref_no"),
		Rows:        rows,
	})
	if err != nil {
		fmt.Println("❌ CreateJournal error:", err) // <-- add this
		return response.Error(c, err)
	}

	return response.Created(c, fiber.Map{
		"status":     "created",
		"journal_id": journal.ID,
		"voucher_no": journal.VoucherNo,
	})
}
