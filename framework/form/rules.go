package form

import (
	"fmt"
	"reflect"
	"strings"
)

// =============================================================================
// Rule — the reactive form engine
// =============================================================================
//
// A Rule is a pure function from *FormState → bool. The pointer receiver
// avoids copying the sync.RWMutex embedded in FormState.
//
// Rules compose freely:
//
//   VisibleIf(And(FieldEquals("type", "company"), Not(FieldEmpty("name"))))
//   RequiredIf(FieldEquals("country", "US"))
//   EnabledIf(FieldNotEquals("status", "locked"))
//
// All rules are evaluated lazily against the live *FormState, so they
// automatically react when any field value changes.

// Rule is the single interface every predicate must satisfy.
// Eval receives *FormState (never FormState by value) to avoid copying
// the embedded sync.RWMutex.
type Rule interface {
	Eval(state *FormState) bool
}

// RuleFunc lets you write a one-off rule as a plain function without
// defining a new type.
//
//   VisibleIf(RuleFunc(func(s *form.FormState) bool {
//       return s.Get("age") != nil
//   }))
type RuleFunc func(state *FormState) bool

func (f RuleFunc) Eval(state *FormState) bool { return f(state) }

// =============================================================================
// Logical combinators
// =============================================================================

// andRule is true when ALL of its children are true.
type andRule struct{ rules []Rule }

func (a andRule) Eval(s *FormState) bool {
	for _, r := range a.rules {
		if !r.Eval(s) {
			return false
		}
	}
	return true
}

// orRule is true when ANY of its children is true.
type orRule struct{ rules []Rule }

func (o orRule) Eval(s *FormState) bool {
	for _, r := range o.rules {
		if r.Eval(s) {
			return true
		}
	}
	return false
}

// notRule inverts its child.
type notRule struct{ rule Rule }

func (n notRule) Eval(s *FormState) bool { return !n.rule.Eval(s) }

// And returns a Rule that is true when ALL supplied rules are true.
func And(rules ...Rule) Rule { return andRule{rules: rules} }

// Or returns a Rule that is true when AT LEAST ONE supplied rule is true.
func Or(rules ...Rule) Rule { return orRule{rules: rules} }

// Not returns a Rule that inverts the supplied rule.
func Not(rule Rule) Rule { return notRule{rule: rule} }

// Always is a constant Rule that is always true. Useful for making a
// field unconditionally required via RequiredIf(Always).
var Always Rule = RuleFunc(func(_ *FormState) bool { return true })

// Never is a constant Rule that is always false. Useful for disabling a
// field unconditionally via EnabledIf(Never).
var Never Rule = RuleFunc(func(_ *FormState) bool { return false })

// =============================================================================
// Field-value predicates
// =============================================================================

// fieldEqRule checks whether a field's value equals a target.
type fieldEqRule struct {
	key    string
	target any
}

func (r fieldEqRule) Eval(s *FormState) bool {
	return looseEqual(s.Get(r.key), r.target)
}

// fieldNotEqRule is the inverse.
type fieldNotEqRule struct {
	key    string
	target any
}

func (r fieldNotEqRule) Eval(s *FormState) bool {
	return !looseEqual(s.Get(r.key), r.target)
}

// fieldEmptyRule is true when the field has no meaningful value.
type fieldEmptyRule struct{ key string }

func (r fieldEmptyRule) Eval(s *FormState) bool { return isEmpty(s.Get(r.key)) }

// fieldNotEmptyRule is true when the field has a meaningful value.
type fieldNotEmptyRule struct{ key string }

func (r fieldNotEmptyRule) Eval(s *FormState) bool { return !isEmpty(s.Get(r.key)) }

// fieldContainsRule is true when a string field contains the substring.
type fieldContainsRule struct {
	key       string
	substring string
}

func (r fieldContainsRule) Eval(s *FormState) bool {
	v := fmt.Sprint(s.Get(r.key))
	return strings.Contains(v, r.substring)
}

// fieldGTRule is true when a numeric field is greater than threshold.
type fieldGTRule struct {
	key       string
	threshold float64
}

func (r fieldGTRule) Eval(s *FormState) bool {
	return toFloat(s.Get(r.key)) > r.threshold
}

// fieldGTERule is true when a numeric field is ≥ threshold.
type fieldGTERule struct {
	key       string
	threshold float64
}

func (r fieldGTERule) Eval(s *FormState) bool {
	return toFloat(s.Get(r.key)) >= r.threshold
}

// fieldLTRule is true when a numeric field is less than threshold.
type fieldLTRule struct {
	key       string
	threshold float64
}

func (r fieldLTRule) Eval(s *FormState) bool {
	return toFloat(s.Get(r.key)) < r.threshold
}

// fieldLTERule is true when a numeric field is ≤ threshold.
type fieldLTERule struct {
	key       string
	threshold float64
}

func (r fieldLTERule) Eval(s *FormState) bool {
	return toFloat(s.Get(r.key)) <= r.threshold
}

// fieldInRule is true when a field's value is one of the allowed set.
type fieldInRule struct {
	key     string
	allowed []any
}

func (r fieldInRule) Eval(s *FormState) bool {
	v := s.Get(r.key)
	for _, a := range r.allowed {
		if looseEqual(v, a) {
			return true
		}
	}
	return false
}

// fieldNotInRule is the inverse of fieldInRule.
type fieldNotInRule struct {
	key     string
	allowed []any
}

func (r fieldNotInRule) Eval(s *FormState) bool {
	v := s.Get(r.key)
	for _, a := range r.allowed {
		if looseEqual(v, a) {
			return false
		}
	}
	return true
}

