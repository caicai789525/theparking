package controllers

import (
	"github.com/gin-gonic/gin"
	"modules/internal/models"
	"modules/internal/services"
	"net/http"
	"time"
)

type LeaseController struct {
	service *services.LeaseService
}

func NewLeaseController(service *services.LeaseService) *LeaseController {
	return &LeaseController{service: service}
}

// CreateLease 创建租赁订单
// @Summary 创建租赁订单
// @Description 用户根据车位ID和租赁时长创建订单
// @Tags lease
// @Accept json
// @Produce json
// @Example {"spot_id": 2, "months": 3, "rate": 300}
// @Param input body LeaseRequest true "租赁信息"
// @Security BearerAuth
// @Success 200 {object} LeaseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /lease [post]
func (c *LeaseController) CreateLease(ctx *gin.Context) {
	var req LeaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := ctx.MustGet("userID").(uint)

	lease, err := c.service.CreateLease(ctx, userID, req.SpotID, req.Months, req.Rate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, ToLeaseResponse(lease))
}

//
// ========== DTO 定义与响应结构 ==========
//

// LeaseRequest 租赁订单请求结构
type LeaseRequest struct {
	SpotID uint    `json:"spot_id" binding:"required"`      // 车位ID
	Months int     `json:"months" binding:"required,min=1"` // 租赁月数
	Rate   float64 `json:"rate" binding:"required"`         // 每月租金
}

// LeaseResponse 租赁订单响应结构
type LeaseResponse struct {
	ID        uint    `json:"id"`         // 订单ID
	SpotID    uint    `json:"spot_id"`    // 车位ID
	StartDate string  `json:"start_date"` // 起始日期（格式：YYYY-MM-DD）
	EndDate   string  `json:"end_date"`   // 结束日期
	Total     float64 `json:"total"`      // 总金额
	Status    string  `json:"status"`     // 当前状态
}

// RentParkingSpotRequest 出租车位请求
type RentParkingSpotRequest struct {
	// 车位ID
	SpotID uint `json:"spotID" binding:"required"`
	// 出租价格
	RentPrice float64 `json:"rentPrice" binding:"required"`
	// 出租天数
	RentDays int `json:"rentDays" binding:"required"`
}

// PaymentRequest 支付请求结构（预留接口可用）
// 用于支付租赁费用
type PaymentRequest struct {
	LeaseID uint    `json:"lease_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
}

// CreateLeaseOrderRequest 创建租赁订单请求
type CreateLeaseOrderRequest struct {
	// 车位ID
	SpotID uint `json:"spotID" binding:"required"`
	// 租赁开始时间
	StartTime time.Time `json:"startTime" binding:"required"`
	// 租赁结束时间
	EndTime time.Time `json:"endTime" binding:"required"`
}

// PaymentResponse 支付响应结构
type PaymentResponse struct {
	ID          string  `json:"id"`          // 支付单号
	Amount      float64 `json:"amount"`      // 实付金额
	Description string  `json:"description"` // 支付说明
	Status      string  `json:"status"`      // 支付状态
}

// ToLeaseResponse 将租赁模型转为响应结构
func ToLeaseResponse(l *models.LeaseOrder) *LeaseResponse {
	return &LeaseResponse{
		ID:        l.ID,
		SpotID:    l.SpotID,
		StartDate: l.StartDate.Format("2006-01-02"),
		EndDate:   l.EndDate.Format("2006-01-02"),
		Total:     l.TotalPrice,
		Status:    string(l.Status),
	}
}
