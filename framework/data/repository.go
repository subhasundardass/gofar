package data

import (
	"context"
)

// Entity is a constraint for any Ent-generated entity struct.
// All Ent entities have an integer ID field.
type Entity interface {
	any
}

// Repository defines the standard CRUD contract every module repo should satisfy.
// T is the entity type (e.g. *ent.Customer), C is the create-input type.
type Repository[T Entity] interface {
	List(ctx context.Context) ([]T, error)
	FindByID(ctx context.Context, id int) (T, error)
	Delete(ctx context.Context, id int) error
}

// PaginatedResult is a standard paginated response wrapper.
type PaginatedResult[T Entity] struct {
	Items      []T
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

// PaginationParams carries page/limit from query strings.
type PaginationParams struct {
	Page    int
	PerPage int
	Search  string
	OrderBy string
	Desc    bool
}

// DefaultPagination returns safe defaults.
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 20,
		OrderBy: "id",
		Desc:    false,
	}
}

// Offset calculates the DB offset from page/perPage.
func (p PaginationParams) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}

// TotalPages calculates total pages from total count.
func CalcTotalPages(total, perPage int) int {
	if perPage == 0 {
		return 0
	}
	pages := total / perPage
	if total%perPage != 0 {
		pages++
	}
	return pages
}
