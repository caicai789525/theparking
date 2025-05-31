// internal/controllers/vehicle_controller.go
package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"modules/internal/models"
	"modules/internal/services"
	"net/http"
	"strconv"
)

type VehicleController struct {
	service *services.VehicleService
}

func NewVehicleController(service *services.VehicleService) *VehicleController {
	return &VehicleController{service: service}
}

// @Summary 绑定车辆
// @Description 将车辆绑定到当前用户账户，支持填写车牌号、品牌及车型信息。业主绑定车辆（可绑定多个）
// @Tags vehicle
// @Accept json
// @Produce json
// @Example {"license": "粤B12345", "brand": "Tesla", "model": "Model 3"}
// @Param input body BindVehicleRequest true "绑定车辆请求体"
// @Security BearerAuth
// @Success 200 {object} VehicleResponse "绑定成功返回车辆信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /vehicles [post]
func (c *VehicleController) BindVehicle(ctx *gin.Context) {
	var req BindVehicleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.MustGet("userID").(uint)
	vehicle, err := c.service.BindVehicle(ctx, userID, req.License, req.Brand, req.Model)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, ToVehicleResponse(vehicle))
}

// @Summary 出租车位
// @Description 用户将自己的车位发布出租，需指定车位ID、出租价格及出租天数。
// @Tags parking
// @Accept json
// @Produce json
// @Example {"spot_id": 1, "days": 30, "rate": 280}
// @Param input body RentRequest true "出租车位请求体"
// @Security BearerAuth
// @Success 200 {object} LeaseResponse "发布成功返回租赁信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /parking/rent [post]
func (c *VehicleController) PublishForRent(ctx *gin.Context) {
	var req RentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.MustGet("userID").(uint)
	log.Printf("Received request to publish spot for rent: userID=%d, spotID=%d, rate=%f, period=%d", userID, req.SpotID, req.Rate, req.Days)

	lease, err := c.service.PublishSpotForRent(ctx, userID, req.SpotID, req.Rate, req.Days)
	if err != nil {
		log.Printf("Error in PublishSpotForRent: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, ToLeaseResponse(lease))
}

// DTOs
type BindVehicleRequest struct {
	License string `json:"license" binding:"required"`
	Brand   string `json:"brand"`
	Model   string `json:"model"`
}

type VehicleResponse struct {
	ID        uint   `json:"id"`
	License   string `json:"license"`
	Brand     string `json:"brand"`
	Model     string `json:"model"`
	IsDefault bool   `json:"is_default"`
}

type RentRequest struct {
	SpotID uint    `json:"spot_id" binding:"required"`
	Rate   float64 `json:"rate" binding:"required"`
	Days   int     `json:"days" binding:"required,min=1"`
}

// 转换方法
func ToVehicleResponse(v *models.Vehicle) *VehicleResponse {
	return &VehicleResponse{
		ID:        v.ID,
		License:   v.LicensePlate,
		Brand:     v.Brand,
		Model:     v.Model,
		IsDefault: v.IsDefault,
	}
}

// RemoveVehicle 删除车辆
// @Summary 删除车辆
// @Description 用户删除自己绑定的车辆
// @Tags vehicle
// @Accept json
// @Produce json
// @Param id path int true "车辆ID"
// @Security BearerAuth
// @Success 200 {object} controllers.MessageResponse "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 404 {object} ErrorResponse "车辆不存在或无权操作"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /vehicles/{id} [delete]
func (c *VehicleController) RemoveVehicle(ctx *gin.Context) {
	vehicleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的车辆 ID"})
		return
	}

	userID := ctx.MustGet("userID").(uint)

	err = c.service.RemoveVehicle(ctx, userID, uint(vehicleID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "车辆删除成功"})
}

// GetUserVehicles 查询自己的车辆
// @Summary 查询自己的车辆
// @Description 查询当前用户名下的所有车辆信息
// @Tags vehicle
// @Produce json
// @Security BearerAuth
// @Success 200 {array} VehicleResponse "返回用户的车辆列表"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /vehicles [get]
func (c *VehicleController) GetUserVehicles(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)

	vehicles, err := c.service.GetUserVehicles(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	var response []VehicleResponse
	for _, vehicle := range vehicles {
		response = append(response, VehicleResponse{
			ID:        vehicle.ID,
			License:   vehicle.LicensePlate,
			Brand:     vehicle.Brand,
			Model:     vehicle.Model,
			IsDefault: vehicle.IsDefault,
		})
	}

	ctx.JSON(http.StatusOK, response)
}
