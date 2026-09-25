package service

import (
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"log/slog"
	"time"
)

type ControlService struct {
	repo   *repository.DeviceRepository
	logger *slog.Logger
	hub    *ws.Hub
}

func NewControlService(r *repository.DeviceRepository, l *slog.Logger, h *ws.Hub) *ControlService {
	return &ControlService{r, l, h}
}
func (s *ControlService) List(gid uint) ([]model.Device, error) { return s.repo.List(gid) }
func (s *ControlService) Toggle(id uint, status string) (*model.Device, error) {
	return s.toggle(id, status, constants.OperatorAdmin)
}
func (s *ControlService) toggle(id uint, status, operator string) (*model.Device, error) {
	d, e := s.repo.Toggle(id, status, operator)
	if e == nil {
		s.hub.Broadcast(constants.EventDevice, d)
	}
	return d, e
}
func (s *ControlService) CreateSchedule(row *model.Schedule) error {
	existing, err := s.repo.FindEnabledSchedule(row.DeviceID, row.Hour, row.Minute)
	if err != nil {
		return err
	}
	if existing != nil {
		return apperrors.ErrScheduleConflict
	}
	return s.repo.CreateSchedule(row)
}
func (s *ControlService) SetScheduleEnabled(id uint, enabled bool) (*model.Schedule, error) {
	return s.repo.SetScheduleEnabled(id, enabled)
}
func (s *ControlService) Schedules(id uint) ([]model.Schedule, error) {
	return s.repo.ListSchedules(id)
}

// RunDueSchedules 执行当前时刻到点的启用任务：切换设备状态并留下操作记录。
// 通过 last_run_at 保证同一分钟内不会重复触发。
func (s *ControlService) RunDueSchedules(now time.Time) {
	rows, err := s.repo.DueSchedules(now.Hour(), now.Minute(), now.Truncate(time.Minute))
	if err != nil {
		s.logger.Error("list due schedules", "error", err)
		return
	}
	for _, row := range rows {
		if _, err = s.toggle(row.DeviceID, row.Action, constants.OperatorSchedule); err != nil {
			s.logger.Error("execute schedule", "schedule", row.ID, "error", err)
			continue
		}
		if err = s.repo.MarkScheduleExecuted(row.ID, now); err != nil {
			s.logger.Error("mark schedule executed", "schedule", row.ID, "error", err)
		}
		s.logger.Info("schedule executed", "schedule", row.ID, "device", row.DeviceID, "action", row.Action)
	}
}
