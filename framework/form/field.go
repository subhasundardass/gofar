package form

// Field is the public contract every form field must satisfy. It exposes
// everything a renderer (templ, Datastar, JSON schema) needs to draw or
// describe a field without depending on the concrete *BaseField struct.
//
// Concrete field types (TextField, NumberField, …) embed BaseField and
// therefore satisfy this interface for free.
type Field interface {
	// Identity
	GetKey() string
	GetType() string

	// Presentation
	GetLabel() string
	GetPlaceholder() string
	GetHelpText() string
	GetDefaultValue() any

	// Behaviour
	GetRequired() bool
	GetReadOnly() bool
	GetDisabled() bool

	// Validation
	GetValidators() []Validator

	// Reactive Rule Engine
	GetVisibleIf() Rule
	GetEnabledIf() Rule
	GetRequiredIf() Rule
	GetComputeWith() ComputeFunc
}
