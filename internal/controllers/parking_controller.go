// internal/controllers/parking_controller.go
package controllers

import (
	"modules/internal/models"
	"modules/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ParkingController struct {
	service *services.ParkingService
}

// CreateParkingSpotRequest 创建车位请求
type CreateParkingSpotRequest struct {
	// 每小时费率
	HourlyRate float64 `json:"hourlyRate" binding:"required"`
	// 每月费率
	MonthlyRate float64 `json:"monthlyRate" binding:"required"`
	// 车位类型
	Type string `json:"type" binding:"required,oneof=permanent short_term temporary"`
	// 备注
	Notes string `json:"notes"`
}

func NewParkingController(service *services.ParkingService) *ParkingController {
	return &ParkingController{
		service: service,
	}
}

// @Summary 获取车位列表
// @Description 获取所有车位的详细列表，包括类型、状态和收费标准
// @Tags parking
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ParkingSpot "返回车位列表"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /parking/spots [get]
func (c *ParkingController) ListSpots(ctx *gin.Context) {
	spots, err := c.service.ListSpots(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, spots)
}

// @Summary 车辆入场登记
// @Description 车辆入场时登记车牌号，开始计费
// @Tags parking
// @Accept json
// @Produce json
// @Param input body EntryRequest true "入场信息"
// @Security BearerAuth
// @Success 200 {object} RecordResponse "入场记录"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /parking/entry [post]
func (c *ParkingController) Entry(ctx *gin.Context) {
	var req EntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, _ := ctx.Get("userID")
	var uid *uint
	if v, ok := userID.(uint); ok {
		uid = &v
	}

	record, err := c.service.ProcessEntry(ctx, req.License, uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, ToRecordResponse(record))
}

// @Summary 车辆出场结算
// @Description 车辆出场时结算停车费用，返回停车记录和费用信息
// @Tags parking
// @Accept json
// @Produce json
// @Example {"cost": 25.5, "entry_time": "2023-10-01T09:00:00Z", "exit_time": "2023-10-01T12:30:00Z"}
// @Param id path int true "停车记录ID"
// @Security BearerAuth
// @Success 200 {object} RecordResponse "出场结算记录"
// @Failure 400 {object} ErrorResponse "无效的ID参数"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /parking/exit/{id} [post]
func (c *ParkingController) Exit(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	record, err := c.service.ProcessExit(ctx, uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, ToRecordResponse(record))
}

// DTOs和转换方法
type EntryRequest struct {
	// 车牌号
	License string `json:"license" binding:"required"`
}

// RecordResponse 停车记录响应
type RecordResponse struct {
	// 记录ID
	ID uint `json:"id"`
	// 车位ID
	SpotID uint `json:"spot_id"`
	// 车牌号
	License string `json:"license"`
	// 入场时间
	EntryTime string `json:"entry_time"`
	// 出场时间
	ExitTime string `json:"exit_time,omitempty"`
	// 停车费用
	Cost float64 `json:"cost"`
}

func ToRecordResponse(r *models.ParkingRecord) *RecordResponse {
	res := &RecordResponse{
		ID:        r.ID,
		SpotID:    r.SpotID,
		License:   r.License,
		EntryTime: r.EntryTime.Format(time.RFC3339),
	}

	if r.ExitTime != nil {
		res.ExitTime = r.ExitTime.Format(time.RFC3339)
		res.Cost = r.TotalCost
	}

	return res
}

func ToParkingSpotResponse(spot *models.ParkingSpot) *ParkingSpotResponse {
	return &ParkingSpotResponse{
		ID:         spot.ID,
		Type:       string(spot.Type),
		Status:     string(spot.Status),
		HourlyRate: spot.HourlyRate,
	}
}

type CreateSpotRequest struct {
	Type       models.ParkingType `json:"type" binding:"required"`
	HourlyRate float64            `json:"hourly_rate"`
}

// @Summary 创建车位
// @Description 新增一个车位，指定车位类型和收费标准
// @Tags parking
// @Accept json
// @Produce json
// @Example {"type": "temporary", "hourly_rate": 5}
// @Param input body CreateSpotRequest true "车位信息"
// @Security BearerAuth
// @Success 201 {object} ParkingSpotResponse "创建成功的车位信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /parking/spots [post]
func (c *ParkingController) CreateSpot(ctx *gin.Context) {
	var req CreateSpotRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spot := &models.ParkingSpot{
		Type:       string(req.Type),
		HourlyRate: req.HourlyRate,
		// 显式将 ParkingStatus 类型转换为 string 类型

		Status: string(models.Idle),
	}

	createdSpot, err := c.service.CreateSpot(ctx, spot)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, ToParkingSpotResponse(createdSpot))
}

// @Router /parking/my-spots [get]
// @Summary 查询自己的车位
// @Description 查询当前用户名下的所有车位信息
// @Tags parking
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ParkingSpot "返回用户的车位列表"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func (c *ParkingController) GetUserSpots(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	spots, err := c.service.GetUserSpots(ctx, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, spots)
}
