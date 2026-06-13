// Package service contains business logic for the Accounting module.
// Services sit between handlers and repositories — enforce domain rules here.
package service

import (
	"context"
	"fmt"

	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/services"
	"github.com/subhasundardas/gofar/modules/accounting/repository"
)

// Services bundles all Accounting domain services.
type Services struct {
	Accounting *AccountingService
	Journal    *JournalServices
}

// NewServices constructs all services, injecting the shared repository group.
func NewServices(repos *repository.Repositories) *Services {
	return &Services{
		Accounting: NewAccountingService(repos.Accounting),
		Journal:    NewJournalServices(repos.Journal),
	}
}

// ==============================================================================
// AccountingService implements all Accounting use-cases.
type AccountingService struct {
	repo *repository.AccountingRepository
}

// NewAccountingService constructs a AccountingService.
func NewAccountingService(repo *repository.AccountingRepository) *AccountingService {
	return &AccountingService{repo: repo}
}

// Create validates input and persists a new Accounting record.
func (s *AccountingService) Create(name string) (int, error) {
	if name == "" {
		return 0, fmt.Errorf("accounting: name is required")
	}
	return s.repo.Create(name)
}

// List returns all Accounting records.
func (s *AccountingService) List() ([]repository.Accounting, error) {
	return s.repo.List()
}

// Get returns a single record. Returns an error if not found.
func (s *AccountingService) Get(id int) (*repository.Accounting, error) {

	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("accounting: get %d: %w", id, err)
	}
	if item == nil {
		return nil, fmt.Errorf("accounting: %d not found", id)
	}
	return item, nil
}

// Update validates and persists changes to an existing record.
func (s *AccountingService) Update(id int, name string) error {
	if name == "" {
		return fmt.Errorf("accounting: name is required")
	}
	return s.repo.Update(id, name)
}

// Delete permanently removes a record.
func (s *AccountingService) Delete(id int) error {
	return s.repo.Delete(id)
}

// =========Account Group
func (s *AccountingService) ListGroups(ctx context.Context, params data.PaginationParams) (*data.PaginatedResult[*ent.Acct_Group], error) {
	return services.Paginate(
		ctx,
		params,
		s.repo.ListGroupsPaginated,
	)
}

func (s *AccountingService) ListLedgers(ctx context.Context, params data.PaginationParams) (*data.PaginatedResult[*ent.Ledger], error) {
	return services.Paginate(
		ctx,
		params,
		s.repo.ListLedgersPaginated,
	)
}

// ListAllLedgers returns every ledger ordered by code. Used by dropdowns
// where pagination is unnecessary.
func (s *AccountingService) ListAllLedgers(ctx context.Context) ([]*ent.Ledger, error) {
	return s.repo.ListAllLedgers(ctx)
}
