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

func newScheduleTestService(t *testing.T, dsn string) (*ScheduleService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	devices := repository.NewDeviceRepository(db)
	svc := NewScheduleService(repository.NewScheduleRepository(db), devices, slog.New(slog.NewTextHandler(io.Discard, nil)), ws.NewHub())
	return svc, db
}

func seedScheduleDevice(t *testing.T, db *gorm.DB) model.Device {
	t.Helper()
	g := model.Greenhouse{Name: "定时温室"}
	if err := db.Create(&g).Error; err != nil {
		t.Fatal(err)
	}
	device := model.Device{GreenhouseID: g.ID, Name: "灌溉泵", Type: "pump", Status: "off"}
	if err := db.Create(&device).Error; err != nil {
		t.Fatal(err)
	}
	return device
}

func TestScheduleCreateConflictAndRecreate(t *testing.T) {
	svc, db := newScheduleTestService(t, "file:schedule_conflict_test?mode=memory&cache=shared")
	device := seedScheduleDevice(t, db)
	first, err := svc.Create(&model.Schedule{DeviceID: device.ID, Hour: 8, Minute: 30, Action: "on"})
	if err != nil {
		t.Fatal(err)
	}
	if !first.Enabled {
		t.Fatal("new schedule should be enabled")
	}
	_, err = svc.Create(&model.Schedule{DeviceID: device.ID, Hour: 8, Minute: 30, Action: "off"})
	var business *apperrors.BusinessError
	if !errors.As(err, &business) || business.Code != apperrors.ScheduleConflictCode {
		t.Fatalf("want schedule conflict error, got %v", err)
	}
	kept, err := svc.ListByDevice(device.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(kept) != 1 || kept[0].ID != first.ID || kept[0].Action != "on" {
		t.Fatalf("original schedule should be kept untouched, got %#v", kept)
	}
	if _, err = svc.SetEnabled(first.ID, false); err != nil {
		t.Fatal(err)
	}
	second, err := svc.Create(&model.Schedule{DeviceID: device.ID, Hour: 8, Minute: 30, Action: "off"})
	if err != nil {
		t.Fatalf("recreate after disable should succeed: %v", err)
	}
	if second.ID == first.ID {
		t.Fatal("recreate should insert a new schedule")
	}
}

func TestScheduleRunDueTogglesDeviceOnce(t *testing.T) {
	svc, db := newScheduleTestService(t, "file:schedule_run_test?mode=memory&cache=shared")
	device := seedScheduleDevice(t, db)
	now := time.Now()
	if _, err := svc.Create(&model.Schedule{DeviceID: device.ID, Hour: now.Hour(), Minute: now.Minute(), Action: "on"}); err != nil {
		t.Fatal(err)
	}
	svc.RunDue(now)
	updated, err := repository.NewDeviceRepository(db).Get(device.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "on" {
		t.Fatalf("device should be switched on, got %s", updated.Status)
	}
	var actions []model.DeviceAction
	if err = db.Where("device_id=?", device.ID).Find(&actions).Error; err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Operator != "定时任务" || actions[0].Action != "on" {
		t.Fatalf("want one schedule operation record, got %#v", actions)
	}
	schedules, err := svc.ListByDevice(device.ID)
	if err != nil {
		t.Fatal(err)
	}
	if schedules[0].LastRunAt == nil {
		t.Fatal("last run time should be recorded")
	}
	svc.RunDue(now)
	if err = db.Where("device_id=?", device.ID).Find(&actions).Error; err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatalf("schedule must not run twice in the same minute, got %d records", len(actions))
	}
}
