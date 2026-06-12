package formui

import "github.com/subhasundardas/gofar/framework/form"

// fieldID returns props.ID if set, otherwise field.Key.
func fieldID(field form.UIField, props FieldProps) string {
	if props.ID != "" {
		return props.ID
	}
	return field.Key
}

// fieldType maps a semantic FieldType to the matching HTML input type
// attribute. Anything we don't recognise falls back to "text" — Datastar
// and the form engine don't care which input type is rendered as long as
// the value round-trips.
func fieldType(t form.FieldType) string {
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
	case form.Image:
		return "image"
	case form.URL:
		return "url"
	case form.Phone:
		return "tel"
	case form.Color:
		return "color"
	case form.Range:
		return "range"
	case form.Hidden:
		return "hidden"
	default:
		return "text"
	}
}

// liveUpdateAttr returns the Datastar input-debounce attribute for the
// given endpoint. Returns an empty string if no endpoint is configured —
// the caller should then wrap the result in `if` to skip emitting the
// attribute entirely. Centralising this here means the debounce delay and
// plugin key are identical across every input template.
//
// The format is: data-on:input__debounce.250ms="@post('/path')"
func liveUpdateAttr(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	return "@post('" + endpoint + "')"
}

// rowsOrDefault returns rows if positive, otherwise 3 — the standard
// single-line textarea height.
func rowsOrDefault(rows int) int {
	if rows <= 0 {
		return 3
	}
	return rows
}

// wrapperID is a deterministic id for the outer <div> so error/hint
// messages and Datastar effects can target it directly. We avoid clashing
// with the input's own id by prefixing "wrapper-".
func wrapperID(key string) string {
	return "wrapper-" + key
}

// hintID is the id used on the hint paragraph so the input can wire
// aria-describedby to it.
func hintID(key string) string {
	return "hint-" + key
}

// errorID is the id used on the error paragraph so the input can wire
// aria-describedby + aria-errormessage to it.
func errorID(key string) string {
	return "error-" + key
}
