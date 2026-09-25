package repository

import (
	"errors"
	"fmt"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
	"time"
)

type DeviceRepository struct{ db *gorm.DB }

func NewDeviceRepository(db *gorm.DB) *DeviceRepository { return &DeviceRepository{db: db} }
func (r *DeviceRepository) List(greenhouseID uint) ([]model.Device, error) {
	var rows []model.Device
	if err := r.db.Where("greenhouse_id=?", greenhouseID).Order("id asc").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	return rows, nil
}
func (r *DeviceRepository) Toggle(id uint, status, operator string) (*model.Device, error) {
	var row model.Device
	err := r.db.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	row.Status = status
	if err = r.db.Save(&row).Error; err != nil {
		return nil, fmt.Errorf("save device: %w", err)
	}
	if err = r.db.Create(&model.DeviceAction{DeviceID: id, Action: status, Operator: operator}).Error; err != nil {
		return nil, fmt.Errorf("create device action: %w", err)
	}
	return &row, nil
}
func (r *DeviceRepository) CreateSchedule(row *model.Schedule) error {
	if err := r.db.Create(row).Error; err != nil {
		return fmt.Errorf("create schedule: %w", err)
	}
	return nil
}
func (r *DeviceRepository) FindEnabledSchedule(deviceID uint, hour, minute int) (*model.Schedule, error) {
	var row model.Schedule
	err := r.db.Where("device_id=? AND hour=? AND minute=? AND enabled=?", deviceID, hour, minute, true).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find enabled schedule: %w", err)
	}
	return &row, nil
}
func (r *DeviceRepository) SetScheduleEnabled(id uint, enabled bool) (*model.Schedule, error) {
	var row model.Schedule
	err := r.db.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	row.Enabled = enabled
	if err = r.db.Save(&row).Error; err != nil {
		return nil, fmt.Errorf("save schedule: %w", err)
	}
	return &row, nil
}
func (r *DeviceRepository) DueSchedules(hour, minute int, minuteStart time.Time) ([]model.Schedule, error) {
	var rows []model.Schedule
	err := r.db.Where("enabled=? AND hour=? AND minute=? AND (last_run_at IS NULL OR last_run_at < ?)", true, hour, minute, minuteStart).Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list due schedules: %w", err)
	}
	return rows, nil
}
func (r *DeviceRepository) MarkScheduleExecuted(id uint, at time.Time) error {
	if err := r.db.Model(&model.Schedule{}).Where("id=?", id).Update("last_run_at", at).Error; err != nil {
		return fmt.Errorf("mark schedule executed: %w", err)
	}
	return nil
}
func (r *DeviceRepository) ListSchedules(deviceID uint) ([]model.Schedule, error) {
	var rows []model.Schedule
	if err := r.db.Where("device_id=?", deviceID).Order("hour asc, minute asc, id asc").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	return rows, nil
}
