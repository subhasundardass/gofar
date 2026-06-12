// Package gen contains all code generation logic for the GoFar CLI.
// Each public function (MakeModule, MakeHandler, MakeService) corresponds to
// one CLI command and is responsible for writing files and updating the
// modules/all/all.go registry.
package gen

import (
	"strings"
	"unicode"
)

// Names holds every casing variant of a module name so templates never
// need to compute them inline.
//
//	Names for "BillingNotification":
//	  Pascal  = "BillingNotification"   (Go type names, file package names)
//	  Camel   = "billingNotification"   (JSON keys)
//	  Snake   = "billing_notification"  (file names, DB tables)
//	  Kebab   = "billing-notification"  (URL paths)
//	  Lower   = "billingnotification"   (Go package name — no underscores)
type Names struct {
	Pascal string // BillingNotification
	Camel  string // billingNotification
	Snake  string // billing_notification
	Kebab  string // billing-notification
	Lower  string // billingnotification  (used as Go package name)
}

// newNames derives all casing variants from a raw input string.
// Input may be PascalCase, camelCase, snake_case, or kebab-case.
func newNames(raw string) Names {
	words := splitWords(raw)
	if len(words) == 0 {
		return Names{}
	}

	// Pascal: each word Title-cased
	pascal := ""
	for _, w := range words {
		pascal += strings.Title(strings.ToLower(w)) //nolint:staticcheck
	}

	// Camel: first word lower, rest Title-cased
	camel := strings.ToLower(words[0])
	for _, w := range words[1:] {
		camel += strings.Title(strings.ToLower(w)) //nolint:staticcheck
	}

	lower := make([]string, len(words))
	for i, w := range words {
		lower[i] = strings.ToLower(w)
	}

	return Names{
		Pascal: pascal,
		Camel:  camel,
		Snake:  strings.Join(lower, "_"),
		Kebab:  strings.Join(lower, "-"),
		Lower:  strings.Join(lower, ""), // Go package names have no separator
	}
}

// splitWords splits a string on uppercase transitions, underscores, and hyphens.
// "BillingNotification" → ["Billing","Notification"]
// "billing_notification" → ["billing","notification"]
// "billing-notification" → ["billing","notification"]
func splitWords(s string) []string {
	// Normalise separators first
	s = strings.ReplaceAll(s, "-", "_")
	if strings.Contains(s, "_") {
		return strings.Split(s, "_")
	}
	// Split on uppercase transitions (PascalCase / camelCase)
	var words []string
	var current []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && len(current) > 0 {
			words = append(words, string(current))
			current = nil
		}
		current = append(current, r)
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}
	return words
}
