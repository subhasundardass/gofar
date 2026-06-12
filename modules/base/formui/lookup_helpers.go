package formui

import (
	"encoding/json"
	"fmt"

	"github.com/subhasundardas/gofar/framework/form"
)

// ── Lookup-field helpers ─────────────────────────────────────────────────────
//
// The lookup combobox stores its data in three Datastar signals:
//
//   - search: the text in the visible input box
//   - open:   boolean — whether the dropdown is shown
//   - value:  the option's value, written to the hidden form input
//   - items:  the full options array (as JSON) used by the filter expression
//
// We need three helpers that produce attribute values for templ:
//
//   - getInitialSearchLabel returns the option whose Value == field.Value,
//     or "" if no match. This is what populates the search box on first
//     render so a previously selected value is visible.
//   - getInitialItemsJSON returns the options array as a JSON string,
//     safe to embed in a data-signals attribute.
//   - lookupClearClick and lookupPickClick return the JS source for the
//     data-on:click attribute on the ✕ and option buttons respectively.
//
// All string returns are trusted (built by us) but the option Label and
// Value still need to survive embedding inside a JS source string. We
// use json.Marshal for items, and a hand-rolled escape for the small
// click-handler strings.

// getInitialSearchLabel returns the label of the option whose Value matches
// the field's current value, or "" if none does.
func getInitialSearchLabel(field form.UIField) string {
	for _, opt := range field.Meta.Options {
		if opt.Value == field.Value {
			return opt.Label
		}
	}
	return ""
}

// getInitialItemsJSON returns the options as a JSON string. Empty options
// return "[]".
func getInitialItemsJSON(field form.UIField) string {
	if len(field.Meta.Options) == 0 {
		return "[]"
	}
	b, err := json.Marshal(field.Meta.Options)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// escapeForJSSingleQuoteString escapes a Go string for safe embedding
// inside a single-quoted JS source literal. It handles backslashes and
// single quotes only — anything more exotic (newlines, control chars)
// would warrant a more careful encoder, but option values and keys are
// short human-readable strings in practice.
func escapeForJSSingleQuoteString(s string) string {
	out := make([]byte, 0, len(s)+8)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' || c == '\'' {
			out = append(out, '\\', c)
			continue
		}
		out = append(out, c)
	}
	return string(out)
}

// lookupClearClick returns the JS for the ✕ button. It clears the search
// box, the value, and closes the dropdown — no string interpolation of
// the key is needed because we use a constant id "hidden-{key}".
func lookupClearClick(key string) string {
	hiddenID := escapeForJSSingleQuoteString("hidden-" + key)
	return fmt.Sprintf(
		"search = ''; value = ''; open = false; "+
			"const el = document.getElementById('%s'); "+
			"if (el) { el.value = ''; el.dispatchEvent(new Event('input', { bubbles: true })); }",
		hiddenID,
	)
}

// lookupPickClick returns the JS for an option button. It writes the
// option's label into the search box, the option's value into the value
// signal, closes the dropdown, and dispatches an input event on the
// hidden form input so any bound Datastar effect fires.
func lookupPickClick(key string, opt form.Option) string {
	hiddenID := escapeForJSSingleQuoteString("hidden-" + key)
	label := escapeForJSSingleQuoteString(opt.Label)
	value := escapeForJSSingleQuoteString(opt.Value)
	return fmt.Sprintf(
		"search = '%s'; value = '%s'; open = false; "+
			"const el = document.getElementById('%s'); "+
			"if (el) { el.value = '%s'; el.dispatchEvent(new Event('input', { bubbles: true })); }",
		label,
		value,
		hiddenID,
		value,
	)
}
