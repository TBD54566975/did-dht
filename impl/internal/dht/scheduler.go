package dht

import (
	"time"

	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
)

// Scheduler runs crob jobs asynchronously, designed to just schedule one job
type Scheduler struct {
	scheduler *gocron.Scheduler
	job       *gocron.Job
}

// NewScheduler creates a new scheduler
func NewScheduler() Scheduler {
	s := gocron.NewScheduler(time.UTC)
	s.SingletonModeAll()
	return Scheduler{scheduler: s}
}

// Schedule schedules a job to run and starts it asynchronously
func (s *Scheduler) Schedule(_ string, job func()) error {
	if s.job != nil {
		return errors.New("job already scheduled")
	}
	j, err := s.scheduler.Cron("* * * * *").Do(job)
	if err != nil {
		return err
	}
	s.job = j
	s.Start()
	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.scheduler.StartAsync()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.scheduler.Stop()
}
