package formui

// FieldProps carries the CSS class overrides for each part of a form field
// widget. It is the only coupling between the form engine and the renderer —
// all layout decisions live here, not inside the templ components.
//
// Two-layer pattern:
//
//  1. Module-level theme — built once at startup via DefaultProps() (or one
//     of the preset constructors) and passed to every @formui.Field call.
//  2. Per-call overrides — merged via With* fluent helpers so only differences
//     are applied without mutating the original.
//
// Example:
//
//	props := formui.DefaultProps()
//	@formui.Field(emailField, props.WithInputClass("border-blue-500"))
type FieldProps struct {
	// Outer wrapper div / fieldset
	WrapperClass string
	// <label> element
	LabelClass string
	// <input> / <select> / <textarea> / <input type="checkbox"> element
	InputClass string
	// Error message <p>
	ErrorClass string
	// Hint / help text <p>
	HintClass string
	// Extra class appended to InputClass when the field has a validation error
	InputErrorClass string
}

// Merge returns a new FieldProps where any non-empty field in override
// replaces the corresponding field in the receiver. The receiver is unchanged.
func (p FieldProps) Merge(override FieldProps) FieldProps {
	if override.WrapperClass != "" {
		p.WrapperClass = override.WrapperClass
	}
	if override.LabelClass != "" {
		p.LabelClass = override.LabelClass
	}
	if override.InputClass != "" {
		p.InputClass = override.InputClass
	}
	if override.ErrorClass != "" {
		p.ErrorClass = override.ErrorClass
	}
	if override.HintClass != "" {
		p.HintClass = override.HintClass
	}
	if override.InputErrorClass != "" {
		p.InputErrorClass = override.InputErrorClass
	}
	return p
}

// inputClass returns the effective input class — base + error modifier when
// there is a non-empty error string.
func (p FieldProps) inputClass(hasError bool) string {
	if hasError && p.InputErrorClass != "" {
		return cx(p.InputClass, p.InputErrorClass)
	}
	return p.InputClass
}

// ── Fluent overrides ──────────────────────────────────────────────────────────

func (p FieldProps) WithWrapperClass(c string) FieldProps    { p.WrapperClass = c; return p }
func (p FieldProps) WithLabelClass(c string) FieldProps      { p.LabelClass = c; return p }
func (p FieldProps) WithInputClass(c string) FieldProps      { p.InputClass = c; return p }
func (p FieldProps) WithErrorClass(c string) FieldProps      { p.ErrorClass = c; return p }
func (p FieldProps) WithHintClass(c string) FieldProps       { p.HintClass = c; return p }
func (p FieldProps) WithInputErrorClass(c string) FieldProps { p.InputErrorClass = c; return p }

// ── Preset themes ─────────────────────────────────────────────────────────────

// DefaultProps returns the standard stacked-label style.
// Suitable for most full-width form layouts.
func DefaultProps() FieldProps {
	return FieldProps{
		WrapperClass:    "form-group mb-4",
		LabelClass:      "block text-sm font-semibold text-gray-700 mb-1",
		InputClass:      "w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500",
		ErrorClass:      "text-red-600 text-xs mt-1",
		HintClass:       "text-gray-500 text-xs mt-1",
		InputErrorClass: "border-red-500 focus:ring-red-500",
	}
}

// CompactProps returns a tighter inline style.
// Use for forms where vertical space is constrained.
func CompactProps() FieldProps {
	return FieldProps{
		WrapperClass:    "flex gap-2 items-center mb-2",
		LabelClass:      "font-semibold w-32 shrink-0 text-sm text-gray-700",
		InputClass:      "border rounded px-2 py-1 text-xs flex-1",
		ErrorClass:      "text-red-500 text-xs",
		HintClass:       "text-gray-400 text-xs",
		InputErrorClass: "border-red-500",
	}
}

// InlineProps returns a minimal bottom-border-only style.
// Use for inline field groups or data-entry rows.
func InlineProps() FieldProps {
	return FieldProps{
		WrapperClass:    "flex items-center gap-1 flex-1",
		LabelClass:      "text-sm font-semibold shrink-0 w-32 text-gray-700",
		InputClass:      "border-b border-gray-300 px-1 py-0.5 text-sm flex-1 bg-transparent focus:outline-none focus:border-blue-500",
		ErrorClass:      "text-red-500 text-xs",
		HintClass:       "text-gray-400 text-xs",
		InputErrorClass: "border-red-500",
	}
}
