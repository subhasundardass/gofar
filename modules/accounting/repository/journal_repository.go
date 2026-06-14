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
	return r.db.Journal.
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
	builders := make([]*ent.JournalLineCreate, 0, len(lines))

	for i, line := range lines {
		// Debug: remove after confirming
		fmt.Printf("Line %d: JournalID=%d LedgerID=%d Debit=%.2f Credit=%.2f\n",
			i+1, line.JournalID, line.LedgerID, line.Debit, line.Credit)

		lb := r.db.Journal_Line.
			Create().
			SetJournalID(line.JournalID).
			SetLedgerID(line.LedgerID).
			SetDebit(line.Debit).
			SetCredit(line.Credit).
			SetLineNo(i + 1).
			SetNillableDescription(line.Description)

		builders = append(builders, lb)
	}

	return r.db.Journal_Line.
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
