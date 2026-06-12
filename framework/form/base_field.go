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

// ── Validation ────────────────────────────────────────────────────────────────

func (f BaseField) GetValidators() []Validator { return f.Validators }
