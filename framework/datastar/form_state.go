package datastar

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/form"
)

// =============================================================================
// FormState — Datastar integration
// =============================================================================
//
// These helpers bridge the form.FormState type with the Datastar SSE
// machinery. They are intentionally thin so a handler stays in control
// of validation order and error rendering.
//
// Typical handler:
//
//	func SignupHandler(c *fiber.Ctx) error {
//	    signup := lookupForm("signup")
//
//	    // 1. Pull the patch Datastar posted to us.
//	    patch, err := datastar.ReadFormState(c, signup.Name())
//	    if err != nil {
//	        return response.BadRequest(c, err.Error())
//	    }
//
//	    // 2. Hydrate an instance and validate.
//	    inst := form.NewInstance(signup).Apply(patch)
//	    inst.Validate()
//
//	    // 3. Push the post-validation state back so the client can
//	    //    re-render error messages / disable the submit button / etc.
//	    return datastar.PushFormState(c, form.NewFormState(signup.Name(), inst))
//	}

// ReadFormState deserialises the incoming Datastar request body into a
// new *form.FormState. The body is expected to be JSON encoded with the
// shape produced by form.FormState.MarshalJSON.
//
// An empty body is treated as an empty FormState (Name only) so a handler
// that just wants to know "what name did the client send" still works.
//
// The name argument is the server-known form name and ALWAYS overrides
// whatever the client sent — this prevents a malicious client from
// routing a patch into the wrong form.
func ReadFormState(c *fiber.Ctx, name string) (*form.FormState, error) {
	state := form.NewFormStateFromValues(name, nil)
	body := c.Body()
	if len(body) == 0 {
		state.SetName(name)
		return state, nil
	}
	if err := json.Unmarshal(body, state); err != nil {
		return nil, fmt.Errorf("datastar.ReadFormState: %w", err)
	}
	state.SetName(name)
	return state, nil
}

// PushFormState serialises a FormState and pushes it back to the client
// as a Datastar `merge-signals` event. The client will then re-render any
// reactive expressions that reference the signals — typically an
// "errors.flag" boolean, per-field error text, or a "submitting" spinner.
func PushFormState(c *fiber.Ctx, state *form.FormState) error {
	if state == nil {
		return nil
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("datastar.PushFormState: %w", err)
	}
	return MergeSignals(c, payload)
}
