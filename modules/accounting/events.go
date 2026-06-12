// events.go wires cross-module event subscriptions for the Accounting module.
// Called from Module.Boot — do not call directly.
package accounting

import (
	"github.com/subhasundardas/gofar/framework/event"
	"github.com/subhasundardas/gofar/framework/logger"
	"github.com/subhasundardas/gofar/modules/accounting/service"
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
//	    svc.Accounting.HandleOrderPlaced(ev.OrderID)
//	})
func RegisterEvents(bus *event.Bus, svc *service.Services, log *logger.Logger) {
	// TODO: subscribe to cross-module events here.
	_ = bus
	_ = svc
	_ = log
}
