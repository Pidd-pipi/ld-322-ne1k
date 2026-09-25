package service

import (
	"fmt"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"log/slog"
	"time"
)

type ScheduleService struct {
	schedules *repository.ScheduleRepository
	devices   *repository.DeviceRepository
	logger    *slog.Logger
	hub       *ws.Hub
}

func NewScheduleService(schedules *repository.ScheduleRepository, devices *repository.DeviceRepository, l *slog.Logger, h *ws.Hub) *ScheduleService {
	return &ScheduleService{schedules, devices, l, h}
}
func (s *ScheduleService) Create(row *model.Schedule) (*model.Schedule, error) {
	if _, err := s.devices.Get(row.DeviceID); err != nil {
		return nil, err
	}
	if existing, err := s.schedules.FindEnabled(row.DeviceID, row.Hour, row.Minute); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, apperrors.NewScheduleConflict(fmt.Sprintf("%02d:%02d", row.Hour, row.Minute))
	}
	row.Enabled = true
	if err := s.schedules.Create(row); err != nil {
		return nil, err
	}
	s.hub.Broadcast(constants.EventSchedule, row)
	return row, nil
}
func (s *ScheduleService) List(greenhouseID uint) ([]model.Schedule, error) {
	return s.schedules.ListByGreenhouse(greenhouseID)
}
func (s *ScheduleService) ListByDevice(deviceID uint) ([]model.Schedule, error) {
	return s.schedules.ListByDevice(deviceID)
}
func (s *ScheduleService) SetEnabled(id uint, enabled bool) (*model.Schedule, error) {
	row, err := s.schedules.SetEnabled(id, enabled)
	if err == nil {
		s.hub.Broadcast(constants.EventSchedule, row)
	}
	return row, err
}
func (s *ScheduleService) RunDue(now time.Time) {
	due, err := s.schedules.ListDue(now.Hour(), now.Minute())
	if err != nil {
		s.logger.Error("list due schedules", "error", err)
		return
	}
	for _, row := range due {
		if row.LastRunAt != nil && row.LastRunAt.Truncate(time.Minute).Equal(now.Truncate(time.Minute)) {
			continue
		}
		device, err := s.devices.Toggle(row.DeviceID, row.Action, constants.OperatorSchedule)
		if err != nil {
			s.logger.Error("execute schedule", "schedule_id", row.ID, "device_id", row.DeviceID, "error", err)
			continue
		}
		if err = s.schedules.MarkExecuted(row.ID, now); err != nil {
			s.logger.Error("mark schedule executed", "schedule_id", row.ID, "error", err)
		}
		s.hub.Broadcast(constants.EventDevice, device)
		s.logger.Info("schedule executed", "schedule_id", row.ID, "device_id", row.DeviceID, "action", row.Action)
	}
}
