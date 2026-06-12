package services

import (
	"context"
	"fmt"

	"github.com/subhasundardas/gofar/framework/data"
)

// BaseService provides common service-layer utilities every module service
// can embed. It is not coupled to any specific entity or repo.
type BaseService struct {
	Logger interface {
		Info(msg string, args ...any)
		Warn(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// WrapError adds context to an error without hiding it.
func (s *BaseService) WrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// ValidateRequired returns an error if any of the named fields are empty.
//
// Usage:
//
//	if err := s.ValidateRequired(map[string]string{
//	    "name":  req.Name,
//	    "email": req.Email,
//	}); err != nil {
//	    return err
//	}
func (s *BaseService) ValidateRequired(fields map[string]string) error {
	for field, value := range fields {
		if value == "" {
			return fmt.Errorf("field %q is required", field)
		}
	}
	return nil
}

// ValidatePositive returns an error if value is not greater than zero.
func (s *BaseService) ValidatePositive(field string, value float64) error {
	if value <= 0 {
		return fmt.Errorf("field %q must be positive", field)
	}
	return nil
}

// ValidateMinLength returns an error if value is shorter than min characters.
func (s *BaseService) ValidateMinLength(field, value string, min int) error {
	if len(value) < min {
		return fmt.Errorf("field %q must be at least %d characters", field, min)
	}
	return nil
}

// ValidateMaxLength returns an error if value exceeds max characters.
func (s *BaseService) ValidateMaxLength(field, value string, max int) error {
	if len(value) > max {
		return fmt.Errorf("field %q must not exceed %d characters", field, max)
	}
	return nil
}

// Paginate is a helper for service methods that return paginated results.
// fn is called with offset and limit; it returns items and total count.
func Paginate[T data.Entity](
	ctx context.Context,
	params data.PaginationParams,
	fn func(ctx context.Context, offset, limit int) ([]T, int, error),
) (*data.PaginatedResult[T], error) {
	items, total, err := fn(ctx, params.Offset(), params.PerPage)
	if err != nil {
		return nil, err
	}

	return &data.PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: data.CalcTotalPages(total, params.PerPage),
	}, nil
}
