// Package repository is the data-access layer for the Accounting module.
// All DB queries live here. Services must never call the DB directly.
package repository

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/ent/acct_group"
	"github.com/subhasundardas/gofar/ent/ledger"
)

// Repositories bundles all Accounting repositories.
// Pass the whole struct to service.NewServices so adding a repo never changes signatures.
type Repositories struct {
	Accounting *AccountingRepository
	Voucher    *VoucherRepository
}

// NewRepositories constructs all repositories.
func NewRepositories(db *ent.Client) *Repositories {
	return &Repositories{
		Accounting: NewAccountingRepository(db),
		Voucher:    NewVoucherRepository(db),
	}
}

//==============================================================================

// Accounting is the read-model returned by repository queries.
// Replace with the generated ent.Accounting type once Ent is wired.
type Accounting struct {
	ID   int
	Name string
}

// AccountingRepository handles all Accounting persistence operations.
type AccountingRepository struct {
	db *ent.Client
}

// NewAccountingRepository constructs a AccountingRepository.
func NewAccountingRepository(db *ent.Client) *AccountingRepository {
	return &AccountingRepository{db: db}
}

// Create persists a new record and returns its generated ID.
func (r *AccountingRepository) Create(name string) (int, error) {
	// TODO: r.db.Accounting.Create().SetName(name).Save(ctx)
	return 0, nil
}

// List returns all records.
func (r *AccountingRepository) List() ([]Accounting, error) {
	// TODO: r.db.Accounting.Query().All(ctx)
	return []Accounting{}, nil
}

// FindByID returns one record or nil if not found.
func (r *AccountingRepository) FindByID(id int) (*Accounting, error) {
	// TODO: r.db.Accounting.Get(ctx, id)
	return nil, nil
}

// Update persists changes to an existing record.
func (r *AccountingRepository) Update(id int, name string) error {
	// TODO: r.db.Accounting.UpdateOneID(id).SetName(name).Exec(ctx)
	return nil
}

// Delete hard-deletes a record.
func (r *AccountingRepository) Delete(id int) error {
	// TODO: r.db.Accounting.DeleteOneID(id).Exec(ctx)
	return nil
}

// -- List Groups

// ListGroupsPaginated returns a page of account groups ordered by id DESC,
// along with the total count of groups matching the (currently unfiltered)
// query. The query is cloned for the count so filters added later on the
// returned builder are not applied twice.
func (r *AccountingRepository) ListGroupsPaginated(ctx context.Context, offset, limit int) ([]*ent.Acct_Group, int, error) {
	query := r.db.Acct_Group.Query()

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	items, err := query.
		Order(acct_group.ByID(sql.OrderDesc())).
		Limit(limit).
		Offset(offset).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *AccountingRepository) ListLedgersPaginated(ctx context.Context, offset, limit int) ([]*ent.Ledger, int, error) {
	query := r.db.Ledger.Query()

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	items, err := query.
		Order(ledger.ByID(sql.OrderDesc())).
		WithGroup().
		Limit(limit).
		Offset(offset).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
