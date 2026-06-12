// Package search provides tiny, dependency-free helpers for turning a
// raw user search term into a SQL/Ent-friendly predicate. It is shared
// across modules so every list endpoint treats "search" the same way.
//
// Typical use inside a repository:
//
//	term := strings.TrimSpace(params.Search)
//	if term != "" {
//	    q = q.Where(customer.Or(
//	        customer.NameContains(term),
//	        customer.EmailContains(term),
//	    ))
//	}
//
// This package deliberately contains no Ent/DB-specific code so it can
// be unit-tested without spinning up a database.
package search

import (
	"strings"
	"unicode"
)

// Normalize trims surrounding whitespace, collapses internal runs of
// whitespace to single spaces, and lower-cases the result. It is the
// canonical "search term" form used by every list endpoint.
//
//	Normalize("  John   DOE ") -> "john doe"
func Normalize(term string) string {
	term = strings.TrimSpace(term)
	if term == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(term))
	prevSpace := false
	for _, r := range term {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// Tokens splits a normalized term into whitespace-separated tokens.
// Useful when a module wants to match "all words" rather than the
// whole phrase.
//
//	Tokens("  john  doe ") -> []string{"john", "doe"}
func Tokens(term string) []string {
	term = Normalize(term)
	if term == "" {
		return nil
	}
	return strings.Fields(term)
}
