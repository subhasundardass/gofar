package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job is a function that runs on a schedule.
type Job struct {
	ID       string
	Module   string
	Schedule string // cron-like: "every:1h", "daily:02:00", etc. (simple impl)
	Fn       func(ctx context.Context) error
	interval time.Duration
}

// Scheduler runs registered background jobs.
type Scheduler struct {
	mu     sync.RWMutex
	jobs   []*Job
	logger *slog.Logger
	stop   chan struct{}
}

func New(logger *slog.Logger) *Scheduler {
	return &Scheduler{
		logger: logger,
		stop:   make(chan struct{}),
	}
}

// Register adds a job to the scheduler.
func (s *Scheduler) Register(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job.interval = parseInterval(job.Schedule)
	s.jobs = append(s.jobs, job)
	s.logger.Info("job registered", "id", job.ID, "module", job.Module, "schedule", job.Schedule)
}

// Start launches all registered jobs.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.RLock()
	jobs := make([]*Job, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.RUnlock()

	for _, job := range jobs {
		go s.run(ctx, job)
	}
}

// Stop signals all jobs to stop.
func (s *Scheduler) Stop() {
	close(s.stop)
}

func (s *Scheduler) run(ctx context.Context, job *Job) {
	ticker := time.NewTicker(job.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := job.Fn(ctx); err != nil {
				s.logger.Error("job failed", "id", job.ID, "error", err)
			}
		case <-s.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

func parseInterval(schedule string) time.Duration {
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
		return time.Hour
	}
}
