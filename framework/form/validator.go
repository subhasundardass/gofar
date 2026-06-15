package form

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/subhasundardas/gofar/framework/utils"
)

// =============================================================================
// Core Validator Interface
// =============================================================================
//
// A Validator is a small, composable unit of business logic that knows how
// to inspect a single field value and report whether it is acceptable.
//
// The form engine never requires the user to write one. They get the common
// rules (required, email, length, range, pattern, …) as ready-made types and
// field-option builders, and they can always drop down to the interface
// below for fully custom rules.
//
// Two methods:
//
//   - Name()    returns a short, stable, lower-snake-case identifier
//     ("required", "min_length", "email", "between", …). The name is
//     surfaced to the client so UI code can style or translate errors
//     by category ("all 'required' errors in red", etc.).
//
//   - Validate(value) returns a non-nil error if the value is invalid.
//     Returning nil means "this validator passed". A nil value passed to
//     a non-required validator is always treated as "passed" — that's
//     the Required validator's job, not every validator's.

type Validator interface {
	Name() string
	Validate(value any) error
}

// =============================================================================
// Structured Validation Error
// =============================================================================
//
// This is what every validator SHOULD return. The form layer surfaces it
// to the UI (templ / Datastar / API responses), so it carries everything
// the client needs to render an inline error message without having to
// parse free-form text.
//
// The "Params" map is a free-form bag of validator-specific metadata
// (e.g. {"min": 8, "max": 64}) that the client can use to render a
// localised, parameterised message.

