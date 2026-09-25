package service

import (
	"errors"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"testing"
	"time"
)

func newControlTestService(t *testing.T, dbName string) (*ControlService, *repository.DeviceRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewDeviceRepository(db)
	svc := NewControlService(repo, slog.New(slog.NewTextHandler(io.Discard, nil)), ws.NewHub())
	return svc, repo, db
}

func TestCreateScheduleConflictKeepsOriginal(t *testing.T) {
	svc, repo, db := newControlTestService(t, "control_conflict_test")
	device := model.Device{GreenhouseID: 1, Name: "灌溉泵", Type: "pump", Status: "off"}
	if err := db.Create(&device).Error; err != nil {
		t.Fatal(err)
	}
	original := &model.Schedule{DeviceID: device.ID, Hour: 8, Minute: 30, Action: "on", Enabled: true}
	if err := svc.CreateSchedule(original); err != nil {
		t.Fatal(err)
	}
	duplicate := &model.Schedule{DeviceID: device.ID, Hour: 8, Minute: 30, Action: "off", Enabled: true}
	if err := svc.CreateSchedule(duplicate); !errors.Is(err, apperrors.ErrScheduleConflict) {
		t.Fatalf("want schedule conflict, got %v", err)
	}
	rows, err := repo.ListSchedules(device.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != original.ID || rows[0].Action != "on" || !rows[0].Enabled {
		t.Fatalf("original schedule not preserved: %#v", rows)
	}
	if _, err = svc.SetScheduleEnabled(original.ID, false); err != nil {
		t.Fatal(err)
	}
	rescheduled := &model.Schedule{DeviceID: device.ID, Hour: 8, Minute: 30, Action: "off", Enabled: true}
	if err = svc.CreateSchedule(rescheduled); err != nil {
		t.Fatalf("reschedule after disable should succeed: %v", err)
	}
}

func TestRunDueSchedulesTogglesDeviceAndRecordsAction(t *testing.T) {
	svc, _, db := newControlTestService(t, "control_due_test")
	device := model.Device{GreenhouseID: 1, Name: "补光灯", Type: "light", Status: "off"}
	if err := db.Create(&device).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	due := model.Schedule{DeviceID: device.ID, Hour: now.Hour(), Minute: now.Minute(), Action: "on", Enabled: true}
	disabled := model.Schedule{DeviceID: device.ID, Hour: now.Hour(), Minute: now.Minute(), Action: "off", Enabled: false}
	if err := db.Create(&due).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&disabled).Error; err != nil {
		t.Fatal(err)
	}
	svc.RunDueSchedules(now)
	var updated model.Device
	if err := db.First(&updated, device.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Status != "on" {
		t.Fatalf("want device on, got %s", updated.Status)
	}
	var actions []model.DeviceAction
	if err := db.Where("device_id=?", device.ID).Find(&actions).Error; err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Operator != "定时任务" || actions[0].Action != "on" {
		t.Fatalf("want one scheduled action record, got %#v", actions)
	}
	svc.RunDueSchedules(now)
	if err := db.Where("device_id=?", device.ID).Find(&actions).Error; err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatalf("schedule must not fire twice in the same minute, got %d actions", len(actions))
	}
}
