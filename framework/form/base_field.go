package form

// BaseField carries the shared, presentation- and behaviour-level metadata
// every form field needs. Concrete field types (TextField, NumberField, …)
// embed BaseField and inherit all getters below, which means they satisfy
// the form.Field interface for free.
type BaseField struct {
	Key          string
	Type         string
	Label        string
	Placeholder  string
	HelpText     string
	DefaultValue any

	Required bool
	ReadOnly bool
	Disabled bool

	// ── Reactive Rule Engine ─────────────────────────────────────────────
	//
	// Rules are evaluated lazily against the live FormState each time the
	// form engine renders or validates. They compose freely:
	//
	//   VisibleIf(And(FieldEquals("role", "admin"), FieldNotEmpty("org")))
	//   RequiredIf(FieldEquals("country", "US"))
	//   EnabledIf(Not(FieldTrue("locked")))
	//   ComputedAs(func(s form.FormState) any { return s.Get("a").(string) + s.Get("b").(string) })

	// VisibleIf controls conditional visibility. When the Rule evaluates to
	// false the field is hidden from the rendered form.
	VisibleIf Rule

	// EnabledIf controls whether the field is interactive. When false the
	// field renders as disabled and is excluded from validation.
	EnabledIf Rule

	// RequiredIf makes required-status dynamic. When true the field is
	// treated as required even if the static Required flag is false. A
	// RequiredValidator is injected automatically during Validate().
	RequiredIf Rule

	// ComputeWith, if set, runs before any validation or rule evaluation.
	// Its return value is written back into the FormState so downstream
	// rules and validators see the derived value. Fields with ComputeWith
	// are normally ReadOnly so the user cannot override the computed result.
	ComputeWith ComputeFunc

	Validators []Validator
}

// ── Identity ──────────────────────────────────────────────────────────────────

func (f BaseField) GetKey() string  { return f.Key }
func (f BaseField) GetType() string { return f.Type }

// ── Presentation ──────────────────────────────────────────────────────────────

func (f BaseField) GetLabel() string       { return f.Label }
func (f BaseField) GetPlaceholder() string { return f.Placeholder }
func (f BaseField) GetHelpText() string    { return f.HelpText }
func (f BaseField) GetDefaultValue() any   { return f.DefaultValue }

// ── Behaviour ─────────────────────────────────────────────────────────────────

func (f BaseField) GetRequired() bool { return f.Required }
func (f BaseField) GetReadOnly() bool { return f.ReadOnly }
func (f BaseField) GetDisabled() bool { return f.Disabled }

// ── Rule accessors (used by the runtime) ─────────────────────────────────────

func (f BaseField) GetVisibleIf() Rule    { return f.VisibleIf }
func (f BaseField) GetEnabledIf() Rule    { return f.EnabledIf }
func (f BaseField) GetRequiredIf() Rule   { return f.RequiredIf }
func (f BaseField) GetComputeWith() ComputeFunc { return f.ComputeWith }

// ── Validation ────────────────────────────────────────────────────────────────

func (f BaseField) GetValidators() []Validator { return f.Validators }
