package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cygreenenv/greenhouse-panel/internal/config"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type apiEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func newScheduleTestEngine(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:router_schedule_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	greenhouse := model.Greenhouse{Name: "接口温室"}
	if err = db.Create(&greenhouse).Error; err != nil {
		t.Fatal(err)
	}
	device := model.Device{GreenhouseID: greenhouse.ID, Name: "补光灯", Type: "light", Status: "off"}
	if err = db.Create(&device).Error; err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := ws.NewHub()
	greenhouses := repository.NewGreenhouseRepository(db)
	sensors := repository.NewSensorRepository(db)
	alerts := repository.NewAlertRepository(db)
	devices := repository.NewDeviceRepository(db)
	schedules := repository.NewScheduleRepository(db)
	auth := service.NewAuthService("router-test-secret")
	engine := New(Dependencies{
		Config:     config.Config{JWTSecret: "router-test-secret"},
		Logger:     logger,
		Auth:       auth,
		Monitoring: service.NewMonitoringService(greenhouses, sensors, alerts, logger, hub),
		Alerts:     service.NewAlertService(alerts, logger),
		Control:    service.NewControlService(devices, logger, hub),
		Schedules:  service.NewScheduleService(schedules, devices, logger, hub),
		Reports:    service.NewReportService(sensors, alerts),
		Hub:        hub,
	})
	token, err := auth.Login("admin", "admin123")
	if err != nil {
		t.Fatal(err)
	}
	return engine, token
}

func callAPI(t *testing.T, engine *gin.Engine, method string, path string, token string, body any) (int, apiEnvelope) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	var envelope apiEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response %s: %v", recorder.Body.String(), err)
	}
	return recorder.Code, envelope
}

func TestScheduleLifecycleEndpoints(t *testing.T) {
	engine, token := newScheduleTestEngine(t)
	payload := gin.H{"deviceId": 1, "hour": 8, "minute": 30, "action": "on"}
	status, _ := callAPI(t, engine, http.MethodPost, "/api/v1/schedules", token, payload)
	if status != http.StatusCreated {
		t.Fatalf("create schedule: want 201, got %d", status)
	}
	status, envelope := callAPI(t, engine, http.MethodPost, "/api/v1/schedules", token, payload)
	if status != http.StatusConflict || envelope.Code != 40901 {
		t.Fatalf("duplicate create: want 409/40901, got %d/%d", status, envelope.Code)
	}
	status, envelope = callAPI(t, engine, http.MethodGet, "/api/v1/schedules?greenhouse_id=1", "", nil)
	if status != http.StatusOK {
		t.Fatalf("list schedules: want 200, got %d", status)
	}
	var listed []model.Schedule
	if err := json.Unmarshal(envelope.Data, &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || !listed[0].Enabled || listed[0].Device == nil || listed[0].Device.Name != "补光灯" {
		t.Fatalf("unexpected schedule list: %s", envelope.Data)
	}
	status, _ = callAPI(t, engine, http.MethodPatch, fmt.Sprintf("/api/v1/schedules/%d/status", listed[0].ID), token, gin.H{"enabled": false})
	if status != http.StatusOK {
		t.Fatalf("disable schedule: want 200, got %d", status)
	}
	status, _ = callAPI(t, engine, http.MethodPost, "/api/v1/schedules", token, payload)
	if status != http.StatusCreated {
		t.Fatalf("recreate after disable: want 201, got %d", status)
	}
	status, _ = callAPI(t, engine, http.MethodPost, "/api/v1/schedules", "", payload)
	if status != http.StatusUnauthorized {
		t.Fatalf("create without token: want 401, got %d", status)
	}
}
