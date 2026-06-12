// events.go wires cross-module event subscriptions for the Base module.
// Called from Module.Boot — do not call directly.
package base

import (
	"github.com/subhasundardas/gofar/framework/event"
	"github.com/subhasundardas/gofar/framework/logger"
	"github.com/subhasundardas/gofar/modules/base/service"
)

// RegisterEvents subscribes to events published by other modules.
// Add new subscriptions here as the application grows.
// The event bus dispatches by reflect.Type — use typed structs, never strings.
//
// Example:
//
//	type OrderPlaced struct { OrderID int }
//
//	bus.Subscribe(OrderPlaced{}, func(e any) {
//	    ev := e.(OrderPlaced)
//	    svc.Base.HandleOrderPlaced(ev.OrderID)
//	})
func RegisterEvents(bus *event.Bus, svc *service.Services, log *logger.Logger) {
	// TODO: subscribe to cross-module events here.
	_ = bus
	_ = svc
	_ = log
}
