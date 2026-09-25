package apperrors

import "net/http"

type BusinessError struct {
	Code    int
	Message string
	Status  int
}

func (e *BusinessError) Error() string { return e.Message }
func New(code int, message string, status int) *BusinessError {
	return &BusinessError{Code: code, Message: message, Status: status}
}

var (
	ErrNotFound     = New(40401, "资源不存在", http.StatusNotFound)
	ErrValidation   = New(40001, "请求参数不合法", http.StatusBadRequest)
	ErrUnauthorized = New(40101, "认证失败", http.StatusUnauthorized)
	ErrInternal     = New(50001, "服务器内部错误", http.StatusInternalServerError)
)

const ScheduleConflictCode = 40901

func NewScheduleConflict(detail string) *BusinessError {
	return New(ScheduleConflictCode, "该设备在 "+detail+" 已有启用的定时任务，原任务已保留；如需重新安排请先停用原任务", http.StatusConflict)
}
