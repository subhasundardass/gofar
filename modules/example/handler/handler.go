// Package handler contains HTTP handlers for the Example module.
// Handlers are thin: parse request → call service → return response.
package handler

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/subhasundardas/gofar/framework/datastar"
	"github.com/subhasundardas/gofar/framework/form"
	"github.com/subhasundardas/gofar/framework/render"
	"github.com/subhasundardas/gofar/modules/example/service"
	"github.com/subhasundardas/gofar/modules/example/views"
)

// Handlers bundles all Example handler sets.
type Handlers struct {
	Example *ExampleHandlers
}

// NewHandlers constructs all handler sets. Takes services only — never fiber.App.
func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		Example: &ExampleHandlers{
			counter: svc.Counter,
		},
	}
}

// ExampleHandlers holds all Example HTTP handlers.
type ExampleHandlers struct {
	counter *service.Counter
}

// Home renders the example module's landing page using templ.
// The page loads datastar.js and contains a button wired with
// data-on-click to POST /example/counter (see routes.go).
func (h *ExampleHandlers) Home(c *fiber.Ctx) error {
	h.counter.Reset()
	return render.Component(c, views.Home(h.counter.Current()))
}

// Increment is the datastar SSE endpoint. It increments the in-memory
// counter and streams back the new #counter fragment over SSE, which
// datastar.js merges in place on the client.
//
// Wire format: a single `merge-fragments` event whose data selector
// targets #counter. The framework's datastar helpers handle SSE
// setup, header negotiation, and templ rendering.

func (h *ExampleHandlers) Increment(c *fiber.Ctx) error {
	val := h.counter.Increment()
	log.Printf("counter value: %s", val)
	return datastar.MergeFragmentTempl(c, views.Counter(val))
}

func (h *ExampleHandlers) Reset(c *fiber.Ctx) error {
	h.counter.Reset()
	return datastar.MergeFragmentTempl(c, views.Counter("0"))
}

// registrationForm is built once at startup and reused across every request.
// The schema is immutable; per-request state lives on *FormState / *Instance.
var registrationForm = form.New("registration").Fields(

	// ── Personal details ────────────────────────────────────────────────────

	form.TextInput("first_name",
		form.Label("First Name"),
		form.Placeholder("Jane"),
		form.Required(),
		form.WithValidator(form.BetweenLengthValidator{Field: "first_name", Min: 1, Max: 50}),
	),

	form.TextInput("last_name",
		form.Label("Last Name"),
		form.Placeholder("Doe"),
		form.Required(),
		form.WithValidator(form.BetweenLengthValidator{Field: "last_name", Min: 1, Max: 50}),
	),

	form.NumberInput("age",
		form.Label("Age"),
		form.Placeholder("30"),
		form.HelpText("Must be 18 or older"),
		form.Default(25),
		form.WithValidator(form.BetweenValueValidator{Field: "age", Min: 18, Max: 120}),
	),

	form.EmailInput("email",
		form.Label("Email"),
		form.Placeholder("jane@example.com"),
		form.Required(),
		form.WithValidator(form.EmailValidator{}),
	),

	// ── Account type ────────────────────────────────────────────────────────
	// Drives the conditional fields below via VisibleIf / RequiredIf.

	form.TextInput("account_type",
		form.Label("Account Type"), // "personal" | "business"
		form.Default("personal"),
		form.Required(),
	),

	// Company name — only shown and required when account_type = "business".
	form.TextInput("company_name",
		form.Label("Company Name"),
		form.Placeholder("Acme Ltd"),
		form.VisibleIf(form.FieldEquals("account_type", "business")),
		form.RequiredIf(form.FieldEquals("account_type", "business")),
		form.WithValidator(form.BetweenLengthValidator{Field: "company_name", Min: 2, Max: 100}),
	),

	// VAT number — visible only for business accounts in certain countries.
	form.TextInput("vat_number",
		form.Label("VAT Number"),
		form.Placeholder("GB123456789"),
		form.VisibleIf(form.And(
			form.FieldEquals("account_type", "business"),
			form.FieldIn("country", "GB", "DE", "FR"),
		)),
		form.RequiredIf(form.And(
			form.FieldEquals("account_type", "business"),
			form.FieldIn("country", "GB", "DE", "FR"),
		)),
	),

	// Country — always visible, affects VAT field above.
	form.TextInput("country",
		form.Label("Country"),
		form.Default("US"),
		form.Required(),
	),

	// ── Status (admin-set) ──────────────────────────────────────────────────

	// Notes — disabled when account is locked; editors still see the value.
	form.TextAreaInput("notes",
		form.Label("Notes"),
		form.EnabledIf(form.FieldNotEquals("status", "locked")),
	),

	// Status is hidden from the user; set server-side only.
	form.HiddenInput("status",
		form.Default("active"),
	),

	// ── Computed fields ─────────────────────────────────────────────────────

	// full_name is derived automatically; never sent by the client.
	form.TextInput("full_name",
		form.Label("Full Name"),
		form.ReadOnly(),
		form.ComputedAs(func(s *form.FormState) any {
			first := strings.TrimSpace(form.Stringify(s.Get("first_name")))
			last := strings.TrimSpace(form.Stringify(s.Get("last_name")))
			return strings.TrimSpace(first + " " + last)
		}),
	),

	// ── Legal ───────────────────────────────────────────────────────────────

	form.CheckboxInput("terms",
		form.Label("I accept the terms and conditions"),
		form.RequiredIf(form.Always), // always required
	),
)

