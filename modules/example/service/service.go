// Package service contains business logic for the Example module.
// Services sit between handlers and repositories — enforce domain rules here.
package service

import (
	"strconv"
	"sync"

	"github.com/subhasundardas/gofar/modules/example/repository"
)

// Services bundles all Example domain services.
type Services struct {
	Counter *Counter
}

// NewServices constructs all services, injecting the shared repository group.
func NewServices(repos *repository.Repositories) *Services {
	return &Services{
		Counter: NewCounter(),
	}
}

// Counter is a tiny, process-local click counter used by the datastar
// demo. It is intentionally NOT a repository-backed entity — the goal
// is to show the SSE round-trip, not persistence.
type Counter struct {
	mu  sync.Mutex
	val int
}

// NewCounter returns a Counter starting at 0.
func NewCounter() *Counter {
	return &Counter{}
}

// Current returns the current value formatted as a string (suitable
// for direct injection into a templ fragment).
func (c *Counter) Current() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return strconv.Itoa(c.val)
}

// Increment atomically adds 1 and returns the new value as a string.
func (c *Counter) Increment() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.val++
	return strconv.Itoa(c.val)
}

func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.val = 0
}
