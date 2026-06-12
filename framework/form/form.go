package form

// Form is the static, declarative description of a UI form. It holds a
// stable schema (name + ordered fields) that is safe to register globally
// and reuse across many requests. Per-request state — values, errors,
// touched/dirty flags — lives on an *Instance, not on the Form itself.
//
// Lifecycle:
//
//	form := form.New("signup").Fields(...)        // built once at boot
//	inst := form.NewInstance()                     // created per request
//	inst.Set("email", "alice@example.com")
//	errs := inst.Validate()                        // run validators
//	//  -> errs is *form.Errors (form.Errors map[string][]string)
type Form struct {
	Name   string
	fields []Field
}

// New constructs an empty form. Add fields with .Fields(...).
func New(name string) *Form {
	return &Form{Name: name}
}

// Fields appends one or more fields to the form. Returns the receiver so
// calls can be chained. Duplicate keys are allowed; Validate() will run
// every validator on each occurrence.
func (f *Form) Fields(fields ...Field) *Form {
	f.fields = append(f.fields, fields...)
	return f
}

// GetFields returns a snapshot of the registered fields. The returned slice
// is a copy — mutating it does not affect the form.
func (f *Form) GetFields() []Field {
	out := make([]Field, len(f.fields))
	copy(out, f.fields)
	return out
}

// NewInstance returns a fresh per-request instance backed by this form.
// All field DefaultValues are copied into the instance so the very first
// render is already populated.
func (f *Form) NewInstance() *Instance {
	inst := &Instance{
		Form:   f,
		Values: make(map[string]any, len(f.fields)),
		Errors: Errors{},
	}
	for _, field := range f.fields {
		inst.Values[field.GetKey()] = field.GetDefaultValue()
	}
	return inst
}

// FieldByKey returns the first field with the given key, or nil. Useful
// for renderer code that needs field-specific metadata (e.g. the templ
// renderer switching on GetType() to choose the right input element).
func (f *Form) FieldByKey(key string) Field {
	for _, field := range f.fields {
		if field.GetKey() == key {
			return field
		}
	}
	return nil
}

// Keys returns the ordered list of field keys. Stable across calls.
func (f *Form) Keys() []string {
	out := make([]string, len(f.fields))
	for i, field := range f.fields {
		out[i] = field.GetKey()
	}
	return out
}
