package service

import (
	"context"

	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/framework/data"
	"github.com/subhasundardas/gofar/framework/services"
	"github.com/subhasundardas/gofar/modules/accounting/repository"
)

type VoucherServices struct {
	repo *repository.VoucherRepository
}

func NewVoucherServices(repo *repository.VoucherRepository) *VoucherServices {
	return &VoucherServices{
		repo: repo,
	}
}

func (s *VoucherServices) ListVouchers(
	ctx context.Context,
	params data.PaginationParams,
) (*data.PaginatedResult[*ent.Journal], error) {

	return services.Paginate(
		ctx,
		params,
		s.repo.ListPaginated,
	)
}

func (s *VoucherServices) ListVouchersByLedger(ctx context.Context, params data.PaginationParams, ledgerId int) (*data.PaginatedResult[*ent.Journal], error) {
	return services.Paginate(
		ctx,
		params,
		s.repo.ListPaginated,
	)
}
