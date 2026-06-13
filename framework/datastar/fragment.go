package datastar

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/starfederation/datastar-go/datastar"
)

// MergeFragments pushes a raw HTML string as a patched element to the client.
func MergeFragments(c *fiber.Ctx, fragment string, opts ...datastar.PatchElementOption) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.PatchElements(fragment, opts...)
	})
}

// MergeFragmentf pushes a printf-formatted HTML string as a patched element.
func MergeFragmentf(c *fiber.Ctx, format string, args ...any) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.PatchElementf(format, args...)
	})
}

// RemoveFragments pushes a remove event for the given CSS selector.
func RemoveFragments(c *fiber.Ctx, selector string) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.RemoveElement(selector)
	})
}

// MergeFragmentTemplSSE merges a templ component into an already-open SSE
// generator. Use this inside an SSE callback when sending several events in
// one response.
func MergeFragmentTemplSSE(sse *datastar.ServerSentEventGenerator, component templ.Component, opts ...datastar.PatchElementOption) error {
	return sse.PatchElementTempl(component, opts...)
}

// MergeFragmentTempl opens an SSE stream and immediately pushes a single templ
// component as a patched element.
// If you don't specify any options, Datastar will use its default patch behavior.
func MergeFragmentTempl(c *fiber.Ctx, component templ.Component, opts ...datastar.PatchElementOption) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.PatchElementTempl(component, opts...)
	})
}

// Fragmentf is a convenience wrapper to format an HTML fragment string.
func Fragmentf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
