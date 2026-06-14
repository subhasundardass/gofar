package service

import (
	"context"
	"fmt"
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
	EntryType string // "DR" or "CR"
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
	for _, row := range input.Rows {
		if row.EntryType == "D" {
			totalDebit += row.Amount
		} else {
			totalCredit += row.Amount
		}
	}
	if totalDebit != totalCredit {
		return nil, fmt.Errorf(
			"journal not balanced: debit %.2f != credit %.2f",
			totalDebit, totalCredit,
		)
	}

	// 2. Build journal header
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

	// 3. Save journal header
	journal, err := s.repo.Create(ctx, journalInput)
	if err != nil {
		return nil, fmt.Errorf("create journal: %w", err)
	}

	// 4. Build lines
	lines, err := s.createJournalLines(ctx, journal.ID, input.Rows)
	if err != nil {
		return nil, err
	}

	// 5. Bulk save lines
	if err := s.repo.CreateLineBulk(ctx, lines); err != nil {
		return nil, fmt.Errorf("create journal lines: %w", err)
	}

	return journal, nil
}

// createJournalLines builds []*ent.Journal_Line from input rows
func (s *JournalServices) createJournalLines(
	ctx context.Context,
	journalID int,
	rows []JournalLineInput,
) ([]*ent.Journal_Line, error) {
	lines := make([]*ent.Journal_Line, 0, len(rows))

	for i, row := range rows {
		// Debug: remove after fix
		fmt.Printf("Row %d: EntryType=%s LedgerID=%d Amount=%.2f Desc=%s\n",
			i+1, row.EntryType, row.LedgerID, row.Amount, row.Desc)

		if row.LedgerID == 0 || row.Amount == 0 {
			fmt.Printf("Row %d SKIPPED: LedgerID=%d Amount=%.2f\n", i+1, row.LedgerID, row.Amount)
			continue
		}
		// ...
	}

	fmt.Printf("Total valid lines: %d\n", len(lines))

	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid journal lines provided")
	}

	return lines, nil
}
