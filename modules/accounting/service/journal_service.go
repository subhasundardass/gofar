package service

import (
	"context"

	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/services"
	"github.com/subhasundardas/gofar/modules/accounting/repository"
)

type JournalServices struct {
	repo *repository.JournalRepository
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
