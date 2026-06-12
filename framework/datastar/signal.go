package datastar

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/starfederation/datastar-go/datastar"
)

// MergeSignals pushes a raw JSON byte slice as a patch-signals event.
func MergeSignals(c *fiber.Ctx, signalsJSON []byte, opts ...datastar.PatchSignalsOption) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.PatchSignals(signalsJSON, opts...)
	})
}

// MarshalAndMergeSignals marshals any value to JSON and pushes it as a
// patch-signals event.
func MarshalAndMergeSignals(c *fiber.Ctx, signals any, opts ...datastar.PatchSignalsOption) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.MarshalAndPatchSignals(signals, opts...)
	})
}

// MarshalAndMergeSignalsIfMissing is the same as MarshalAndMergeSignals but
// only sets each signal when it is not already present on the client.
func MarshalAndMergeSignalsIfMissing(c *fiber.Ctx, signals any, opts ...datastar.PatchSignalsOption) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.MarshalAndPatchSignalsIfMissing(signals, opts...)
	})
}

// RemoveSignals removes the named signal paths from the client store.
// v1 has no RemoveSignals — patch with null values instead.
func RemoveSignals(c *fiber.Ctx, paths ...string) error {
	return SSE(c, func(sse *datastar.ServerSentEventGenerator) error {
		nulls := make(map[string]any, len(paths))
		for _, p := range paths {
			nulls[p] = nil
		}
		return sse.MarshalAndPatchSignals(nulls)
	})
}

// ReadSignals deserializes the incoming datastar signals from the request body
// or query parameter into the provided struct pointer.
func ReadSignals(c *fiber.Ctx, signals any) error {
	_, r, err := fiberToHTTP(c)
	if err != nil {
		return fmt.Errorf("fiberToHTTP: %w", err)
	}
	return datastar.ReadSignals(r, signals)
}
