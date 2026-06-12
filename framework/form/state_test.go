package form_test

import (
	"encoding/json"
	"testing"

	"github.com/subhasundardas/gofar/framework/form"
)

// TestFormStateRoundTrip verifies that a FormState can be marshalled to
// JSON and unmarshalled back into a value-equal FormState. This is the
// property the Datastar integration depends on.
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

	var got form.FormState
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	checks := []struct {
		field string
		want  any
	}{
		{"name", "Alice"},
		{"email", "alice@example.com"},
		// JSON numbers round-trip as float64, not the original int/string.
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
	if got.Meta["submittedAt"] != "2026-01-15T10:00:00Z" {
		t.Errorf("meta mismatch: %v", got.Meta)
	}
}

// TestFormStateMerge verifies that Merge() applies a patch produced by the
// client (Datastar) to a server-side FormState, and that nil values
// delete keys — matching Datastar's "null = delete" convention.
func TestFormStateMerge(t *testing.T) {
	server := form.NewFormState("signup", nil).
		Set("name", "Alice").
		Set("email", "alice@example.com")

	patch := form.NewFormState("signup", nil)
	patch.Values["email"] = "new@example.com"
	patch.Values["age"] = 30
	patch.Values["stale"] = nil // should be deleted on merge
	patch.Touched["email"] = true

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
	if _, ok := server.Values["stale"]; ok {
		t.Error("nil value should have been deleted")
	}
}

// TestFormStateApplyToInstance verifies the Instance.Apply(FormState)
// helper, which is what a Datastar POST handler uses to hydrate a fresh
// instance from a client patch before running validation.
//
// We attach EmailValidator explicitly because EmailInput currently only
// sets the HTML type — adding type-specific validators is tracked as a
// follow-up enhancement to the form package.
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

// TestFormStateNilDelete verifies that Set with nil effectively removes
// the key from subsequent reads.
func TestFormStateNilDelete(t *testing.T) {
	fs := form.NewFormState("signup", nil).Set("name", "Alice")
	fs.Set("name", nil)
	if v, ok := fs.Values["name"]; ok && v != nil {
		t.Errorf("name should be removed or nil, got %v", v)
	}
}

// TestInstanceValidateRequiresEmptyRequiredField makes sure the
// RequiredValidator's Field gets populated from the field's Key, fixing
// the long-standing bug where error messages lost the field name.
func TestInstanceValidateRequiresEmptyRequiredField(t *testing.T) {
	signup := form.New("signup").Fields(
		form.TextInput("email", form.Required()),
	)

	inst := form.NewInstance(signup) // email stays nil
	inst.Validate()

	if !inst.HasError("email") {
		t.Fatal("email should be required and produce an error")
	}
	if inst.FirstError("email") == "" {
		t.Error("error message should not be empty")
	}
}
