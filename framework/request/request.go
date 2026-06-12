// Package request provides helpers for parsing and validating incoming HTTP
// requests in GoFar Fiber handlers.
//
// The generic Parse[T] function combines body parsing and struct validation
// in a single call, keeping handlers concise:
//
//	type CreateUserDTO struct {
//	    Name  string `json:"name"  validate:"required,min=2"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	func CreateUser(c *fiber.Ctx) error {
//	    dto, err := request.Parse[CreateUserDTO](c)
//	    if err != nil {
//	        return response.Error(c, err)
//	    }
//	    // dto is *CreateUserDTO, validated and ready to use
//	}
package request

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	apperrors "github.com/subhasundardas/gofar/framework/errors"
)

// Parse deserialises the request body into a new value of type T using Fiber's
// built-in BodyParser (supports JSON, form, multipart depending on Content-Type).
//
// Returns a *T on success or an *apperrors.AppError (400) on parse failure.
func Parse[T any](c *fiber.Ctx) (*T, error) {
	var dto T
	if err := c.BodyParser(&dto); err != nil {
		return nil, apperrors.NewBadRequest("invalid request body: " + err.Error())
	}
	return &dto, nil
}

// ParseQuery deserialises the query string into a new value of type T using
// Fiber's QueryParser. Useful for filter/search/sort DTOs.
//
//	type ListUsersQuery struct {
//	    Page  int    `query:"page"`
//	    Limit int    `query:"limit"`
//	    Sort  string `query:"sort"`
//	}
//
//	q, err := request.ParseQuery[ListUsersQuery](c)
func ParseQuery[T any](c *fiber.Ctx) (*T, error) {
	var dto T
	if err := c.QueryParser(&dto); err != nil {
		return nil, apperrors.NewBadRequest("invalid query parameters: " + err.Error())
	}
	return &dto, nil
}

// ParamInt reads a named route parameter and converts it to int.
// Returns a 400 AppError if the parameter is missing or not an integer.
//
//	id, err := request.ParamInt(c, "id")
func ParamInt(c *fiber.Ctx, name string) (int, error) {
	raw := c.Params(name)
	if raw == "" {
		return 0, apperrors.NewBadRequest("missing route parameter: " + name)
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, apperrors.NewBadRequest("route parameter " + name + " must be an integer")
	}
	return n, nil
}

// ParamString reads a named route parameter as a string.
// Returns a 400 AppError if the parameter is missing.
func ParamString(c *fiber.Ctx, name string) (string, error) {
	v := c.Params(name)
	if v == "" {
		return "", apperrors.NewBadRequest("missing route parameter: " + name)
	}
	return v, nil
}

// QueryInt reads a query string parameter and converts it to int.
// Returns def if the parameter is absent or cannot be parsed.
//
//	page := request.QueryInt(c, "page", 1)
func QueryInt(c *fiber.Ctx, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

// QueryString reads a query parameter as a string, returning def if absent.
func QueryString(c *fiber.Ctx, key, def string) string {
	if v := c.Query(key); v != "" {
		return v
	}
	return def
}

// BearerToken extracts the JWT or token from the Authorization header.
// Supports "Bearer <token>" format. Returns an empty string if the header
// is absent or malformed.
func BearerToken(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}
