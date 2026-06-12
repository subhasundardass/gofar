// Package render provides helpers for rendering templ components and raw HTML
// strings to a Fiber response.
//
// GoFar uses the a-h/templ library for type-safe, compiled HTML templates.
// These helpers ensure the correct Content-Type is set and that rendering
// errors are surfaced cleanly rather than producing a partial response.
//
// Example – render a templ component:
//
//	func DashboardPage(c *fiber.Ctx) error {
//	    return render.Component(c, views.Dashboard(data))
//	}
//
// Example – render a component inside an SSE stream (Datastar):
//
//	sse := datastar.New(c)
//	return render.SSEFragment(sse, "main-content", views.UserCard(user))
package render

import (
	"bytes"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

// Component renders a templ component directly into the Fiber response writer.
// Sets Content-Type to "text/html; charset=utf-8".
//
// This is the standard way to render a full-page templ component:
//
//	return render.Component(c, layouts.Base(views.HomePage()))
func Component(c *fiber.Ctx, component templ.Component) error {
	c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	return component.Render(c.Context(), c.Response().BodyWriter())
}

// ComponentToString renders a templ component to a string without writing to
// the response. Useful when you need the HTML as a string (e.g. to pass to an
// SSE fragment or an email body).
func ComponentToString(component templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(nil, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// HTML writes a raw HTML string to the response with the correct Content-Type.
// Prefer templ components for new code; this helper exists for edge-cases like
// rendering a pre-built HTML string from a legacy source.
func HTML(c *fiber.Ctx, html string) error {
	c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	return c.SendString(html)
}

// Text writes a plain-text response.
func Text(c *fiber.Ctx, text string) error {
	c.Set(fiber.HeaderContentType, "text/plain; charset=utf-8")
	return c.SendString(text)
}
