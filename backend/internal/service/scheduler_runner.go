package service

import (
	"context"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	"log/slog"
	"time"
)

type SchedulerRunner struct {
	schedules *ScheduleService
	logger    *slog.Logger
}

func NewSchedulerRunner(schedules *ScheduleService, l *slog.Logger) *SchedulerRunner {
	return &SchedulerRunner{schedules, l}
}
func (r *SchedulerRunner) Start(ctx context.Context) {
	ticker := time.NewTicker(constants.ScheduleTickInterval)
	defer ticker.Stop()
	r.logger.Info("schedule runner started", "tick", constants.ScheduleTickInterval.String())
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			r.schedules.RunDue(now)
		}
	}
}