// HandleForm renders the empty registration form on GET.
func (h *ExampleHandlers) HandleForm(c *fiber.Ctx) error {
	// Build a fresh instance so DefaultValues are pre-populated.
	inst := form.NewInstance(registrationForm)

	// Seed state from the instance defaults.
	state := form.NewFormState("registration", inst)

	// Run computed fields so full_name is populated even on first render.
	form.ApplyComputedFields(registrationForm, state)

	// Project schema + state into a flat []UIField for the renderer.
	// All VisibleIf / EnabledIf / RequiredIf rules are evaluated here.
	fields := form.BuildUIFields(registrationForm, inst, state)

	return render.Component(c, views.ExampleForm(fields))
}

// HandleSubmit processes a POST from the registration form.
func (h *ExampleHandlers) HandleSubmit(c *fiber.Ctx) error {
	// 1. Parse the incoming Datastar/form body into a patch state.
	patch := form.NewFormState("registration", nil)
	if err := c.BodyParser(patch); err != nil {
		return fiber.ErrBadRequest
	}

	// 2. Re-compute derived fields so rules that depend on them are correct.
	form.ApplyComputedFields(registrationForm, patch)

	// 3. Hydrate an instance from the patch and run rule-aware validation.
	//    - Invisible fields are skipped.
	//    - Disabled fields are skipped.
	//    - Dynamic RequiredIf is evaluated against the live state.
	inst := form.NewInstance(registrationForm).Apply(patch)
	form.ValidateWithRules(inst, patch)

	if !inst.Valid() {
		// 4a. Merge validation errors back into state and re-render.
		patch.MergeFromInstance(inst)
		fields := form.BuildUIFields(registrationForm, inst, patch)
		return render.Component(c, views.ExampleForm(fields))
	}

	// 4b. All good — use inst.Values for persistence.
	_ = inst.Values // → pass to your service / repository layer

	return c.Redirect("/registration/success")
}

// HandleFieldChange handles a live Datastar field-change event (partial update).
// Call this from a data-on-input handler to keep the UI reactive without a
// full page reload.
func (h *ExampleHandlers) HandleFieldChange(c *fiber.Ctx) error {
	// Parse only the fields the client sent in this patch.
	patch := form.NewFormState("registration", nil)
	if err := c.BodyParser(patch); err != nil {
		return fiber.ErrBadRequest
	}

	// Re-derive computed fields so downstream visibility rules are current.
	form.ApplyComputedFields(registrationForm, patch)

	// Validate only touched fields (show errors progressively).
	inst := form.NewInstance(registrationForm).Apply(patch)
	form.ValidateWithRules(inst, patch)
	patch.MergeFromInstance(inst)

	fields := form.BuildUIFields(registrationForm, inst, patch)
	return render.Component(c, views.ExampleForm(fields))
}
