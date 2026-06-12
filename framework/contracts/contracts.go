// Package contracts defines the core interfaces used across GoFar.
//
// Programming to interfaces rather than concrete types keeps modules loosely
// coupled and makes unit testing straightforward – swap any implementation
// with a lightweight fake that satisfies the same interface.
package contracts

import (
	"context"
	"io"
)

// Application is the top-level interface for the GoFar application.
// Modules receive a *app.App value that satisfies this interface.
type Application interface {
	// Run starts the HTTP server on addr and blocks until stopped.
	Run(addr string) error
	// Shutdown gracefully stops the server and releases resources.
	Shutdown() error
}

// Cache is the interface that any caching backend must satisfy.
// The default in-memory implementation lives in framework/cache.
type Cache interface {
	// Get retrieves a cached value. Returns (nil, false) on a cache miss.
	Get(ctx context.Context, key string) ([]byte, bool)
	// Set stores value under key with an optional TTL in seconds (0 = no expiry).
	Set(ctx context.Context, key string, value []byte, ttlSec int) error
	// Delete removes a key from the cache.
	Delete(ctx context.Context, key string) error
	// Flush removes all keys from the cache.
	Flush(ctx context.Context) error
}

// Mailer is the interface for sending email.
// Implementations can wrap SMTP, SendGrid, SES, etc.
type Mailer interface {
	// Send sends a single email message.
	Send(ctx context.Context, msg *Mail) error
}

// Mail represents an outgoing email message.
type Mail struct {
	To      []string
	Subject string
	HTML    string
	Text    string
}

// Storage is the interface for file/object storage.
// Implementations can wrap local disk, S3, GCS, etc.
type Storage interface {
	// Put stores r under key and returns the public URL or path.
	Put(ctx context.Context, key string, r io.Reader) (string, error)
	// Get retrieves the file at key.
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes the file at key.
	Delete(ctx context.Context, key string) error
}

// Logger is the minimal logging interface used by framework packages.
// Use this in package signatures rather than the concrete *logger.Logger so
// callers can inject any compatible logger.
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...any)
	Info(msg string)
	Infof(format string, args ...any)
	Warn(msg string)
	Warnf(format string, args ...any)
	Error(msg string)
	Errorf(format string, args ...any)
}
