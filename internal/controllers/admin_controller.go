package controllers

import (
	"modules/internal/models"
	"modules/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	parkingService *services.ParkingService
	reportService  *services.ReportService
}

func NewAdminController(ps *services.ParkingService, rs *services.ReportService) *AdminController {
	return &AdminController{
		parkingService: ps,
		reportService:  rs,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateSpotStatusRequest 修改车位状态请求结构
type UpdateSpotStatusRequest struct {
	Status models.ParkingStatus `json:"status" binding:"required,oneof=idle occupied faulty"` // 车位状态
	Notes  string               `json:"notes"`                                                // 备注说明
}

// ParkingSpotResponse 车位返回结构
type ParkingSpotResponse struct {
	ID         uint    `json:"id"`
	Type       string  `json:"type"`
	Status     string  `json:"status"`
	HourlyRate float64 `json:"hourly_rate"`
}

// SystemStatsResponse 系统统计响应结构
type SystemStatsResponse struct {
	TotalSpots       int                `json:"total_spots"`
	AvailableSpots   int                `json:"available_spots"`
	UtilizationRates map[string]float64 `json:"utilization_rates"`
}

// UpdateSpotStatus 更新车位状态
// @Summary 更新车位状态
// @Description 管理员根据车位ID修改车位状态（空闲/占用/故障）
// @Tags admin
// @Accept json
// @Produce json
// @Example {"status": "occupied", "notes": "车辆停放中"}
// @Param id path int true "车位ID"
// @Param input body UpdateSpotStatusRequest true "状态信息"
// @Security BearerAuth
// @Success 200 {object} ParkingSpotResponse
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /admin/spots/{id}/status [put]
func (c *AdminController) UpdateSpotStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的车位 ID"})
		return
	}

	var req UpdateSpotStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	status := models.ParkingStatus(req.Status)

	if status != models.Idle && status != models.Occupied && status != models.Faulty {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的车位状态"})
		return
	}

	spot, err := c.parkingService.UpdateSpotStatus(ctx, uint(id), status, req.Notes)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, ToParkingSpotResponse(spot))
}

// GetSystemStats 获取系统统计数据
// @Summary 获取系统统计数据
// @Description 返回当前系统的车位总数、可用车位数、各类型车位利用率等
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SystemStatsResponse
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /admin/stats [get]
func (c *AdminController) GetSystemStats(ctx *gin.Context) {
	stats, err := c.reportService.GetSpotStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
