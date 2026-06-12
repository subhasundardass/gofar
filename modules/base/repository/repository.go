// Package repository is the data-access layer for the Base module.
//
// The Base module is a *cross-cutting* module: it does not own a domain
// entity. This file is therefore intentionally minimal — it only exposes
// registry hooks for shared cross-module repositories (e.g. the Lookup
// registry that backs the generic autocomplete used by formui/lookup).
package repository

import (
	"github.com/subhasundardas/gofar/ent"
	"github.com/subhasundardas/gofar/modules/base/lookup"
)

// Repositories bundles all cross-cutting Base repositories.
// Pass the whole struct to service.NewServices so adding a shared repo
// never changes signatures.
type Repositories struct {
	// Lookup is the shared autocomplete/lookup registry. Modules register
	// their own lookup providers on it; formui/lookup reads from it.
	Lookup *lookup.Registry
}

// NewRepositories constructs all shared repositories.
func NewRepositories(*ent.Client) *Repositories {
	return &Repositories{
		Lookup: lookup.NewRegistry(),
	}
}
