package middleware

import (
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/config"
	"github.com/subhasundardas/gofar/framework/logger"
)

type MiddlewareRegistry struct {
	middlewares []fiber.Handler
}

func New(app *fiber.App, cfg *config.Config, logger logger.Logger) *MiddlewareRegistry {

	r := &MiddlewareRegistry{
		middlewares: make([]fiber.Handler, 0),
	}

	r.Register(RecoveryMiddleware(logger))
	// Disabled
	// r.Register(LoggingMiddleware(logger))
	r.Register(CORSMiddleware(cfg))
	r.Register(RequestTimingMiddleware(logger))

	// IMPORTANT: attach to fiber app
	for _, m := range r.middlewares {
		app.Use(m)
	}

	return r
}

func (r *MiddlewareRegistry) Register(m fiber.Handler) {
	r.middlewares = append(r.middlewares, m)
}

// RecoveryMiddleware recovers from panics, logs the stack trace, and returns 500.
// response even after a panic.
func RecoveryMiddleware(logger logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) (retErr error) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.With(
					"error", rec,
					"stack", string(debug.Stack()),
				).Error("panic recovered")
				retErr = c.Status(fiber.StatusInternalServerError).
					JSON(fiber.Map{"error": "internal server error"})
			}
		}()

		return c.Next()
	}
}

// LoggingMiddleware logs each incoming request's method and path.
func LoggingMiddleware(logger logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger.With(
			"method", c.Method(),
			"path", c.Path(),
		).Info("request")

		return c.Next()
	}
}

// CORSMiddleware sets CORS headers using the configured app URL as the allowed origin.
func CORSMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", cfg.Get("CORS_ALLOWED_ORIGINS"))
		c.Set("Access-Control-Allow-Methods", cfg.Get("CORS_ALLOWED_METHODS"))
		c.Set("Access-Control-Allow-Headers", cfg.Get("CORS_ALLOWED_HEADERS"))

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		return c.Next()
	}
}

// RequestTimingMiddleware logs the method, path, status code, and duration of each request.
func RequestTimingMiddleware(logger logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		logger.With(
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).String(),
		).Info("response")

		return err
	}
}
