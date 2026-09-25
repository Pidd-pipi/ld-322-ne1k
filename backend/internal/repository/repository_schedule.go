package repository

import (
	"errors"
	"fmt"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
	"time"
)

type ScheduleRepository struct{ db *gorm.DB }

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository { return &ScheduleRepository{db: db} }
func (r *ScheduleRepository) Create(row *model.Schedule) error {
	if err := r.db.Create(row).Error; err != nil {
		return fmt.Errorf("create schedule: %w", err)
	}
	return nil
}
func (r *ScheduleRepository) Get(id uint) (*model.Schedule, error) {
	var row model.Schedule
	err := r.db.Preload("Device").First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	return &row, nil
}
func (r *ScheduleRepository) ListByGreenhouse(greenhouseID uint) ([]model.Schedule, error) {
	var rows []model.Schedule
	err := r.db.Preload("Device").
		Joins("JOIN devices ON devices.id = schedules.device_id").
		Where("devices.greenhouse_id = ?", greenhouseID).
		Order("schedules.hour asc, schedules.minute asc, schedules.id asc").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list schedules by greenhouse: %w", err)
	}
	return rows, nil
}
func (r *ScheduleRepository) ListByDevice(deviceID uint) ([]model.Schedule, error) {
	var rows []model.Schedule
	if err := r.db.Where("device_id=?", deviceID).Order("hour asc, minute asc, id asc").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	return rows, nil
}
func (r *ScheduleRepository) FindEnabled(deviceID uint, hour int, minute int) (*model.Schedule, error) {
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
func (r *ScheduleRepository) SetEnabled(id uint, enabled bool) (*model.Schedule, error) {
	row, err := r.Get(id)
	if err != nil {
		return nil, err
	}
	if err = r.db.Model(row).Update("enabled", enabled).Error; err != nil {
		return nil, fmt.Errorf("update schedule enabled: %w", err)
	}
	row.Enabled = enabled
	return row, nil
}
func (r *ScheduleRepository) ListDue(hour int, minute int) ([]model.Schedule, error) {
	var rows []model.Schedule
	if err := r.db.Where("enabled=? AND hour=? AND minute=?", true, hour, minute).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list due schedules: %w", err)
	}
	return rows, nil
}
func (r *ScheduleRepository) MarkExecuted(id uint, at time.Time) error {
	if err := r.db.Model(&model.Schedule{}).Where("id=?", id).Update("last_run_at", at).Error; err != nil {
		return fmt.Errorf("mark schedule executed: %w", err)
	}
	return nil
}
