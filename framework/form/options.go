package form

// FieldOption is the functional-options type used to configure a BaseField
// when constructing a concrete field via the public helpers (TextInput,
// NumberInput, …). Adding a new option never breaks the public API — you
// just append a new function.
type FieldOption func(*BaseField)

// Label sets the human-readable label rendered next to the field.
func Label(label string) FieldOption {
	return func(f *BaseField) {
		f.Label = label
	}
}

// Placeholder sets the placeholder text shown when the field is empty.
func Placeholder(placeholder string) FieldOption {
	return func(f *BaseField) {
		f.Placeholder = placeholder
	}
}

// HelpText sets supplementary guidance rendered beneath the field.
func HelpText(help string) FieldOption {
	return func(f *BaseField) {
		f.HelpText = help
	}
}

// Default sets the initial value for the field. This value is what the form
// hydrates with when no user input has been received yet.
func Default(value any) FieldOption {
	return func(f *BaseField) {
		f.DefaultValue = value
	}
}

// ReadOnly marks the field as read-only.
func ReadOnly() FieldOption {
	return func(f *BaseField) {
		f.ReadOnly = true
	}
}

// Disabled marks the field as disabled (static, state-independent).
// For dynamic enable/disable logic use EnabledIf(rule) instead.
func Disabled() FieldOption {
	return func(f *BaseField) {
		f.Disabled = true
	}
}

// Required marks the field as statically required AND appends a
// RequiredValidator so the error message names the field.
// For dynamic required logic use RequiredIf(rule) instead.
func Required() FieldOption {
	return func(f *BaseField) {
		f.Required = true
		f.Validators = append(f.Validators, RequiredValidator{Field: f.Key})
	}
}

// WithValidator appends a custom validator to the field.
func WithValidator(v Validator) FieldOption {
	return func(f *BaseField) {
		f.Validators = append(f.Validators, v)
	}
}
