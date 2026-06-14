// Package scheduler provides a lightweight job scheduler for running background tasks
// at regular intervals. Jobs are executed concurrently using goroutines and can be
// stopped gracefully.
//
// Supported schedule formats:
//   - "every:1m"  → every 1 minute
//   - "every:5m"  → every 5 minutes
//   - "every:1h"  → every 1 hour
//   - "daily"     → every 24 hours (runs at current time each day)
//
// Usage example:
//
//	scheduler := scheduler.New(logger)
//
//	job := scheduler.NewJob("daily-cleanup", "mymodule", "daily", func(ctx context.Context) error {
//	    return mymodule.CleanupOldData(ctx)
//	})
//	scheduler.Register(job)
//	scheduler.Start(ctx)
//
//	// Later...
//	scheduler.Stop()
package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job represents a background task scheduled to run at regular intervals.
type Job struct {
	// ID is a unique identifier for this job (e.g., "daily-cleanup").
	ID string

	// Module is the module that registered this job (e.g., "auth", "tasks").
	Module string

	// Schedule is the job's execution interval in the format "every:Xm", "every:Xh", or "daily".
	// Unknown formats default to 1 hour with a warning logged.
	Schedule string

	// Fn is the function to execute. It receives a context that will be cancelled
	// when the scheduler stops or the context is cancelled externally.
	// Fn should return an error if the job fails; errors are logged but do not stop
	// the scheduler.
	Fn func(ctx context.Context) error

	// interval is the parsed duration between job executions (computed from Schedule).
	interval time.Duration
}

// Scheduler runs registered background jobs concurrently at their specified intervals.
// All operations are thread-safe. Scheduler is safe for concurrent use after creation.
type Scheduler struct {
	mu       sync.RWMutex
	jobs     []*Job
	logger   *slog.Logger
	stop     chan struct{}
	stopOnce sync.Once // ensures Stop() can only close the channel once
}

// New creates a new Scheduler with the given logger.
func New(logger *slog.Logger) *Scheduler {
	return &Scheduler{
		logger: logger,
		stop:   make(chan struct{}),
	}
}

// Register adds a job to the scheduler. The job will not start executing until
// Start() is called. Register is thread-safe and can be called before or after Start().
// Jobs registered after Start() is called will begin executing on their schedule.
func (s *Scheduler) Register(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job.interval = parseInterval(job.Schedule, s.logger)
	s.jobs = append(s.jobs, job)
	s.logger.Info("job registered",
		"id", job.ID,
		"module", job.Module,
		"schedule", job.Schedule,
		"interval", job.interval.String(),
	)
}

// Start launches all registered jobs in separate goroutines. Each job executes
// on its configured interval. Start should only be called once; subsequent calls
// are safe but will spawn duplicate goroutines for jobs already running.
//
// The provided context can be used to cancel all jobs; when ctx.Done() is signalled,
// all job goroutines will exit gracefully.
//
// This method returns immediately; jobs run asynchronously in the background.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.RLock()
	// Snapshot the current jobs to avoid holding the read lock during goroutine creation
	jobs := make([]*Job, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.RUnlock()

	for _, job := range jobs {
		go s.run(ctx, job)
	}
}

// Stop signals all running jobs to stop gracefully. It is safe to call Stop()
// multiple times; only the first call will have an effect. After Stop() returns,
// no further job executions will occur.
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
}

// run executes a job repeatedly on its interval until the scheduler is stopped
// or the context is cancelled. This method runs in its own goroutine and should
// not be called directly.
func (s *Scheduler) run(ctx context.Context, job *Job) {
	ticker := time.NewTicker(job.interval)
	defer ticker.Stop()

	s.logger.Debug("job started", "id", job.ID, "interval", job.interval.String())

	for {
		select {
		case <-ticker.C:
			// Execute the job and log any errors; a single failure does not stop the scheduler
			if err := job.Fn(ctx); err != nil {
				s.logger.Error("job failed",
					"id", job.ID,
					"module", job.Module,
					"error", err,
				)
			}

		case <-s.stop:
			// Scheduler was stopped; exit this job's goroutine
			s.logger.Debug("job stopped", "id", job.ID)
			return

		case <-ctx.Done():
			// External context was cancelled; exit this job's goroutine
			s.logger.Debug("job cancelled", "id", job.ID, "reason", ctx.Err())
			return
		}
	}
}

// parseInterval converts a schedule string to a time.Duration. Returns a sensible
// default (1 hour) if the format is unrecognized.
//
// Supported formats:
//   - "every:1m", "every:5m" → minutes
//   - "every:1h" → hours
//   - "daily" → 24 hours
func parseInterval(schedule string, logger *slog.Logger) time.Duration {
	switch schedule {
	case "every:1m":
		return time.Minute
	case "every:5m":
		return 5 * time.Minute
	case "every:1h":
		return time.Hour
	case "daily":
		return 24 * time.Hour
	default:
		logger.Warn("unknown schedule format; defaulting to 1 hour",
			"schedule", schedule,
			"supported", []string{"every:1m", "every:5m", "every:1h", "daily"},
		)
		return time.Hour
	}
}
