// Package container provides a lightweight inversion-of-control (IoC) service
// container for GoFar. Services are keyed by their reflect.Type, meaning each
// concrete type can be registered once. This design keeps things simple and
// fast – there is no runtime overhead beyond a single map lookup on Resolve.
//
// Typical usage:
//
//	c := container.New()
//	c.Register(myService)          // register by concrete type
//	svc, err := container.Resolve[*MyService](c)  // retrieve typed
//
// Thread-safety: Register and Resolve are protected by a read/write mutex and
// are therefore safe to call from concurrent goroutines.
package container

import (
	"fmt"
	"reflect"
	"sync"
)

// Container is a type-safe service registry. Zero value is not usable; always
// create via New().
type Container struct {
	mu       sync.RWMutex
	services map[reflect.Type]any
}

// New returns an initialised, empty Container ready for registrations.
func New() *Container {
	return &Container{
		services: make(map[reflect.Type]any),
	}
}

// Register stores service in the container keyed by its concrete reflect.Type.
// If a service with the same type is already registered it is silently
// overwritten – this allows tests and bootstrap overrides to swap
// implementations without error.
//
// Panics if service is nil.
func (c *Container) Register(service any) {
	if service == nil {
		panic("container: cannot register nil service")
	}
	t := reflect.TypeOf(service)
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.services[t]; exists {
		panic("container: service already registered: " + t.String())
	}

	c.services[t] = service
}

// RegisterAs stores service under the interface type I rather than its
// concrete type. This is the preferred way to register services that are used
// through an interface:
//
//	container.RegisterAs[contracts.Logger](c, &myLogger{})
//
// The type parameter I must be an interface; calling this with a concrete type
// is identical to Register.
func RegisterAs[I any](c *Container, service I) {
	t := reflect.TypeOf((*I)(nil)).Elem()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[t] = service
}

// Resolve retrieves a previously registered service of type T. It returns an
// error if no service of that type has been registered.
//
//	fiber, err := container.Resolve[*fiber.App](c)
func Resolve[T any](c *Container) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()
	c.mu.RLock()
	defer c.mu.RUnlock()
	svc, ok := c.services[t]
	if !ok {
		return zero, fmt.Errorf("container: service not registered for type %v", t)
	}
	typed, ok := svc.(T)
	if !ok {
		return zero, fmt.Errorf("container: type assertion failed for %v", t)
	}
	return typed, nil
}

// MustResolve is like Resolve but panics instead of returning an error.
// Useful in bootstrap code where a missing service is always a programming
// error.
func MustResolve[T any](c *Container) T {
	svc, err := Resolve[T](c)
	if err != nil {
		panic(err)
	}
	return svc
}

// Has reports whether a service of type T is currently registered.
func Has[T any](c *Container) bool {
	t := reflect.TypeOf((*T)(nil)).Elem()
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.services[t]
	return ok
}
