package form

func newField(fieldType, key string, opts ...FieldOption) BaseField {
	field := BaseField{
		Key:  key,
		Type: fieldType,
	}

	for _, opt := range opts {
		opt(&field)
	}

	return field
}

// Text

type TextField struct {
	BaseField
}

func TextInput(key string, opts ...FieldOption) *TextField {
	return &TextField{
		BaseField: newField("text", key, opts...),
	}
}

// TextArea

type TextAreaField struct {
	BaseField
}

func TextAreaInput(key string, opts ...FieldOption) *TextAreaField {
	return &TextAreaField{
		BaseField: newField("textarea", key, opts...),
	}
}

// Number

type NumberField struct {
	BaseField
}

func NumberInput(key string, opts ...FieldOption) *NumberField {
	return &NumberField{
		BaseField: newField("number", key, opts...),
	}
}

// Email

type EmailField struct {
	BaseField
}

func EmailInput(key string, opts ...FieldOption) *EmailField {
	return &EmailField{
		BaseField: newField("email", key, opts...),
	}
}

// Password

type PasswordField struct {
	BaseField
}

func PasswordInput(key string, opts ...FieldOption) *PasswordField {
	return &PasswordField{
		BaseField: newField("password", key, opts...),
	}
}

// Date

type DateField struct {
	BaseField
}

func DateInput(key string, opts ...FieldOption) *DateField {
	return &DateField{
		BaseField: newField("date", key, opts...),
	}
}

// Time

type TimeField struct {
	BaseField
}

func TimeInput(key string, opts ...FieldOption) *TimeField {
	return &TimeField{
		BaseField: newField("time", key, opts...),
	}
}

// DateTime

type DateTimeField struct {
	BaseField
}

func DateTimeInput(key string, opts ...FieldOption) *DateTimeField {
	return &DateTimeField{
		BaseField: newField("datetime", key, opts...),
	}
}

// Checkbox

type CheckboxField struct {
	BaseField
}

func CheckboxInput(key string, opts ...FieldOption) *CheckboxField {
	return &CheckboxField{
		BaseField: newField("checkbox", key, opts...),
	}
}

// Radio

type RadioField struct {
	BaseField
}

func RadioInput(key string, opts ...FieldOption) *RadioField {
	return &RadioField{
		BaseField: newField("radio", key, opts...),
	}
}

// Hidden

type HiddenField struct {
	BaseField
}

func HiddenInput(key string, opts ...FieldOption) *HiddenField {
	return &HiddenField{
		BaseField: newField("hidden", key, opts...),
	}
}

// File

type FileField struct {
	BaseField
}

func FileInput(key string, opts ...FieldOption) *FileField {
	return &FileField{
		BaseField: newField("file", key, opts...),
	}
}

// Image

type ImageField struct {
	BaseField
}

func ImageInput(key string, opts ...FieldOption) *ImageField {
	return &ImageField{
		BaseField: newField("image", key, opts...),
	}
}

// URL

type URLField struct {
	BaseField
}

func URLInput(key string, opts ...FieldOption) *URLField {
	return &URLField{
		BaseField: newField("url", key, opts...),
	}
}

// Phone

type PhoneField struct {
	BaseField
}

func PhoneInput(key string, opts ...FieldOption) *PhoneField {
	return &PhoneField{
		BaseField: newField("phone", key, opts...),
	}
}

// Color

type ColorField struct {
	BaseField
}

func ColorInput(key string, opts ...FieldOption) *ColorField {
	return &ColorField{
		BaseField: newField("color", key, opts...),
	}
}

// Range

type RangeField struct {
	BaseField
}

func RangeInput(key string, opts ...FieldOption) *RangeField {
	return &RangeField{
		BaseField: newField("range", key, opts...),
	}
}
