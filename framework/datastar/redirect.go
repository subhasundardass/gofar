package datastar

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/starfederation/datastar-go/datastar"
)

// Redirect pushes a client-side redirect via a datastar execute-script event.
func Redirect(c *fiber.Ctx, url string, opts ...datastar.ExecuteScriptOption) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.Redirect(url, opts...)
	})
}

// Redirectf is the same as Redirect but accepts a printf-style format string.
func Redirectf(c *fiber.Ctx, format string, args ...any) error {
	return Redirect(c, fmt.Sprintf(format, args...))
}
