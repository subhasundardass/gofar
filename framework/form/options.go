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

// Disabled marks the field as disabled.
func Disabled() FieldOption {
	return func(f *BaseField) {
		f.Disabled = true
	}
}

// Required marks the field as required AND appends a RequiredValidator
// that knows the field's Key. The Key is captured here (at construction
// time) rather than at validation time so that error messages can always
// identify the offending field even if the field is later passed around
// generically.
func Required() FieldOption {
	return func(f *BaseField) {
		f.Required = true
		f.Validators = append(f.Validators, RequiredValidator{Field: f.Key})
	}
}

// WithValidator appends a custom validator to the field. Useful for
// domain-specific rules that don't ship in the box.
func WithValidator(v Validator) FieldOption {
	return func(f *BaseField) {
		f.Validators = append(f.Validators, v)
	}
}
