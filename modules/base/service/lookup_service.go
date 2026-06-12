package service

import (
	"fmt"

	"github.com/subhasundardas/gofar/modules/base/lookup"
)

// LookupService is the public, container-friendly face of the shared
// lookup registry. Domain modules resolve it and call Register to add
// their own autocomplete providers.
type LookupService struct {
	registry *lookup.Registry
}

// NewLookupService constructs a LookupService over the given registry.
func NewLookupService(reg *lookup.Registry) *LookupService {
	if reg == nil {
		reg = lookup.NewRegistry()
	}
	return &LookupService{registry: reg}
}

// Register adds a provider. It is a thin pass-through kept here so the
// service interface is the only thing other modules need to depend on.
func (s *LookupService) Register(p lookup.Provider) error {
	return s.registry.Register(p)
}

// Search runs the named provider with a built Query. Returns an error
// if the provider is not found.
func (s *LookupService) Search(name string, q lookup.Query) ([]lookup.Hit, error) {
	p := s.registry.Get(name)
	if p == nil {
		return nil, fmt.Errorf("lookup: provider %q not registered", name)
	}
	return p.Search(q)
}

// Names returns all registered provider names. Useful for diagnostics.
func (s *LookupService) Names() []string {
	return s.registry.Names()
}

// Raw returns the underlying registry. Reserved for tests and tools —
// domain modules should not need this.
func (s *LookupService) Raw() *lookup.Registry { return s.registry }
