package form

import "fmt"

// FieldType is a typed alias for the string values stored on BaseField.Type.
type FieldType = string

// Canonical field types.
const (
	String   FieldType = "text"
	Email    FieldType = "email"
	Password FieldType = "password"
	Number   FieldType = "number"
	Bool     FieldType = "checkbox"
	TextArea FieldType = "textarea"
	Select   FieldType = "select"
	Radio    FieldType = "radio"
	Date     FieldType = "date"
	Time     FieldType = "time"
	DateTime FieldType = "datetime"
	File     FieldType = "file"
	Image    FieldType = "image"
	URL      FieldType = "url"
	Phone    FieldType = "phone"
	Money    FieldType = "money"
	Color    FieldType = "color"
	Range    FieldType = "range"
	Hidden   FieldType = "hidden"
	Lookup   FieldType = "lookup"
)

// Option is a single (value, label) pair rendered by SelectField / LookupField.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FieldMeta is the presentation-level metadata about a field that a renderer
// needs but should not live on the per-request *Instance.
type FieldMeta struct {
	Type        FieldType
	Label       string
	Placeholder string
	Required    bool
	Options     []Option
	Help        string
}

// UIField is the per-request view model a renderer (templ, Datastar, JSON
// schema) consumes. It bundles the field's identity, its current value,
// any validation error, presentation hints, and the evaluated rule state.
//
// The Visible/Enabled/Required fields here reflect the live rule evaluations
// for the current FormState — renderers should read from UIField, not call
// IsVisible/IsEnabled/IsRequired themselves.
type UIField struct {
	Key      string
	Value    string
	RawValue any
	Error    string

	// Rule-evaluated state — always populated by BuildUIFields.
	Visible  bool // false → hide the field entirely
	Enabled  bool // false → render as disabled
	Required bool // true  → show required indicator, validate presence
	Computed bool // true  → field value is derived, not editable

	// Writable is false when Computed or !Enabled or ReadOnly.
	Writable bool

	Meta FieldMeta
}

// NewUIField constructs a UIField from raw pieces. Used by handler code
// that builds a custom field on the fly without a Form schema.
func NewUIField(key, value, err string, meta FieldMeta) UIField {
	return UIField{
		Key:      key,
		Value:    value,
		RawValue: value,
		Error:    err,
		Visible:  true,
		Enabled:  true,
		Required: meta.Required,
		Computed: false,
		Writable: true,
		Meta:     meta,
	}
}

// ToString coerces the field's current value to a string. Useful inside templates.
func (f UIField) ToString() string {
	if f.Value != "" {
		return f.Value
	}
	return Stringify(f.RawValue)
}

// Stringify is the package-wide stringification helper.
func Stringify(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(v)
	}
}

// BuildUIFields projects a *Form + *Instance + *FormState into a []UIField
// suitable for a renderer. All rules (VisibleIf, EnabledIf, RequiredIf,
// ComputedAs) are evaluated against the live state.
//
// Pass state=nil to skip rule evaluation (all fields visible, enabled,
// using static Required). This keeps callers that don't use the reactive
// engine working without changes.
func BuildUIFields(f *Form, inst *Instance, state *FormState) []UIField {
	if f == nil {
		return nil
	}

	// Build an eval state pointer for rule evaluation. We never pass
	// FormState by value — it contains sync.RWMutex.
	var evalState *FormState
	if state != nil {
		evalState = state.Snapshot()
	} else if inst != nil {
		// Synthesise a minimal *FormState from instance values.
		evalState = &FormState{
			Values:  cloneAnyMap(inst.Values),
			Touched: map[string]bool{},
			Errors:  Errors{},
		}
	} else {
		evalState = &FormState{
			Values:  map[string]any{},
			Touched: map[string]bool{},
			Errors:  Errors{},
		}
	}

	fields := f.GetFields()
	out := make([]UIField, 0, len(fields))

	for _, field := range fields {
		eval := EvalField(field, evalState)

		meta := FieldMeta{
			Type:        FieldType(field.GetType()),
			Label:       field.GetLabel(),
			Placeholder: field.GetPlaceholder(),
			Required:    eval.Required,
			Help:        field.GetHelpText(),
		}

		writable := eval.Enabled && !eval.Computed && !field.GetReadOnly()

		uf := UIField{
			Key:      field.GetKey(),
			Visible:  eval.Visible,
			Enabled:  eval.Enabled,
			Required: eval.Required,
			Computed: eval.Computed,
			Writable: writable,
			Meta:     meta,
		}

		// Pull value from instance or state, falling back to default.
		var raw any
		if state != nil {
			raw = state.Get(field.GetKey())
		}
		if raw == nil && inst != nil {
			raw = inst.Get(field.GetKey())
		}
		if raw == nil {
			raw = field.GetDefaultValue()
		}
		uf.RawValue = raw
		uf.Value = Stringify(raw)

		// Pull the first error message, if any.
		if inst != nil {
			uf.Error = inst.FirstError(field.GetKey())
		}

		out = append(out, uf)
	}
	return out
}

// BuildUIFieldMap returns the BuildUIFields result indexed by Key.
func BuildUIFieldMap(fields []UIField) map[string]UIField {
	result := make(map[string]UIField, len(fields))
	for _, field := range fields {
		result[field.Key] = field
	}
	return result
}
