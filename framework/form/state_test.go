package form_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/subhasundardas/gofar/framework/form"
)

// =============================================================================
// FormState — serialisation
// =============================================================================

// TestFormStateRoundTrip verifies that a FormState can be marshalled to JSON
// and unmarshalled back into a value-equal FormState. This is the property
// the Datastar integration depends on.
func TestFormStateRoundTrip(t *testing.T) {
	src := form.NewFormState("signup", nil).
		Set("name", "Alice").
		Set("email", "alice@example.com").
		Set("age", 30).
		SetSilent("role", "admin").
		SetError("name", "must not be blank").
		AddError("email", "domain blocked").
		SetMeta("submittedAt", "2026-01-15T10:00:00Z")

	data, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	got := new(form.FormState)
	if err := json.Unmarshal(data, got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	checks := []struct {
		field string
		want  any
	}{
		{"name", "Alice"},
		{"email", "alice@example.com"},
		// JSON numbers always round-trip as float64.
		{"age", float64(30)},
		{"role", "admin"},
	}
	for _, c := range checks {
		if got.Get(c.field) != c.want {
			t.Errorf("value mismatch for %q: got %v (%T), want %v (%T)",
				c.field, got.Get(c.field), got.Get(c.field), c.want, c.want)
		}
	}

	if !got.IsTouched("name") {
		t.Error("name should be marked touched after Set")
	}
	if got.IsTouched("role") {
		t.Error("role should NOT be touched after SetSilent")
	}
	if !got.HasError("name") {
		t.Error("name should have an error")
	}
	if got.FirstError("name") != "must not be blank" {
		t.Errorf("name error mismatch: %q", got.FirstError("name"))
	}
	if got.FirstError("email") != "domain blocked" {
		t.Errorf("email error mismatch: %q", got.FirstError("email"))
	}
	if got.Name != "signup" {
		t.Errorf("name mismatch: %q", got.Name)
	}
	if got.GetMeta("submittedAt") != "2026-01-15T10:00:00Z" {
		t.Errorf("meta mismatch: %v", got.GetMeta("submittedAt"))
	}
}

// =============================================================================
// FormState — mutation and merge
// =============================================================================

// TestFormStateMerge verifies that Merge() applies a client patch to a
// server-side FormState and that nil values delete keys — matching
// Datastar's "null = delete" convention.
// Uses only the safe API (Set/SetSilent/MarkTouched) — never direct map
// writes — to avoid racing with the embedded sync.RWMutex.
func TestFormStateMerge(t *testing.T) {
	server := form.NewFormState("signup", nil).
		Set("name", "Alice").
		Set("email", "alice@example.com")

	patch := form.NewFormState("signup", nil).
		Set("email", "new@example.com").
		Set("age", 30).
		Set("stale", nil) // nil → delete on merge

	server.Merge(patch)

	if server.Get("email") != "new@example.com" {
		t.Errorf("email not updated: %v", server.Get("email"))
	}
	if server.Get("age") != 30 {
		t.Errorf("age not added: %v", server.Get("age"))
	}
	if !server.IsTouched("email") {
		t.Error("email should be marked touched after merge")
	}
	if server.Get("stale") != nil {
		t.Error("nil-patched key should be absent/nil after merge")
	}
}

// TestFormStateSnapshot verifies that Snapshot() deep-copies the state so
// mutations to the original do not affect the copy and vice-versa.
func TestFormStateSnapshot(t *testing.T) {
	original := form.NewFormState("test", nil).Set("x", "1")
	snap := original.Snapshot()

	original.Set("x", "2")

	if snap.Get("x") != "1" {
		t.Errorf("snapshot should not reflect mutation: got %v", snap.Get("x"))
	}
	if original.Get("x") != "2" {
		t.Errorf("original should reflect mutation: got %v", original.Get("x"))
	}
}

// TestFormStateNilDelete verifies that Set(key, nil) removes the key from
// subsequent reads via the safe API.
func TestFormStateNilDelete(t *testing.T) {
	fs := form.NewFormState("signup", nil).Set("name", "Alice")
	fs.Set("name", nil)
	if fs.Get("name") != nil {
		t.Errorf("name should be nil after Set(nil), got %v", fs.Get("name"))
	}
}

// TestFormStateResetTouched verifies that ResetTouched clears all flags.
func TestFormStateResetTouched(t *testing.T) {
	fs := form.NewFormState("signup", nil).
		Set("a", "1").
		Set("b", "2")

	fs.ResetTouched()

	if fs.IsTouched("a") || fs.IsTouched("b") {
		t.Error("all touched flags should be cleared after ResetTouched")
	}
}

// TestFormStateClearErrors verifies that ClearErrors wipes the error map.
func TestFormStateClearErrors(t *testing.T) {
	fs := form.NewFormState("signup", nil).
		SetError("email", "invalid").
		AddError("name", "required")

	fs.ClearErrors()

	if fs.HasError("email") || fs.HasError("name") {
		t.Error("all errors should be cleared after ClearErrors")
	}
}

// =============================================================================
// Instance — Apply and Validate
// =============================================================================

// TestFormStateApplyToInstance verifies Instance.Apply(*FormState) which is
// what a Datastar POST handler uses to hydrate a fresh instance from a client
// patch before running validation.
func TestFormStateApplyToInstance(t *testing.T) {
	signup := form.New("signup").Fields(
		form.TextInput("name",
			form.Required(),
			form.WithValidator(form.MinLengthValidator{Min: 2}),
		),
		form.EmailInput("email",
			form.Required(),
			form.WithValidator(form.EmailValidator{}),
		),
	)

	state := form.NewFormState("signup", nil).
		Set("name", "Alice").
		Set("email", "not-an-email")

	inst := form.NewInstance(signup).Apply(state)
	inst.Validate()

	if inst.Valid() {
		t.Error("instance should be invalid (bad email)")
	}
	if !inst.HasError("email") {
		t.Error("email should have a validation error")
	}
	if inst.FirstError("name") != "" {
		t.Errorf("name should be valid, got error %q", inst.FirstError("name"))
	}
}

// TestInstanceValidateRequiresEmptyRequiredField ensures the RequiredValidator
// names the field in its error message.
func TestInstanceValidateRequiresEmptyRequiredField(t *testing.T) {
	signup := form.New("signup").Fields(
		form.TextInput("email", form.Required()),
	)

	inst := form.NewInstance(signup) // email stays nil/default
	inst.Validate()

	if !inst.HasError("email") {
		t.Fatal("email should be required and produce an error")
	}
	if inst.FirstError("email") == "" {
		t.Error("error message should not be empty")
	}
}

// =============================================================================
// Rule engine — ValidateWithRules
// =============================================================================

// TestValidateWithRules_SkipsInvisible verifies that hidden fields are not
// validated even when their RequiredIf rule would fire.
func TestValidateWithRules_SkipsInvisible(t *testing.T) {
	f := form.New("order").Fields(
		form.TextInput("type", form.Required()),
		form.TextInput("vatNumber",
			form.VisibleIf(form.FieldEquals("type", "business")),
			form.RequiredIf(form.FieldEquals("type", "business")),
		),
	)

	state := form.NewFormState("order", nil).
		Set("type", "personal").
		Set("vatNumber", "")

	inst := form.NewInstance(f).Apply(state)
	form.ValidateWithRules(inst, state)

	if inst.HasError("vatNumber") {
		t.Error("vatNumber should not error when hidden")
	}
}

// TestValidateWithRules_DynamicRequired verifies that RequiredIf injects
// validation when its rule fires.
// func TestValidateWithRules_DynamicRequired(t *testing.T) {
// 	f := form.New("order").Fields(
// 		form.TextInput("type"),
// 		form.TextInput("vatNumber",
// 			form.RequiredIf(form.FieldEquals("type", "business")),
// 		),
// 	)

// 	state := form.NewFormState("order", nil).
// 		Set("type", "business").
// 		Set("vatNumber", "")

// 	inst := form.NewInstance(f).Apply(state)
// 	form.ValidateWithRules(inst, state)

// 	if !inst.HasError("vatNumber") {
// 		t.Error("vatNumber should be required when type=business")
// 	}
// }

// TestValidateWithRules_SkipsDisabled verifies that disabled fields are
// excluded from validation entirely.
func TestValidateWithRules_SkipsDisabled(t *testing.T) {
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

// TestValidateWithRules_NoDuplicateRequiredError ensures that when both a
// static Required validator and a RequiredIf rule are active on the same
// field, only one "required" error is produced.
func TestValidateWithRules_NoDuplicateRequiredError(t *testing.T) {
	f := form.New("x").Fields(
		form.TextInput("field",
			form.Required(),
			form.RequiredIf(form.Always),
		),
	)

	state := form.NewFormState("x", nil).Set("field", "")
	inst := form.NewInstance(f).Apply(state)
	form.ValidateWithRules(inst, state)

	errs := inst.Errors["field"]
	if len(errs) != 1 {
		t.Errorf("expected exactly 1 required error, got %d: %v", len(errs), errs)
	}
}

// =============================================================================
// Rule engine — ApplyComputedFields
// =============================================================================

// TestApplyComputedFields_Basic verifies that a ComputedAs function runs and
// its result is written back to the state before any rules fire.
func TestApplyComputedFields_Basic(t *testing.T) {
	f := form.New("profile").Fields(
		form.TextInput("firstName", form.Default("Jane")),
		form.TextInput("lastName", form.Default("Doe")),
		form.TextInput("fullName",
			form.ComputedAs(func(s *form.FormState) any {
				first := form.Stringify(s.Get("firstName"))
				last := form.Stringify(s.Get("lastName"))
				return strings.TrimSpace(first + " " + last)
			}),
		),
	)

	state := form.NewFormState("profile", f.NewInstance())
	form.ApplyComputedFields(f, state)

	if got := form.Stringify(state.Get("fullName")); got != "Jane Doe" {
		t.Errorf("expected 'Jane Doe', got %q", got)
	}
}

// TestApplyComputedFields_SeesUpdatedValues verifies that the compute fn
// sees values that were set after instance creation (i.e. user input).
func TestApplyComputedFields_SeesUpdatedValues(t *testing.T) {
	f := form.New("calc").Fields(
		form.NumberInput("qty"),
		form.NumberInput("price"),
		form.TextInput("total",
			form.ComputedAs(func(s *form.FormState) any {
				qty := s.Get("qty")
				price := s.Get("price")
				q, _ := qty.(float64)
				p, _ := price.(float64)
				return q * p
			}),
		),
	)

	state := form.NewFormState("calc", nil).
		Set("qty", float64(3)).
		Set("price", float64(9.99))

	form.ApplyComputedFields(f, state)

	got, _ := state.Get("total").(float64)
	if got != 3*9.99 {
		t.Errorf("expected total=29.97, got %v", got)
	}
}

// =============================================================================
// Rule engine — BuildUIFields
// =============================================================================

// TestBuildUIFields_RulesEvaluated verifies that BuildUIFields reflects live
// VisibleIf / RequiredIf / EnabledIf state in the returned UIField slice.
func TestBuildUIFields_RulesEvaluated(t *testing.T) {
	f := form.New("checkout").Fields(
		form.TextInput("type"),
		form.TextInput("company",
			form.VisibleIf(form.FieldEquals("type", "business")),
			form.RequiredIf(form.FieldEquals("type", "business")),
		),
		form.TextInput("notes",
			form.EnabledIf(form.FieldNotEquals("status", "locked")),
		),
	)

	// personal + locked
	s1 := form.NewFormState("checkout", nil).
		Set("type", "personal").
		Set("status", "locked")
	m1 := form.BuildUIFieldMap(form.BuildUIFields(f, nil, s1))

	if m1["company"].Visible {
		t.Error("company should be hidden when type=personal")
	}
	if m1["company"].Required {
		t.Error("company should not be required when type=personal")
	}
	if m1["notes"].Enabled {
		t.Error("notes should be disabled when status=locked")
	}

	// business + unlocked
	s2 := form.NewFormState("checkout", nil).
		Set("type", "business").
		Set("status", "active")
	m2 := form.BuildUIFieldMap(form.BuildUIFields(f, nil, s2))

	if !m2["company"].Visible {
		t.Error("company should be visible when type=business")
	}
	if !m2["company"].Required {
		t.Error("company should be required when type=business")
	}
	if !m2["notes"].Enabled {
		t.Error("notes should be enabled when status=active")
	}
}

// TestBuildUIFields_NilState verifies backward-compatibility: passing nil for
// state still returns all fields with their static Required/Disabled flags.
func TestBuildUIFields_NilState(t *testing.T) {
	f := form.New("simple").Fields(
		form.TextInput("name", form.Required()),
		form.TextInput("bio", form.Disabled()),
	)

	fields := form.BuildUIFields(f, nil, nil)
	m := form.BuildUIFieldMap(fields)

	if !m["name"].Required {
		t.Error("name should be statically required")
	}
	if m["bio"].Enabled {
		t.Error("bio should be disabled")
	}
}

// TestBuildUIFields_ComputedMarked verifies that fields with ComputedAs are
// marked Computed=true and Writable=false in the UIField.
func TestBuildUIFields_ComputedMarked(t *testing.T) {
	f := form.New("c").Fields(
		form.TextInput("a", form.Default("hello")),
		form.TextInput("upper",
			form.ComputedAs(func(s *form.FormState) any {
				return strings.ToUpper(form.Stringify(s.Get("a")))
			}),
		),
	)

	state := form.NewFormState("c", f.NewInstance())
	form.ApplyComputedFields(f, state)
	m := form.BuildUIFieldMap(form.BuildUIFields(f, nil, state))

	if !m["upper"].Computed {
		t.Error("upper should be marked Computed")
	}
	if m["upper"].Writable {
		t.Error("computed field should not be Writable")
	}
	if m["upper"].Value != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", m["upper"].Value)
	}
}
