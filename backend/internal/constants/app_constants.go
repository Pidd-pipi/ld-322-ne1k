package constants

const (
	APIPrefix        = "/api/v1"
	HealthPath       = "/healthz"
	WebSocketPath    = "/ws"
	DefaultPage      = 1
	DefaultPageSize  = 100
	MaxPageSize      = 500
	StatusOnline     = "online"
	StatusOff        = "off"
	StatusOn         = "on"
	AlertPending     = "pending"
	AlertHandled     = "handled"
	RoleAdmin        = "admin"
	SuccessMessage   = "ok"
	EventAlert       = "alert.created"
	EventDevice      = "device.updated"
	OperatorAdmin    = "admin"
	OperatorSchedule = "定时任务"
	// ScheduleTickSeconds 是定时任务调度器的巡检间隔（秒）
	ScheduleTickSeconds = 20
)
