package constants

import "time"

const (
	APIPrefix            = "/api/v1"
	HealthPath           = "/healthz"
	WebSocketPath        = "/ws"
	DefaultPage          = 1
	DefaultPageSize      = 100
	MaxPageSize          = 500
	StatusOnline         = "online"
	StatusOff            = "off"
	StatusOn             = "on"
	AlertPending         = "pending"
	AlertHandled         = "handled"
	RoleAdmin            = "admin"
	SuccessMessage       = "ok"
	EventAlert           = "alert.created"
	EventDevice          = "device.updated"
	EventSchedule        = "schedule.updated"
	OperatorAdmin        = "admin"
	OperatorSchedule     = "定时任务"
	ScheduleTickInterval = 10 * time.Second
)
