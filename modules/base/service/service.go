// Package service contains the cross-cutting services of the Base module.
//
// The Base module is shared utilities — pagination, search, lookup — so
// this file only wires those primitives together. It does NOT define
// domain CRUD on a "Base" entity.
package service

import (
	"github.com/subhasundardas/gofar/framework/logger"
	"github.com/subhasundardas/gofar/modules/base/repository"
)

// Services bundles all Base services exposed to other modules.
type Services struct {
	// Lookup exposes a thin wrapper around the shared lookup registry.
	// Other modules resolve this from the container to register their
	// own lookup providers during their Register phase.
	Lookup *LookupService
}

// NewServices constructs all services, injecting the shared repository group.
func NewServices(repos *repository.Repositories, logger *logger.Logger) *Services {
	return &Services{
		Lookup: NewLookupService(repos.Lookup),
	}
}
