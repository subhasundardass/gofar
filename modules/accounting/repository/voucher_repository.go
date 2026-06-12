package repository

import (
	"context"

	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/ent/journal"
)

type VoucherRepository struct {
	db *ent.Client
}

func NewVoucherRepository(db *ent.Client) *VoucherRepository {
	return &VoucherRepository{db: db}
}

// ── Queries ───────────────────────────────────────────────────────────────────

func (r *VoucherRepository) GetByID(
	ctx context.Context,
	id int,
) (*ent.Journal, error) {

	return r.db.Journal.
		Query().
		Where(journal.IDEQ(id)).
		Only(ctx)
}

func (r *VoucherRepository) GetByIDWithLines(
	ctx context.Context,
	id int,
) (*ent.Journal, error) {

	return r.db.Journal.
		Query().
		Where(journal.IDEQ(id)).
		WithLines().
		Only(ctx)
}

func (r *VoucherRepository) List(
	ctx context.Context,
) ([]*ent.Journal, error) {

	return r.db.Journal.
		Query().
		Order(ent.Desc(journal.FieldID)).
		All(ctx)
}

func (r *VoucherRepository) ListPaginated(
	ctx context.Context,
	offset int,
	limit int,
) ([]*ent.Journal, int, error) {

	query := r.db.Journal.Query()

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.
		Order(ent.Desc(journal.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *VoucherRepository) Exists(
	ctx context.Context,
	id int,
) (bool, error) {

	return r.db.Journal.
		Query().
		Where(journal.IDEQ(id)).
		Exist(ctx)
}

// ── Actions ───────────────────────────────────────────────────────────────────

func (r *VoucherRepository) Create(
	ctx context.Context,
	input *ent.Journal,
) (*ent.Journal, error) {

	return r.db.Journal.
		Create().
		SetVoucherNo(input.VoucherNo).
		SetVoucherDate(input.VoucherDate).
		// SetNarration(input.Narration).
		Save(ctx)
}

func (r *VoucherRepository) Update(
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

func (r *VoucherRepository) Delete(
	ctx context.Context,
	id int,
) error {

	return r.db.Journal.
		DeleteOneID(id).
		Exec(ctx)
}
