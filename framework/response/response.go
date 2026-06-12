// Package response provides helper functions for sending consistent JSON
// responses from Fiber handlers.
//
// All responses follow the envelope pattern:
//
//	{ "success": true,  "data": {...} }          // success
//	{ "success": false, "message": "...",
//	  "code": "notFound", "errors": {...} }       // error
//
// Example usage in a handler:
//
//	func GetUser(c *fiber.Ctx) error {
//	    user, err := userRepo.Find(c.Params("id"))
//	    if err != nil {
//	        return response.NotFound(c, "user")
//	    }
//	    return response.OK(c, user)
//	}
package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	apperrors "github.com/subhasundardas/gofar/framework/errors"
)

// ── Success responses ─────────────────────────────────────────────────────────

// OK sends a 200 response with data wrapped in the success envelope.
func OK(c *fiber.Ctx, data any) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// Created sends a 201 response. Use after successfully creating a resource.
func Created(c *fiber.Ctx, data any) error {
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// NoContent sends a 204 response with no body. Use after DELETE or an action
// that produces no meaningful response payload.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

// Paginated sends a 200 response with both data and pagination metadata.
//
//	response.Paginated(c, users, response.Pager{Page: 1, PerPage: 20, Total: 350})
func Paginated(c *fiber.Ctx, data any, pager Pager) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success":    true,
		"data":       data,
		"pagination": pager,
	})
}

// Pager carries pagination metadata returned with list responses.
type Pager struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewPager constructs a Pager. page and perPage are 1-indexed.
func NewPager(page, perPage, total int) Pager {
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}
	return Pager{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}
}

// ── Error responses ───────────────────────────────────────────────────────────

// Error sends an appropriate HTTP error response for any error value.
// If err is an *apperrors.AppError the stored status code is used; otherwise
// a generic 500 is returned. The internal cause is never exposed.
func Error(c *fiber.Ctx, err error) error {
	var appErr *apperrors.AppError
	if apperrors.As(err, &appErr) {
		body := fiber.Map{
			"success": false,
			"code":    appErr.Code,
			"message": appErr.Message,
		}
		// Include field-level validation errors if present.
		var valErr *apperrors.ValidationError
		if apperrors.As(err, &valErr) {
			body["errors"] = valErr.Fields
		}
		return c.Status(appErr.Status).JSON(body)
	}
	// Unknown error – return generic 500 without leaking details.
	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"code":    "internal",
		"message": "an internal server error occurred",
	})
}

// BadRequest sends a 400 with the given message.
func BadRequest(c *fiber.Ctx, message string) error {
	return Error(c, apperrors.NewBadRequest(message))
}

// Unauthorized sends a 401.
func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, apperrors.NewUnauthorized(message))
}

// Forbidden sends a 403.
func Forbidden(c *fiber.Ctx, message string) error {
	return Error(c, apperrors.NewForbidden(message))
}

// NotFound sends a 404 for the named resource.
func NotFound(c *fiber.Ctx, resource string) error {
	return Error(c, apperrors.NewNotFound(resource, ""))
}

// ValidationFailed sends a 422 with per-field validation errors.
func ValidationFailed(c *fiber.Ctx, fields map[string]string) error {
	return Error(c, apperrors.NewValidation(fields))
}
