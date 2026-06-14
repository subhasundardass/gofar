package repository

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/ent/journal"
)

type JournalRepository struct {
	db *ent.Client
}

func NewJournalRepository(db *ent.Client) *JournalRepository {
	return &JournalRepository{db: db}
}

// ── Queries ───────────────────────────────────────────────────────────────────

func (r *JournalRepository) GetByID(
	ctx context.Context,
	id int,
) (*ent.Journal, error) {

	return r.db.Journal.
		Query().
		Where(journal.IDEQ(id)).
		Only(ctx)
}

func (r *JournalRepository) GetByIDWithLines(
	ctx context.Context,
	id int,
) (*ent.Journal, error) {

	return r.db.Journal.
		Query().
		Where(journal.IDEQ(id)).
		WithLines().
		Only(ctx)
}

func (r *JournalRepository) List(
	ctx context.Context,
) ([]*ent.Journal, error) {

	return r.db.Journal.
		Query().
		Order(ent.Desc(journal.FieldID)).
		All(ctx)
}

func (r *JournalRepository) ListJournalPaginated(ctx context.Context, offset, limit int) ([]*ent.Journal, int, error) {
	query := r.db.Journal.Query()

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	items, err := query.
		Order(journal.ByID(sql.OrderDesc())).
		Limit(limit).
		Offset(offset).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *JournalRepository) Exists(
	ctx context.Context,
	id int,
) (bool, error) {

	return r.db.Journal.
		Query().
		Where(journal.IDEQ(id)).
		Exist(ctx)
}

// ── Actions ───────────────────────────────────────────────────────────────────

func (r *JournalRepository) Create(
	ctx context.Context,
	input *ent.Journal,
) (*ent.Journal, error) {
	return createJournal(ctx, r.db, input)
}

func (r *JournalRepository) CreateWithLines(
	ctx context.Context,
	input *ent.Journal,
	lines []*ent.Journal_Line,
) (journal *ent.Journal, err error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("journal lines are required")
	}

	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin journal transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	client := tx.Client()
	journal, err = createJournal(ctx, client, input)
	if err != nil {
		return nil, fmt.Errorf("create journal: %w", err)
	}

	if err = createLineBulk(ctx, client, journal.ID, lines); err != nil {
		return nil, fmt.Errorf("create journal lines: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit journal transaction: %w", err)
	}

	return journal, nil
}

func createJournal(ctx context.Context, client *ent.Client, input *ent.Journal) (*ent.Journal, error) {
	return client.Journal.
		Create().
		SetDate(input.Date).
		SetVoucherType(input.VoucherType).
		SetVoucherNo(input.VoucherNo).
		SetVoucherDate(input.VoucherDate).
		SetNillableReferenceNo(input.ReferenceNo).
		SetNillableNarration(input.Narration).
		SetJournalStatus(input.JournalStatus).
		SetTotalDebit(input.TotalDebit).
		SetTotalCredit(input.TotalCredit).
		Save(ctx)
}

func (r *JournalRepository) CreateLineBulk(
	ctx context.Context,
	lines []*ent.Journal_Line,
) error {
	return createLineBulk(ctx, r.db, 0, lines)
}

func createLineBulk(
	ctx context.Context,
	client *ent.Client,
	journalID int,
	lines []*ent.Journal_Line,
) error {
	builders := make([]*ent.JournalLineCreate, 0, len(lines))

	for i, line := range lines {
		lineJournalID := line.JournalID
		if journalID > 0 {
			lineJournalID = journalID
		}
		if lineJournalID <= 0 {
			return fmt.Errorf("line %d: journal id is required", i+1)
		}

		lb := client.Journal_Line.
			Create().
			SetJournalID(lineJournalID).
			SetLedgerID(line.LedgerID).
			SetDebit(line.Debit).
			SetCredit(line.Credit).
			SetLineNo(i + 1).
			SetNillableDescription(line.Description)

		builders = append(builders, lb)
	}

	return client.Journal_Line.
		CreateBulk(builders...).
		Exec(ctx)
}

func (r *JournalRepository) Update(
	ctx context.Context,
	id int,
	input *ent.Journal,
) (*ent.Journal, error) {

	return r.db.Journal.
		UpdateOneID(id).
		SetVoucherDate(input.VoucherDate).
		// SetNarration(input.Narration).
		Save(ctx)
}

func (r *JournalRepository) Delete(
	ctx context.Context,
	id int,
) error {

	return r.db.Journal.
		DeleteOneID(id).
		Exec(ctx)
}
