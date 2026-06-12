package form

import "fmt"

// FieldType is a typed alias for the string values stored on BaseField.Type.
// Using a typed alias (not a `type FieldType string` distinct type) keeps
// the existing BaseField.Type field backward-compatible while giving the
// renderer and dispatcher a single source of truth and a set of canonical
// constants to switch on.
type FieldType = string

// Canonical field types. The constants match the strings produced by the
// TextInput / NumberInput / … helpers in inputs.go. Use these names in
// switch statements and in tests so a typo becomes a compile error.
const (
	String   FieldType = "text"
	Email    FieldType = "email"
	Password FieldType = "password"
	Number   FieldType = "number"
	Bool     FieldType = "checkbox" // boolean — single checkbox
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
	Color    FieldType = "color"
	Range    FieldType = "range"
	Hidden   FieldType = "hidden"
	Lookup   FieldType = "lookup" // search-as-you-type combobox
)

// Option is a single (value, label) pair rendered by SelectField / LookupField.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FieldMeta is the presentation-level metadata about a field that a
// renderer needs but that should not live on the per-request *Instance.
// Pulling it out keeps the form schema stable and shareable across many
// requests while still letting the renderer know "this is a required
// email field" or "this select has these options".
type FieldMeta struct {
	Type     FieldType
	Label    string
	Required bool
	Options  []Option
	// Help is widget-level guidance (e.g. "min 8 chars"). Exposed to
	// renderers as helper text under the input.
	Help string
}

// UIField is the per-request view model a renderer (templ, Datastar, JSON
// schema) consumes. It bundles together the field's identity, its current
// value, any validation error, and the presentation hints from FieldMeta.
//
// A UIField is intentionally a value type: renderers usually loop over a
// []UIField and the cost of a copy is negligible. The convenience builders
// below (BuildUIFields, NewUIField) make it cheap to produce a slice from
// an *Instance.
type UIField struct {
	Key      string
	Value    string
	RawValue any
	Error    string
	Writable bool // false → field is read-only or hidden
	Visible  bool
	Meta     FieldMeta
}

// NewUIField constructs a UIField from the pieces you already have. Use
// this from handler code that builds a custom field on the fly.
func NewUIField(key, value, err string, meta FieldMeta) UIField {
	return UIField{
		Key:      key,
		Value:    value,
		RawValue: value,
		Error:    err,
		Writable: true,
		Visible:  true,
		Meta:     meta,
	}
}

// ToString coerces the field's current value to a string, regardless of
// whether it was set as int / bool / nil / etc. Useful inside templates.
func (f UIField) ToString() string {
	if f.Value != "" {
		return f.Value
	}
	return Stringify(f.RawValue)
}

// Stringify is the package-wide stringification helper. It treats nil as
// "", bool as "true"/"false", and everything else via fmt.Sprint.
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

// BuildUIFields projects a *Form + *Instance pair into a []UIField suitable
// for a renderer. The order matches the order fields were registered on
// the Form. Hidden fields are emitted with Visible=false so the dispatcher
// can decide what to do with them; absent values fall back to the field's
// DefaultValue so the very first render is populated.
func BuildUIFields(f *Form, inst *Instance) []UIField {
	if f == nil {
		return nil
	}
	fields := f.GetFields()
	out := make([]UIField, 0, len(fields))
	for _, field := range fields {
		meta := FieldMeta{
			Type:     FieldType(field.GetType()),
			Label:    field.GetLabel(),
			Required: field.GetRequired(),
			Help:     field.GetHelpText(),
		}
		uf := UIField{
			Key:      field.GetKey(),
			Writable: !field.GetReadOnly() && !field.GetDisabled(),
			Visible:  true,
			Meta:     meta,
		}
		// Pull value from instance, falling back to the field's default.
		var raw any
		if inst != nil {
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
