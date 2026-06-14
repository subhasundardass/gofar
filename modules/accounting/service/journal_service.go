package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/services"
	"github.com/subhasundardas/gofar/framework/utils"
	"github.com/subhasundardas/gofar/modules/accounting/repository"
)

type JournalServices struct {
	repo *repository.JournalRepository
}

type JournalLineInput struct {
	EntryType string // "D"/"DR" for debit, "C"/"CR" for credit
	LedgerID  int
	Desc      string
	Amount    float64
}

type CreateJournalInput struct {
	VoucherNo   string
	VoucherType string
	Date        time.Time
	ReferenceNo string
	Narration   string
	Rows        []JournalLineInput
}

func NewJournalServices(repo *repository.JournalRepository) *JournalServices {
	return &JournalServices{
		repo: repo,
	}
}

func (s *JournalServices) ListJournal(ctx context.Context, params data.PaginationParams) (*data.PaginatedResult[*ent.Journal], error) {
	return services.Paginate(
		ctx,
		params,
		s.repo.ListJournalPaginated,
	)
}

func (s *JournalServices) CreateJournal(
	ctx context.Context,
	input CreateJournalInput,
) (*ent.Journal, error) {

	// 1. Validate balance
	var totalDebit, totalCredit float64
	for i, row := range input.Rows {
		if isBlankJournalLine(row) {
			continue
		}

		entryType, err := normalizeJournalEntryType(row.EntryType)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+1, err)
		}
		if row.LedgerID <= 0 {
			return nil, fmt.Errorf("row %d: ledger is required", i+1)
		}
		if row.Amount <= 0 {
			return nil, fmt.Errorf("row %d: amount must be greater than zero", i+1)
		}

		if entryType == "D" {
			totalDebit += row.Amount
		} else {
			totalCredit += row.Amount
		}
	}
	if math.Abs(totalDebit-totalCredit) > 0.000001 {
		return nil, fmt.Errorf(
			"journal not balanced: debit %.2f != credit %.2f",
			totalDebit, totalCredit,
		)
	}

	// 2. Build lines before touching the database so validation failures do
	// not leave an orphan journal header behind.
	lines, err := s.createJournalLines(input.Rows)
	if err != nil {
		return nil, err
	}

	// 3. Build journal header
	journalInput := &ent.Journal{
		Date:          input.Date,
		VoucherType:   input.VoucherType,
		VoucherNo:     input.VoucherNo,
		VoucherDate:   input.Date,
		ReferenceNo:   utils.NullableString(input.ReferenceNo),
		Narration:     utils.NullableString(input.Narration),
		JournalStatus: "DRAFT",
		TotalDebit:    totalDebit,
		TotalCredit:   totalCredit,
	}

	// 4. Save journal header and lines atomically.
	journal, err := s.repo.CreateWithLines(ctx, journalInput, lines)
	if err != nil {
		return nil, fmt.Errorf("create journal with lines: %w", err)
	}

	return journal, nil
}

// createJournalLines builds []*ent.Journal_Line from input rows
func (s *JournalServices) createJournalLines(
	rows []JournalLineInput,
) ([]*ent.Journal_Line, error) {
	lines := make([]*ent.Journal_Line, 0, len(rows))

	for i, row := range rows {
		if isBlankJournalLine(row) {
			continue
		}

		debit := 0.0
		credit := 0.0

		entryType, err := normalizeJournalEntryType(row.EntryType)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+1, err)
		}
		if row.LedgerID <= 0 {
			return nil, fmt.Errorf("row %d: ledger is required", i+1)
		}
		if row.Amount <= 0 {
			return nil, fmt.Errorf("row %d: amount must be greater than zero", i+1)
		}

		if entryType == "D" {
			debit = row.Amount
		} else {
			credit = row.Amount
		}

		lines = append(lines, &ent.Journal_Line{
			LedgerID:    row.LedgerID,
			Debit:       debit,
			Credit:      credit,
			Description: utils.NullableString(strings.TrimSpace(row.Desc)),
			LineNo:      len(lines) + 1,
		})
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid journal lines provided")
	}

	return lines, nil
}

func isBlankJournalLine(row JournalLineInput) bool {
	return strings.TrimSpace(row.EntryType) == "" &&
		row.LedgerID == 0 &&
		strings.TrimSpace(row.Desc) == "" &&
		row.Amount == 0
}

func normalizeJournalEntryType(entryType string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(entryType)) {
	case "D", "DR", "DEBIT":
		return "D", nil
	case "C", "CR", "CREDIT":
		return "C", nil
	default:
		return "", fmt.Errorf("invalid entry type %q", entryType)
	}
}
