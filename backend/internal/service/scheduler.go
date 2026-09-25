package service

import (
	"context"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	"time"
)

// StartScheduler 启动定时任务调度器，按固定间隔巡检到点的启用任务并执行。
func (s *ControlService) StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(constants.ScheduleTickSeconds) * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				s.RunDueSchedules(now)
			}
		}
	}()
}
