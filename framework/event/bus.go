// Package event provides a simple in-process publish/subscribe event bus.
//
// Events are plain Go structs. Handlers are called in separate goroutines so
// publishing is non-blocking. If a handler panics the panic is recovered and
// logged so it cannot bring down the server.
//
// Example:
//
//	// Define an event
//	type UserCreated struct { UserID int; Email string }
//
//	// Subscribe (returns an unsubscribe func)
//	unsub := bus.Subscribe(UserCreated{}, func(e any) {
//	    ev := e.(UserCreated)
//	    sendWelcomeEmail(ev.Email)
//	})
//	defer unsub() // clean up when no longer needed
//
//	// Publish (non-blocking, runs handler in goroutine)
//	bus.Publish(UserCreated{UserID: 1, Email: "alice@example.com"})
package event

import (
	"fmt"
	"reflect"
	"sync"
)

// Handler is a function that receives a published event. The event value is
// passed as any; cast it to the concrete type inside the handler.
type Handler func(any)

// subscription is an internal handle that allows a specific handler to be
// removed without affecting other subscribers to the same event type.
type subscription struct {
	id      uint64
	handler Handler
}

// Bus is a thread-safe in-process event dispatcher.
// Zero value is NOT usable; create via New().
type Bus struct {
	mu       sync.RWMutex
	handlers map[reflect.Type][]*subscription
	seq      uint64 // monotonically increasing subscription ID
}

// New returns a ready-to-use Bus.
func New() *Bus {
	return &Bus{
		handlers: make(map[reflect.Type][]*subscription),
	}
}

// Subscribe registers handler to be called whenever an event of the same
// dynamic type as event is published.
//
// It returns an unsubscribe function. Call it to deregister the handler, for
// example when a module shuts down:
//
//	unsub := bus.Subscribe(MyEvent{}, myHandler)
//	defer unsub()
func (b *Bus) Subscribe(event any, handler Handler) (unsubscribe func()) {
	t := reflect.TypeOf(event)

	b.mu.Lock()
	b.seq++
	id := b.seq
	sub := &subscription{id: id, handler: handler}
	b.handlers[t] = append(b.handlers[t], sub)
	b.mu.Unlock()

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.handlers[t]
		filtered := subs[:0]
		for _, s := range subs {
			if s.id != id {
				filtered = append(filtered, s)
			}
		}
		b.handlers[t] = filtered
	}
}

// Publish dispatches event to all registered handlers for its type.
// Each handler is called in its own goroutine so Publish returns immediately.
// Panics inside handlers are recovered and printed to stderr rather than
// crashing the caller.
func (b *Bus) Publish(event any) {
	t := reflect.TypeOf(event)

	b.mu.RLock()
	subs := make([]*subscription, len(b.handlers[t]))
	copy(subs, b.handlers[t])
	b.mu.RUnlock()

	for _, sub := range subs {
		h := sub.handler // capture for goroutine
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("event.Bus: handler panic for %v: %v\n", t, r)
				}
			}()
			h(event)
		}()
	}
}

// PublishSync is like Publish but runs handlers sequentially in the calling
// goroutine. Useful in tests where you need to assert side effects before
// continuing.
func (b *Bus) PublishSync(event any) {
	t := reflect.TypeOf(event)

	b.mu.RLock()
	subs := make([]*subscription, len(b.handlers[t]))
	copy(subs, b.handlers[t])
	b.mu.RUnlock()

	for _, sub := range subs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("event.Bus: handler panic for %v: %v\n", t, r)
				}
			}()
			sub.handler(event)
		}()
	}
}

// SubscriberCount returns how many handlers are registered for the event's
// type. Mainly useful in tests.
func (b *Bus) SubscriberCount(event any) int {
	t := reflect.TypeOf(event)
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[t])
}