type ValidationError struct {
	Field   string         `json:"field"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Params  map[string]any `json:"params,omitempty"`
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// =============================================================================
// Common Error Helpers
// =============================================================================
//
// Use these helpers when you write a custom validator. They keep the
// error shape consistent and let the UI render code stay simple.

func ErrRequired(field string) error {
	return &ValidationError{
		Field:   field,
		Code:    "required",
		Message: "This field is required",
	}
}

func ErrInvalid(field, msg string) error {
	return &ValidationError{
		Field:   field,
		Code:    "invalid",
		Message: msg,
	}
}

// ErrCoded is the most flexible helper: it lets a custom validator emit
// its own error code, message, and parameter bag.
func ErrCoded(field, code, message string, params ...map[string]any) error {
	ve := &ValidationError{
		Field:   field,
		Code:    code,
		Message: message,
	}
	if len(params) > 0 && len(params[0]) > 0 {
		ve.Params = params[0]
	}
	return ve
}

// =============================================================================
// Required Validator
// =============================================================================
//
// A value is "missing" if any of the following hold:
//   - the value is nil
//   - the value is an empty / whitespace-only string
//   - the value is an empty slice / map / array
//   - the value is a zero-value number (0 / 0.0) and the field is
//     explicitly marked "required-nonzero" via the NonZero() option

type RequiredValidator struct {
	Field   string
	NonZero bool   // when true, also fail on numeric zero / false
	Custom  string // override the default error message
}

func (v RequiredValidator) Name() string { return "required" }

func (v RequiredValidator) Validate(value any) error {
	if isEmpty(value) {
		return v.fail()
	}
	if v.NonZero && isZero(value) {
		return v.fail()
	}
	return nil
}

func (v RequiredValidator) fail() error {
	msg := v.Custom
	if msg == "" {
		msg = "This field is required"
	}
	return &ValidationError{
		Field:   v.Field,
		Code:    "required",
		Message: msg,
	}
}

// =============================================================================
// Internal helpers
// =============================================================================

// isZero reports whether v is the type's zero value. Used by NonZero.
func isZero(v any) bool {
	if v == nil {
		return true
	}
	switch x := v.(type) {
	case bool:
		return !x
	case int:
		return x == 0
	case int32:
		return x == 0
	case int64:
		return x == 0
	case float32:
		return x == 0
	case float64:
		return x == 0
	}
	return false
}

// =============================================================================
// String format validators
// =============================================================================

// EmailValidator accepts RFC-5322-ish addresses.
type EmailValidator struct {
	Field  string
	Custom string
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

func (v EmailValidator) Name() string { return "email" }

func (v EmailValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	str = strings.TrimSpace(str)
	if _, err := mail.ParseAddress(str); err != nil {
		return v.fail()
	}
	if !emailRegex.MatchString(strings.ToLower(str)) {
		return v.fail()
	}
	return nil
}

func (v EmailValidator) fail() error {
	msg := v.Custom
	if msg == "" {
		msg = "Invalid email format"
	}
	return &ValidationError{
		Field:   v.Field,
		Code:    "email",
		Message: msg,
	}
}

// URLValidator accepts any URL that url.Parse considers well-formed
// AND that has a non-empty host.
type URLValidator struct {
	Field        string
	RequireHTTPS bool
	Custom       string
}

func (v URLValidator) Name() string { return "url" }

func (v URLValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	str = strings.TrimSpace(str)
	u, err := url.Parse(str)
	if err != nil || u.Host == "" || u.Scheme == "" {
		return v.fail()
	}
	if v.RequireHTTPS && !strings.EqualFold(u.Scheme, "https") {
		return v.fail()
	}
	return nil
}

func (v URLValidator) fail() error {
	msg := v.Custom
	if msg == "" {
		msg = "Invalid URL"
	}
	return &ValidationError{
		Field:   v.Field,
		Code:    "url",
		Message: msg,
	}
}

// PhoneValidator is a pragmatic phone check: digits, spaces, dashes,
// parens, leading +.
type PhoneValidator struct {
	Field  string
	Custom string
}

var phoneRegex = regexp.MustCompile(`^\+?[0-9\s\-()]{7,20}$`)

func (v PhoneValidator) Name() string { return "phone" }

func (v PhoneValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	if !phoneRegex.MatchString(strings.TrimSpace(str)) {
		msg := v.Custom
		if msg == "" {
			msg = "Invalid phone number"
		}
		return &ValidationError{Field: v.Field, Code: "phone", Message: msg}
	}
	return nil
}

// AlphaValidator: letters only, any case, any script.
type AlphaValidator struct {
	Field  string
	Custom string
}

var alphaRegex = regexp.MustCompile(`^[\p{L}]+$`)

func (v AlphaValidator) Name() string { return "alpha" }

func (v AlphaValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	if !alphaRegex.MatchString(str) {
		msg := v.Custom
		if msg == "" {
			msg = "Only letters are allowed"
		}
		return &ValidationError{Field: v.Field, Code: "alpha", Message: msg}
	}
	return nil
}

// AlphanumericValidator: letters or digits only.
type AlphanumericValidator struct {
	Field  string
	Custom string
}

var alphanumericRegex = regexp.MustCompile(`^[\p{L}0-9]+$`)

func (v AlphanumericValidator) Name() string { return "alphanumeric" }

func (v AlphanumericValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	if !alphanumericRegex.MatchString(str) {
		msg := v.Custom
		if msg == "" {
			msg = "Only letters and digits are allowed"
		}
		return &ValidationError{Field: v.Field, Code: "alphanumeric", Message: msg}
	}
	return nil
}

// NumericValidator: digits only (no decimals, no sign).
type NumericValidator struct {
	Field  string
	Custom string
}

var numericRegex = regexp.MustCompile(`^[0-9]+$`)

func (v NumericValidator) Name() string { return "numeric" }

func (v NumericValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	if !numericRegex.MatchString(str) {
		msg := v.Custom
		if msg == "" {
			msg = "Only digits are allowed"
		}
		return &ValidationError{Field: v.Field, Code: "numeric", Message: msg}
	}
	return nil
}

// =============================================================================
// Length / size validators
// =============================================================================

// MinLengthValidator counts unicode runes (so "ñ" is length 1, not 2).
type MinLengthValidator struct {
	Field  string
	Min    int
	Custom string
}

func (v MinLengthValidator) Name() string { return "min_length" }

func (v MinLengthValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	if runeLen(str) < v.Min {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Minimum length is %d", v.Min)
		}
		return &ValidationError{
			Field: v.Field, Code: "min_length", Message: msg,
			Params: map[string]any{"min": v.Min},
		}
	}
	return nil
}

type MaxLengthValidator struct {
	Field  string
	Max    int
	Custom string
}

func (v MaxLengthValidator) Name() string { return "max_length" }

func (v MaxLengthValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	if runeLen(str) > v.Max {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Maximum length is %d", v.Max)
		}
		return &ValidationError{
			Field: v.Field, Code: "max_length", Message: msg,
			Params: map[string]any{"max": v.Max},
		}
	}
	return nil
}

// BetweenLengthValidator is min_length + max_length in one.
type BetweenLengthValidator struct {
	Field  string
	Min    int
	Max    int
	Custom string
}

func (v BetweenLengthValidator) Name() string { return "between_length" }

func (v BetweenLengthValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	n := runeLen(str)
	if n < v.Min || n > v.Max {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Length must be between %d and %d", v.Min, v.Max)
		}
		return &ValidationError{
			Field: v.Field, Code: "between_length", Message: msg,
			Params: map[string]any{"min": v.Min, "max": v.Max},
		}
	}
	return nil
}

func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// =============================================================================
// Numeric range validators
// =============================================================================

type MinValueValidator struct {
	Field  string
	Min    float64
	Custom string
}

func (v MinValueValidator) Name() string { return "min_value" }

func (v MinValueValidator) Validate(value any) error {
	num, ok := utils.ToFloat(value)
	if !ok {
		return nil
	}
	if num < v.Min {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Minimum value is %v", v.Min)
		}
		return &ValidationError{
			Field: v.Field, Code: "min_value", Message: msg,
			Params: map[string]any{"min": v.Min},
		}
	}
	return nil
}

type MaxValueValidator struct {
	Field  string
	Max    float64
	Custom string
}

func (v MaxValueValidator) Name() string { return "max_value" }

func (v MaxValueValidator) Validate(value any) error {
	num, ok := utils.ToFloat(value)
	if !ok {
		return nil
	}
	if num > v.Max {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Maximum value is %v", v.Max)
		}
		return &ValidationError{
			Field: v.Field, Code: "max_value", Message: msg,
			Params: map[string]any{"max": v.Max},
		}
	}
	return nil
}

// BetweenValueValidator: numeric value must lie in [Min, Max].
type BetweenValueValidator struct {
	Field  string
	Min    float64
	Max    float64
	Custom string
}

func (v BetweenValueValidator) Name() string { return "between_value" }

func (v BetweenValueValidator) Validate(value any) error {
	num, ok := utils.ToFloat(value)
	if !ok {
		return nil
	}
	if num < v.Min || num > v.Max {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Value must be between %v and %v", v.Min, v.Max)
		}
		return &ValidationError{
			Field: v.Field, Code: "between_value", Message: msg,
			Params: map[string]any{"min": v.Min, "max": v.Max},
		}
	}
	return nil
}

// =============================================================================
// Choice / membership validators
// =============================================================================

// InValidator: the value must be one of the allowed values. The comparison
// uses fmt.Sprint, which is the only sane way to compare mixed JSON-decoded
// payloads in Datastar.
type InValidator struct {
	Field  string
	Values []any
	Custom string
}

func (v InValidator) Name() string { return "in" }

func (v InValidator) Validate(value any) error {
	if value == nil {
		return nil
	}
	got := fmt.Sprintf("%v", value)
	for _, allowed := range v.Values {
		if fmt.Sprintf("%v", allowed) == got {
			return nil
		}
	}
	msg := v.Custom
	if msg == "" {
		msg = fmt.Sprintf("Must be one of: %s", joinAny(v.Values))
	}
	return &ValidationError{
		Field: v.Field, Code: "in", Message: msg,
		Params: map[string]any{"values": v.Values},
	}
}

// NotInValidator: the value must NOT be one of the forbidden values.
type NotInValidator struct {
	Field  string
	Values []any
	Custom string
}

func (v NotInValidator) Name() string { return "not_in" }

func (v NotInValidator) Validate(value any) error {
	if value == nil {
		return nil
	}
	got := fmt.Sprintf("%v", value)
	for _, forbidden := range v.Values {
		if fmt.Sprintf("%v", forbidden) == got {
			msg := v.Custom
			if msg == "" {
				msg = fmt.Sprintf("Must not be any of: %s", joinAny(v.Values))
			}
			return &ValidationError{Field: v.Field, Code: "not_in", Message: msg}
		}
	}
	return nil
}

func joinAny(vs []any) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(parts, ", ")
}

// =============================================================================
// Pattern validators
// =============================================================================

// RegexpValidator fails when the value does not match the pattern.
type RegexpValidator struct {
	Field   string
	Pattern string
	Custom  string
}

func (v RegexpValidator) Name() string { return "regexp" }

func (v RegexpValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	re, err := regexp.Compile(v.Pattern)
	if err != nil {
		// Bad pattern: surface as a code error so devs notice.
		return &ValidationError{
			Field: v.Field, Code: "regexp_compile",
			Message: "Server-side validation pattern is invalid",
		}
	}
	if !re.MatchString(str) {
		msg := v.Custom
		if msg == "" {
			msg = "Invalid format"
		}
		return &ValidationError{
			Field: v.Field, Code: "regexp", Message: msg,
		}
	}
	return nil
}

// =============================================================================
// Date validators
// =============================================================================
//
// DateValidator parses a string as a date in one of the supported layouts
// and checks that it lies after a given minimum, before a given maximum,
// or in a closed [Min, Max] range. Layouts default to the ISO 8601 forms
// most JSON APIs use; pass a different one to support legacy clients.

type DateValidator struct {
	Field   string
	Min     string // ISO date or datetime; "" means no lower bound
	Max     string // ISO date or datetime; "" means no upper bound
	Layouts []string
	Custom  string
}

func (v DateValidator) Name() string { return "date" }

func (v DateValidator) Validate(value any) error {
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	layouts := v.Layouts
	if len(layouts) == 0 {
		layouts = []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
	}
	var t time.Time
	var err error
	for _, layout := range layouts {
		if t, err = time.Parse(layout, str); err == nil {
			break
		}
	}
	if err != nil {
		return v.fail("date_invalid", "Invalid date format", nil)
	}
	if v.Min != "" {
		minT, perr := parseDateLoose(v.Min, layouts)
		if perr == nil && t.Before(minT) {
			return v.fail("date_min", fmt.Sprintf("Date must be on or after %s", v.Min),
				map[string]any{"min": v.Min})
		}
	}
	if v.Max != "" {
		maxT, perr := parseDateLoose(v.Max, layouts)
		if perr == nil && t.After(maxT) {
			return v.fail("date_max", fmt.Sprintf("Date must be on or before %s", v.Max),
				map[string]any{"max": v.Max})
		}
	}
	return nil
}

func (v DateValidator) fail(code, msg string, params map[string]any) error {
	if v.Custom != "" {
		msg = v.Custom
	}
	return &ValidationError{Field: v.Field, Code: code, Message: msg, Params: params}
}

func parseDateLoose(s string, layouts []string) (time.Time, error) {
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse %q", s)
}

// =============================================================================
// Cross-field validators
// =============================================================================
//
// Cross-field validators need access to the *other* field's value, so they
// wrap a function that takes a context containing the instance's full
// value bag. The form engine calls them with the current Instance so they
// can read whatever sibling fields they need.

// FieldValueFunc is the shape every cross-field validator uses. It receives
// the value of THIS field plus the surrounding *Instance and returns nil
// for pass or a non-nil error for fail.
type FieldValueFunc func(thisValue any, inst *Instance) error

// ContextValidator runs a function with full access to the instance.
// Use it for cross-field rules that don't fit one of the built-ins below.
type ContextValidator struct {
	Field    string
	CodeName string // code surfaced on error; defaults to "custom"
	Fn       FieldValueFunc
}

func (v ContextValidator) Name() string {
	if v.CodeName == "" {
		return "custom"
	}
	return v.CodeName
}

// Validate delegates to the wrapped function with a no-instance stand-in.
// Cross-field checks done via ContextValidator still need an Instance to
// look at siblings; the form engine routes the instance through a special
// call path (see Instance.Validate below).
func (v ContextValidator) Validate(value any) error {
	if v.Fn == nil {
		return nil
	}
	return v.Fn(value, nil)
}

// SameAsValidator: this field's value must equal the other field's value.
// Comparison goes through fmt.Sprint so the JSON-decoded payload round-trips.
type SameAsValidator struct {
	Field  string
	Other  string
	Custom string
}

func (v SameAsValidator) Name() string { return "same_as" }

// Validate is the no-context variant; the real check is done by
// ValidateWith so an Instance is available.
func (v SameAsValidator) Validate(value any) error { return nil }

// ValidateWith implements cross-field access.
func (v SameAsValidator) ValidateWith(value any, inst *Instance) error {
	if inst == nil {
		return nil
	}
	other := inst.Get(v.Other)
	if fmt.Sprintf("%v", value) != fmt.Sprintf("%v", other) {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Must match %s", v.Other)
		}
		return &ValidationError{
			Field: v.Field, Code: "same_as", Message: msg,
			Params: map[string]any{"other": v.Other},
		}
	}
	return nil
}

// DifferentFromValidator: this field's value must NOT equal the other field's.
type DifferentFromValidator struct {
	Field  string
	Other  string
	Custom string
}

func (v DifferentFromValidator) Name() string { return "different_from" }
func (v DifferentFromValidator) Validate(value any) error {
	return nil
}
func (v DifferentFromValidator) ValidateWith(value any, inst *Instance) error {
	if inst == nil {
		return nil
	}
	other := inst.Get(v.Other)
	if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", other) {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Must differ from %s", v.Other)
		}
		return &ValidationError{
			Field: v.Field, Code: "different_from", Message: msg,
		}
	}
	return nil
}

// RequiredIfValidator: required only when another field equals a target value.
// When otherField is "truthy" (non-empty, non-zero, non-false) the field
// becomes required.
type RequiredIfValidator struct {
	Field       string
	Other       string
	OtherValue  any
	MatchTruthy bool
	Custom      string
}

func (v RequiredIfValidator) Name() string             { return "required_if" }
func (v RequiredIfValidator) Validate(value any) error { return nil }
func (v RequiredIfValidator) ValidateWith(value any, inst *Instance) error {
	if inst == nil {
		return nil
	}
	other := inst.Get(v.Other)
	trigger := false
	if v.MatchTruthy {
		trigger = !isEmpty(other) && !isZero(other)
	} else {
		trigger = fmt.Sprintf("%v", other) == fmt.Sprintf("%v", v.OtherValue)
	}
	if !trigger {
		return nil
	}
	if isEmpty(value) {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Required when %s is set", v.Other)
		}
		return &ValidationError{Field: v.Field, Code: "required_if", Message: msg}
	}
	return nil
}

// RequiredUnlessValidator: required UNLESS another field equals a target
// value (or, with MatchTruthy=true, unless the other field is truthy).
type RequiredUnlessValidator struct {
	Field       string
	Other       string
	OtherValue  any
	MatchTruthy bool
	Custom      string
}

func (v RequiredUnlessValidator) Name() string             { return "required_unless" }
func (v RequiredUnlessValidator) Validate(value any) error { return nil }
func (v RequiredUnlessValidator) ValidateWith(value any, inst *Instance) error {
	if inst == nil {
		return nil
	}
	other := inst.Get(v.Other)
	skip := false
	if v.MatchTruthy {
		skip = !isEmpty(other) && !isZero(other)
	} else {
		skip = fmt.Sprintf("%v", other) == fmt.Sprintf("%v", v.OtherValue)
	}
	if skip {
		return nil
	}
	if isEmpty(value) {
		msg := v.Custom
		if msg == "" {
			msg = fmt.Sprintf("Required unless %s is set", v.Other)
		}
		return &ValidationError{Field: v.Field, Code: "required_unless", Message: msg}
	}
	return nil
}

// =============================================================================
// Conditional / Optional wrapper
// =============================================================================
//
// ConditionalValidator wraps another validator and only runs it when the
// supplied predicate returns true. Use it for "if A is set, B must be
// numeric" and similar one-off rules without writing a new validator type.

type ConditionalValidator struct {
	Inner Validator
	When  func(inst *Instance) bool
}

func (v ConditionalValidator) Name() string             { return v.Inner.Name() }
func (v ConditionalValidator) Validate(value any) error { return nil }
func (v ConditionalValidator) ValidateWith(value any, inst *Instance) error {
	if v.When != nil && !v.When(inst) {
		return nil
	}
	return v.Inner.Validate(value)
}

// OptionalValidator wraps another validator so it only runs when the
// value is non-empty. It is the moral equivalent of "nullable" rules
// in Laravel / Yup / Zod.
type OptionalValidator struct {
	Inner Validator
}

func (v OptionalValidator) Name() string { return v.Inner.Name() }
func (v OptionalValidator) Validate(value any) error {
	if isEmpty(value) {
		return nil
	}
	return v.Inner.Validate(value)
}
func (v OptionalValidator) ValidateWith(value any, inst *Instance) error {
	if isEmpty(value) {
		return nil
	}
	if cw, ok := v.Inner.(interface {
		ValidateWith(value any, inst *Instance) error
	}); ok {
		return cw.ValidateWith(value, inst)
	}
	return v.Inner.Validate(value)
}
