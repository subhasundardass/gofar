// Package handler contains HTTP handlers for the Example module.
// Handlers are thin: parse request → call service → return response.
package handler

import (
	"log"

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

// --Form private
func (h *ExampleHandlers) HandleForm(c *fiber.Ctx) error {

	fm := form.New("example").Fields(
		form.TextField{BaseField: form.BaseField{
			Key:         "first_name",
			Type:        "text",
			Label:       "First Name",
			Placeholder: "Jane",
			Required:    true,
			Validators: []form.Validator{
				form.RequiredValidator{Field: "first_name"},
				form.BetweenLengthValidator{Field: "first_name", Min: 1, Max: 50},
			},
		}},
		form.TextField{BaseField: form.BaseField{
			Key:         "last_name",
			Type:        "text",
			Label:       "Last Name",
			Placeholder: "Doe",
			Required:    true,
			Validators: []form.Validator{
				form.RequiredValidator{Field: "last_name"},
				form.BetweenLengthValidator{Field: "last_name", Min: 1, Max: 50},
			},
		}},
		form.NumberField{BaseField: form.BaseField{
			Key:          "age",
			Type:         "number",
			Label:        "Age",
			Placeholder:  "30",
			HelpText:     "Must be 18 or older",
			DefaultValue: 25,
			Validators: []form.Validator{
				form.BetweenValueValidator{Field: "age", Min: 18, Max: 120},
			},
		}},
	)

	inst := form.NewInstance(fm)
	fields := form.BuildUIFields(fm, inst)

	return render.Component(c, views.ExampleForm(fields))
}