// fieldTrueRule is true when a boolean field is truthy.
type fieldTrueRule struct{ key string }

func (r fieldTrueRule) Eval(s *FormState) bool { return isTruthy(s.Get(r.key)) }

// fieldFalseRule is true when a boolean field is falsy.
type fieldFalseRule struct{ key string }

func (r fieldFalseRule) Eval(s *FormState) bool { return !isTruthy(s.Get(r.key)) }

// =============================================================================
// Public constructors for field-value predicates
// =============================================================================

// FieldEquals returns a Rule that is true when the field at key equals target.
// Comparison is loose: "42" == 42, "true" == true, etc.
func FieldEquals(key string, target any) Rule { return fieldEqRule{key: key, target: target} }

// FieldNotEquals returns a Rule that is true when the field at key does NOT equal target.
func FieldNotEquals(key string, target any) Rule {
	return fieldNotEqRule{key: key, target: target}
}

// FieldEmpty returns a Rule that is true when the field has no value ("", nil, 0, false).
func FieldEmpty(key string) Rule { return fieldEmptyRule{key: key} }

// FieldNotEmpty returns a Rule that is true when the field has a non-empty value.
func FieldNotEmpty(key string) Rule { return fieldNotEmptyRule{key: key} }

// FieldContains returns a Rule that is true when the string field at key
// contains the given substring.
func FieldContains(key, substring string) Rule {
	return fieldContainsRule{key: key, substring: substring}
}

// FieldGT returns a Rule that is true when the numeric field at key > threshold.
func FieldGT(key string, threshold float64) Rule { return fieldGTRule{key: key, threshold: threshold} }

// FieldGTE returns a Rule that is true when the numeric field at key >= threshold.
func FieldGTE(key string, threshold float64) Rule {
	return fieldGTERule{key: key, threshold: threshold}
}

// FieldLT returns a Rule that is true when the numeric field at key < threshold.
func FieldLT(key string, threshold float64) Rule { return fieldLTRule{key: key, threshold: threshold} }

// FieldLTE returns a Rule that is true when the numeric field at key <= threshold.
func FieldLTE(key string, threshold float64) Rule {
	return fieldLTERule{key: key, threshold: threshold}
}

// FieldIn returns a Rule that is true when the field at key is one of values.
func FieldIn(key string, values ...any) Rule { return fieldInRule{key: key, allowed: values} }

// FieldNotIn returns a Rule that is true when the field at key is NOT in values.
func FieldNotIn(key string, values ...any) Rule { return fieldNotInRule{key: key, allowed: values} }

// FieldTrue returns a Rule that is true when the boolean field at key is truthy.
func FieldTrue(key string) Rule { return fieldTrueRule{key: key} }

// FieldFalse returns a Rule that is true when the boolean field at key is falsy.
func FieldFalse(key string) Rule { return fieldFalseRule{key: key} }

// =============================================================================
// Computed fields
// =============================================================================

// ComputeFunc is the signature for a computed-field function. It receives
// a *FormState (never by value) and returns the computed value to inject.
type ComputeFunc func(state *FormState) any

// computedRule wraps a ComputeFunc so it also satisfies Rule (always true).
type computedRule struct {
	key     string
	compute ComputeFunc
}

func (c computedRule) Eval(_ *FormState) bool { return true }

// ComputedRule extracts a computedRule from a Rule, returning (rule, true)
// if the Rule is a computed field rule, else (_, false).
func ComputedRule(r Rule) (computedRule, bool) {
	cr, ok := r.(computedRule)
	return cr, ok
}

// =============================================================================
// FieldOption wrappers for rules
// =============================================================================

// VisibleIf attaches a visibility rule. The field is rendered only when the rule is true.
func VisibleIf(r Rule) FieldOption {
	return func(f *BaseField) {
		f.VisibleIf = r
	}
}

// EnabledIf attaches an enabled rule. When false the field renders as disabled.
func EnabledIf(r Rule) FieldOption {
	return func(f *BaseField) {
		f.EnabledIf = r
	}
}

// RequiredIf attaches a dynamic required rule. The field is required when
// the rule is true, regardless of the static Required flag.
func RequiredIf(r Rule) FieldOption {
	return func(f *BaseField) {
		f.RequiredIf = r
	}
}

// ComputedAs attaches a compute function. Each time the form engine evaluates
// the field it calls fn(state) and stores the result back into the state so
// downstream rules can depend on it.
//
//   form.TextInput("fullName",
//       form.Label("Full Name"),
//       form.ComputedAs(func(s *form.FormState) any {
//           first := form.Stringify(s.Get("firstName"))
//           last  := form.Stringify(s.Get("lastName"))
//           return strings.TrimSpace(first + " " + last)
//       }),
//   )
func ComputedAs(fn ComputeFunc) FieldOption {
	return func(f *BaseField) {
		f.ComputeWith = fn
	}
}

// =============================================================================
// Internal helpers
// =============================================================================

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x) == ""
	case bool:
		return !x
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() == 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() == 0
	case float32, float64:
		return reflect.ValueOf(v).Float() == 0
	}
	return false
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x == "true" || x == "1" || x == "yes"
	case float64:
		return x != 0
	case int:
		return x != 0
	}
	return false
}

func toFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		var f float64
		fmt.Sscanf(x, "%f", &f)
		return f
	}
	return 0
}

// looseEqual compares two values with type coercion so that "42" == float64(42),
// "true" == true, etc. This matches how Datastar sends field values over the wire
// (strings) against Go values (typed).
func looseEqual(a, b any) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}
