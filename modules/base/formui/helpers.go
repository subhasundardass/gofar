package formui

import (
	"strings"

	"github.com/subhasundardas/gofar/framework/form"
)

// ── Element IDs ───────────────────────────────────────────────────────────────

// fieldID returns the HTML id attribute value for a field's input element.
func fieldID(field form.UIField) string { return field.Key }

// wrapperID returns a deterministic id for the outer wrapper element.
// Used by Datastar effects to target and swap the whole field group.
func wrapperID(key string) string { return "wrapper-" + key }

// hintID returns the id of the hint/help text element.
// Wired to input's aria-describedby.
func hintID(key string) string { return "hint-" + key }

// errorID returns the id of the error message element.
// Wired to input's aria-errormessage.
func errorID(key string) string { return "error-" + key }

// ── HTML type mapping ─────────────────────────────────────────────────────────

// htmlType maps a form.FieldType to the HTML input type attribute value.
// Anything unrecognised falls back to "text".
func htmlType(t form.FieldType) string {
	switch t {
	case form.Email:
		return "email"
	case form.Password:
		return "password"
	case form.Date:
		return "date"
	case form.Time:
		return "time"
	case form.DateTime:
		return "datetime-local"
	case form.File:
		return "file"
	case form.URL:
		return "url"
	case form.Phone:
		return "tel"
	case form.Color:
		return "color"
	case form.Number:
		return "number"
	case form.Hidden:
		return "hidden"
	default:
		return "text"
	}
}

// ── CSS helpers ───────────────────────────────────────────────────────────────

// cx joins multiple class strings, discarding empty ones.
// Use instead of the old concatClasses.
func cx(classes ...string) string {
	out := make([]string, 0, len(classes))
	for _, c := range classes {
		if c = strings.TrimSpace(c); c != "" {
			out = append(out, c)
		}
	}
	return strings.Join(out, " ")
}

// cxIf returns class when condition is true, else "".
// Use instead of the old toClass.
func cxIf(class string, condition bool) string {
	if condition {
		return class
	}
	return ""
}

// ── Value helpers ─────────────────────────────────────────────────────────────

// boolVal reports whether a field value is truthy ("true", "1", "yes", true).
// Used exclusively by CheckboxField.
func boolVal(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		s := strings.ToLower(strings.TrimSpace(x))
		return s == "true" || s == "1" || s == "yes"
	}
	return false
}

// ── Aria helpers ──────────────────────────────────────────────────────────────

// ariaDescribedBy returns a space-separated aria-describedby attribute value,
// including the hint id (when a hint exists) and the error id (when an error exists).
func ariaDescribedBy(key string, hasHint, hasError bool) string {
	var ids []string
	if hasHint {
		ids = append(ids, hintID(key))
	}
	if hasError {
		ids = append(ids, errorID(key))
	}
	return strings.Join(ids, " ")
}
