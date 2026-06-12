package formui

// FieldProps carries presentation-level options for widget components.
// These are separate from the semantic FieldMeta so the form engine stays
// decoupled from rendering details.
//
// A typical layout has two layers:
//
//  1. Module-level defaults — built once at boot via DefaultProps() and
//     passed to every formui.Field call on every page.
//  2. Per-call overrides — built with FieldProps{ … } and merged on top of
//     the defaults with DefaultProps().Merge(override).
//
// That way a page that wants "all inputs bordered blue" sets it once and
// only overrides what's different per-field.
type FieldProps struct {
	// Wrapper / layout
	WrapperClass string // class on the outer <div>
	LabelClass   string // class on the <label>

	// Input styling
	Class string // class on the <input> / <select> / <textarea>

	// Behaviour overrides
	Placeholder  string
	TabIndex     int
	Disabled     bool
	ReadOnly     bool
	AutoComplete string // "on" | "off" | "email" | etc.
	AutoFocus    bool
	ID           string // overrides field.Key as the HTML id attribute

	// Hint / helper text shown below the input
	HintText string

	// Textarea-specific
	Rows int // defaults to 3 if zero

	// Datastar live-update target. Empty string disables live updates.
	// When set, every input emits data-on:input__debounce.250ms="@post('<endpoint>')".
	Endpoint string

	// Number-specific HTML attributes
	Step string // "1", "0.01", "any"
	Min  string
	Max  string

	// Accessibility
	AriaLabel string // overrides field.Meta.Label for screen readers
}

// DefaultProps returns a FieldProps with sensible defaults applied.
// Callers should normally start from this and Merge in overrides.
func DefaultProps() FieldProps {
	return FieldProps{
		Rows: 3,
		Step: "any",
	}
}

// Merge returns a new FieldProps where any non-zero value in override
// replaces the corresponding value in base. Useful for layering
// module-level defaults with per-field overrides.
func (base FieldProps) Merge(override FieldProps) FieldProps {
	if override.WrapperClass != "" {
		base.WrapperClass = override.WrapperClass
	}
	if override.LabelClass != "" {
		base.LabelClass = override.LabelClass
	}
	if override.Class != "" {
		base.Class = override.Class
	}
	if override.Placeholder != "" {
		base.Placeholder = override.Placeholder
	}
	if override.TabIndex != 0 {
		base.TabIndex = override.TabIndex
	}
	if override.Disabled {
		base.Disabled = true
	}
	if override.ReadOnly {
		base.ReadOnly = true
	}
	if override.AutoComplete != "" {
		base.AutoComplete = override.AutoComplete
	}
	if override.AutoFocus {
		base.AutoFocus = true
	}
	if override.ID != "" {
		base.ID = override.ID
	}
	if override.HintText != "" {
		base.HintText = override.HintText
	}
	if override.Rows != 0 {
		base.Rows = override.Rows
	}
	if override.Endpoint != "" {
		base.Endpoint = override.Endpoint
	}
	if override.Step != "" {
		base.Step = override.Step
	}
	if override.Min != "" {
		base.Min = override.Min
	}
	if override.Max != "" {
		base.Max = override.Max
	}
	if override.AriaLabel != "" {
		base.AriaLabel = override.AriaLabel
	}
	return base
}

// ResolveDisabled collapses the three input-disable signals (the prop, the
// field's read-only/disabled state, the per-field Writable flag) into a
// single boolean the templates can render. The rule is:
//
//	disabled = props.Disabled || !field.Writable
//
// Returning true means the input is *visually* disabled; check the prop
// separately if you need to distinguish "user disabled" from "field
// read-only".
func (p FieldProps) ResolveDisabled(fieldWritable bool) bool {
	return p.Disabled || !fieldWritable
}

// ResolveReadonly collapses the read-only signals. The rule is:
//
//	readonly = props.ReadOnly || (props.Disabled == false && !field.Writable)
//
// A disabled input is never also read-only — the HTML spec says disabled
// subsumes read-only, so emitting both is noise.
func (p FieldProps) ResolveReadonly(fieldWritable bool) bool {
	if p.Disabled {
		return false
	}
	return p.ReadOnly || !fieldWritable
}
