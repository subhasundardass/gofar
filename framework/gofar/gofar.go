// Package gofar is the single import for all framework features inside a module.
//
// Every module imports only this package — no need to know where
// individual framework packages live.
//
// Usage inside Register or Boot:
//
//	import fw "github.com/subhasundardas/gofar/framework/gofar"
//
//	func (m *Module) Register(mgr *module.Manager) error {
//	    gofar := fw.Use(mgr)
//	    svc   := gofar.Services
//
//	    // framework dependencies
//	    db     := svc.DB()
//	    log    := svc.Logger()
//	    events := svc.Events()
//	    cfg    := svc.Config()
//	    store  := svc.Store()
//	    fiber  := svc.Fiber()
//
//	    // BaseService helpers — directly on svc
//	    svc.WrapError("op", err)
//	    svc.ValidateRequired(map[string]string{"name": name})
//	    svc.ValidatePositive("amount", 10.0)
//	    svc.ValidateMinLength("password", val, 8)
//	    svc.ValidateMaxLength("bio", val, 500)
//	    svc.Paginate(ctx, params, repo.List)
//
//	    // inject BaseService into module services
//	    m.services = service.NewServices(m.repos, svc.BaseService)
//	    return nil
//	}
//
// Stateless packages — import directly, no mgr needed:
//
//	"github.com/subhasundardas/gofar/framework/form"
//	"github.com/subhasundardas/gofar/framework/datastar"
//	"github.com/subhasundardas/gofar/framework/errors"
//	"github.com/subhasundardas/gofar/framework/render"
//	"github.com/subhasundardas/gofar/framework/request"
//	"github.com/subhasundardas/gofar/framework/data"
package gofar

import (
	"github.com/subhasundardas/gofar/framework/module"
	"github.com/subhasundardas/gofar/framework/services"
)

// Framework is the gofar accessor. Obtain one via Use(mgr).
// It groups every framework feature under a single variable.
type Framework struct {
	// Services gives access to all framework dependencies and BaseService helpers.
	// DB, Logger, Events, Config, Store, Fiber, WrapError, Validate*, Paginate.
	Services *services.Registry
}

// Use builds the gofar Framework accessor from a module Manager.
// Call this once at the top of Register or Boot:
//
//	gofar := fw.Use(mgr)
//	svc   := gofar.Services
//	db    := svc.DB()
//	log   := svc.Logger()
func Use(mgr *module.Manager) *Framework {
	return &Framework{
		Services: mgr.Services(),
	}
}
