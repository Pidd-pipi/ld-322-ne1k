package handler

import (
	"github.com/cygreenenv/greenhouse-panel/internal/dto"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ScheduleHandler struct {
	service   *service.ScheduleService
	validator *validator.Validate
}

func NewScheduleHandler(s *service.ScheduleService, v *validator.Validate) *ScheduleHandler {
	return &ScheduleHandler{s, v}
}
func (h *ScheduleHandler) Create(c *gin.Context) {
	var req dto.ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil {
		Fail(c, apperrors.ErrValidation)
		return
	}
	row := &model.Schedule{DeviceID: req.DeviceID, Hour: req.Hour, Minute: req.Minute, Action: req.Action}
	created, err := h.service.Create(row)
	if err != nil {
		Fail(c, err)
		return
	}
	Created(c, created)
}
func (h *ScheduleHandler) List(c *gin.Context) {
	id, ok := queryID(c, "greenhouse_id")
	if !ok {
		return
	}
	rows, err := h.service.List(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, rows)
}
func (h *ScheduleHandler) ListByDevice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	rows, err := h.service.ListByDevice(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, rows)
}
func (h *ScheduleHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.ScheduleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil {
		Fail(c, apperrors.ErrValidation)
		return
	}
	row, err := h.service.SetEnabled(id, *req.Enabled)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, row)
}
