package form_test

import (
	"strings"
	"testing"

	"github.com/subhasundardas/gofar/framework/form"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func stateWith(vals map[string]any) *form.FormState {
	fs := form.NewFormState("test", nil)
	for k, v := range vals {
		fs.SetSilent(k, v)
	}
	return fs.Snapshot()
}

// ---------------------------------------------------------------------------
// Logical combinators
// ---------------------------------------------------------------------------

func TestAnd(t *testing.T) {
	s := stateWith(map[string]any{"a": "x", "b": "y"})
	if !form.And(form.FieldNotEmpty("a"), form.FieldNotEmpty("b")).Eval(s) {
		t.Error("And(non-empty, non-empty) should be true")
	}
	if form.And(form.FieldNotEmpty("a"), form.FieldEmpty("b")).Eval(s) {
		t.Error("And(non-empty, empty) should be false")
	}
}

func TestOr(t *testing.T) {
	s := stateWith(map[string]any{"a": "", "b": "y"})
	if !form.Or(form.FieldEmpty("a"), form.FieldNotEmpty("b")).Eval(s) {
		t.Error("Or(empty=true, non-empty=true) should be true")
	}
	if form.Or(form.FieldNotEmpty("a"), form.FieldEmpty("b")).Eval(s) {
		t.Error("Or(non-empty=false, empty=false) should be false")
	}
}

func TestNot(t *testing.T) {
	s := stateWith(map[string]any{"active": "yes"})
	if !form.Not(form.FieldEmpty("active")).Eval(s) {
		t.Error("Not(FieldEmpty) should be true when field is non-empty")
	}
}

// ---------------------------------------------------------------------------
// Field equality
// ---------------------------------------------------------------------------

func TestFieldEquals(t *testing.T) {
	s := stateWith(map[string]any{"role": "admin"})
	if !form.FieldEquals("role", "admin").Eval(s) {
		t.Error("FieldEquals should match string")
	}
	if form.FieldEquals("role", "user").Eval(s) {
		t.Error("FieldEquals should not match wrong value")
	}
}

func TestFieldEqualsLooseNumeric(t *testing.T) {
	// Datastar sends numbers as float64 over JSON but field may be int.
	s := stateWith(map[string]any{"age": float64(42)})
	if !form.FieldEquals("age", 42).Eval(s) {
		t.Error("FieldEquals should match int 42 against float64(42)")
	}
}

func TestFieldIn(t *testing.T) {
	s := stateWith(map[string]any{"country": "US"})
	if !form.FieldIn("country", "US", "CA", "MX").Eval(s) {
		t.Error("FieldIn should match")
	}
	if form.FieldIn("country", "GB", "DE").Eval(s) {
		t.Error("FieldIn should not match")
	}
}

func TestFieldGTAndLT(t *testing.T) {
	s := stateWith(map[string]any{"qty": float64(5)})
	if !form.FieldGT("qty", 4).Eval(s) {
		t.Error("FieldGT should be true")
	}
	if form.FieldGT("qty", 5).Eval(s) {
		t.Error("FieldGT with equal threshold should be false")
	}
	if !form.FieldGTE("qty", 5).Eval(s) {
		t.Error("FieldGTE with equal threshold should be true")
	}
	if !form.FieldLT("qty", 6).Eval(s) {
		t.Error("FieldLT should be true")
	}
}

func TestFieldContains(t *testing.T) {
	s := stateWith(map[string]any{"email": "alice@example.com"})
	if !form.FieldContains("email", "@example").Eval(s) {
		t.Error("FieldContains should match")
	}
	if form.FieldContains("email", "@other").Eval(s) {
		t.Error("FieldContains should not match")
	}
}

func TestFieldTrue(t *testing.T) {
	s := stateWith(map[string]any{"agreed": "true"})
	if !form.FieldTrue("agreed").Eval(s) {
		t.Error("FieldTrue should be true for string 'true'")
	}
	s2 := stateWith(map[string]any{"agreed": "false"})
	if form.FieldTrue("agreed").Eval(s2) {
		t.Error("FieldTrue should be false for string 'false'")
	}
}

// ---------------------------------------------------------------------------
// IsVisible / IsEnabled / IsRequired
// ---------------------------------------------------------------------------

func TestIsVisibleNoRule(t *testing.T) {
	f := form.TextInput("name")
	s := stateWith(nil)
	if !form.IsVisible(f, s) {
		t.Error("no VisibleIf rule → always visible")
	}
}

func TestIsVisibleWithRule(t *testing.T) {
	f := form.TextInput("company",
		form.VisibleIf(form.FieldEquals("type", "business")),
	)
	s1 := stateWith(map[string]any{"type": "personal"})
	if form.IsVisible(f, s1) {
		t.Error("company should be hidden when type=personal")
	}
	s2 := stateWith(map[string]any{"type": "business"})
	if !form.IsVisible(f, s2) {
		t.Error("company should be visible when type=business")
	}
}

func TestIsEnabledWithRule(t *testing.T) {
	f := form.TextInput("notes",
		form.EnabledIf(form.FieldNotEquals("status", "locked")),
	)
	s1 := stateWith(map[string]any{"status": "locked"})
	if form.IsEnabled(f, s1) {
		t.Error("notes should be disabled when status=locked")
	}
	s2 := stateWith(map[string]any{"status": "active"})
	if !form.IsEnabled(f, s2) {
		t.Error("notes should be enabled when status=active")
	}
}

func TestIsRequiredWithRule(t *testing.T) {
	f := form.TextInput("vatNumber",
		form.RequiredIf(form.FieldEquals("type", "business")),
	)
	s1 := stateWith(map[string]any{"type": "personal"})
	if form.IsRequired(f, s1) {
		t.Error("vatNumber should not be required when type=personal")
	}
	s2 := stateWith(map[string]any{"type": "business"})
	if !form.IsRequired(f, s2) {
		t.Error("vatNumber should be required when type=business")
	}
}

// ---------------------------------------------------------------------------
// Computed fields
// ---------------------------------------------------------------------------

func TestApplyComputedFields(t *testing.T) {
	f := form.New("test").Fields(
		form.TextInput("firstName", form.Default("Alice")),
		form.TextInput("lastName", form.Default("Smith")),
		form.TextInput("fullName",
			form.ComputedAs(func(s *form.FormState) any {
				first := form.Stringify(s.Get("firstName"))
				last := form.Stringify(s.Get("lastName"))
				return strings.TrimSpace(first + " " + last)
			}),
		),
	)

	state := form.NewFormState("test", f.NewInstance())
	form.ApplyComputedFields(f, state)

	got := form.Stringify(state.Get("fullName"))
	if got != "Alice Smith" {
		t.Errorf("expected 'Alice Smith', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// ValidateWithRules — rule-aware validation
// ---------------------------------------------------------------------------

func TestValidateWithRules_SkipsInvisibleFields(t *testing.T) {
	f := form.New("order").Fields(
		form.TextInput("type", form.Required()),
		form.TextInput("vatNumber",
			form.RequiredIf(form.FieldEquals("type", "business")),
			form.VisibleIf(form.FieldEquals("type", "business")),
		),
	)

	state := form.NewFormState("order", nil).
		Set("type", "personal").
		Set("vatNumber", "")

	inst := form.NewInstance(f).Apply(state)
	form.ValidateWithRules(inst, state)

	if inst.HasError("vatNumber") {
		t.Error("vatNumber should not have errors when invisible")
	}
}

func TestValidateWithRules_DynamicRequired(t *testing.T) {
	f := form.New("order").Fields(
		form.TextInput("type"),
		form.TextInput("vatNumber",
			form.RequiredIf(form.FieldEquals("type", "business")),
		),
	)

	state := form.NewFormState("order", nil).
		Set("type", "business").
		Set("vatNumber", "")

	inst := form.NewInstance(f).Apply(state)
	form.ValidateWithRules(inst, state)

	if !inst.HasError("vatNumber") {
		t.Error("vatNumber should be required when type=business")
	}
}

func TestValidateWithRules_SkipsDisabledFields(t *testing.T) {
	f := form.New("profile").Fields(
		form.TextInput("notes",
			form.EnabledIf(form.FieldNotEquals("status", "locked")),
			form.Required(),
		),
	)

	state := form.NewFormState("profile", nil).
		Set("status", "locked").
		Set("notes", "")

	inst := form.NewInstance(f).Apply(state)
	form.ValidateWithRules(inst, state)

	if inst.HasError("notes") {
		t.Error("notes should not be validated when disabled")
	}
}

// ---------------------------------------------------------------------------
// BuildUIFields with state
// ---------------------------------------------------------------------------

func TestBuildUIFields_WithState(t *testing.T) {
	f := form.New("signup").Fields(
		form.TextInput("type"),
		form.TextInput("company",
			form.Label("Company"),
			form.VisibleIf(form.FieldEquals("type", "business")),
			form.RequiredIf(form.FieldEquals("type", "business")),
		),
	)

	// Personal → company hidden, not required
	state1 := form.NewFormState("signup", nil).Set("type", "personal")
	fields1 := form.BuildUIFields(f, nil, state1)
	m1 := form.BuildUIFieldMap(fields1)

	if m1["company"].Visible {
		t.Error("company should not be visible for personal")
	}
	if m1["company"].Required {
		t.Error("company should not be required for personal")
	}

	// Business → company visible + required
	state2 := form.NewFormState("signup", nil).Set("type", "business")
	fields2 := form.BuildUIFields(f, nil, state2)
	m2 := form.BuildUIFieldMap(fields2)

	if !m2["company"].Visible {
		t.Error("company should be visible for business")
	}
	if !m2["company"].Required {
		t.Error("company should be required for business")
	}
}
