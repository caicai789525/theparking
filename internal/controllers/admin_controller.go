package controllers

import (
	"errors"
	"modules/internal/models"
	"modules/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	parkingService *services.ParkingService
	reportService  *services.ReportService
	authService    *services.AuthService
	// JWT Token
	Token string `json:"token"`
}

// NewAdminController 创建一个新的 AdminController 实例
// @Summary 创建 AdminController 实例
// @Description 根据传入的停车服务、报告服务和认证服务实例创建 AdminController 实例
// @Tags 控制器初始化
// @Param ps body services.ParkingService true "停车服务实例"
// @Param rs body services.ReportService true "报告服务实例"
// @Param as body services.AuthService true "认证服务实例"
// @Success 200 {object} AdminController "成功创建 AdminController 实例"
// @Router /internal/create-admin-controller [post]
// 修改构造函数，添加 authService 参数
func NewAdminController(ps *services.ParkingService, rs *services.ReportService, as *services.AuthService) *AdminController {
	return &AdminController{
		parkingService: ps,
		reportService:  rs,
		authService:    as, // 初始化 authService
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

type AdminLoginResponse struct {
	// JWT Token
	Token string `json:"token"`
}

// AdminLogin 管理员登录
// @Summary 管理员登录
// @Description 管理员登录并返回 JWT token
// @Tags admin
// @Accept json
// @Produce json
// @Param input body AdminLoginRequest true "登录信息"
// @Success 200 {object} AdminLoginResponse "登录成功，返回token"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "认证失败，用户名或密码错误"
// @Failure 403 {object} ErrorResponse "非管理员用户，无权访问"
// @Router /admin/login [post]
func (c *AdminController) AdminLogin(ctx *gin.Context) {
	var req AdminLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	token, err := c.authService.AdminLogin(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, errors.New("非管理员用户，无权访问")) {
			ctx.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, AdminLoginResponse{Token: token})
}

// GetUserInfo 查询用户信息
// @Summary 查询用户信息
// @Description 管理员根据用户 ID 查询用户信息
// @Tags admin
// @Produce json
// @Param userID path uint true "用户 ID"
// @Security BearerAuth
// @Success 200 {object} models.UserInfoResponse "用户信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /admin/users/{userID} [get]
func (c *AdminController) GetUserInfo(ctx *gin.Context) {
	userIDStr := ctx.Param("userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "无效的用户 ID"})
		return
	}

	// 从 Gin 上下文获取 context.Context 对象
	userInfo, err := c.parkingService.GetUserInfo(ctx.Request.Context(), uint(userID))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, userInfo)
}

// BindParkingToUser 管理员将车位绑定给用户
// @Summary 管理员将车位绑定给用户
// @Description 管理员根据用户 ID 和车位 ID 将车位绑定给指定用户
// @Tags admin
// @Accept json
// @Produce json
// @Param input body models.BindParkingRequest true "绑定车位请求体"
// @Security BearerAuth
// @Success 200 {object} models.BindParkingResponse "绑定成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /admin/bind-parking [post]
func (c *AdminController) BindParkingToUser(ctx *gin.Context) {
	var req models.BindParkingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 从 Gin 上下文获取 context.Context 对象
	err := c.parkingService.BindParkingToUser(ctx.Request.Context(), req.UserID, req.ParkingID)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "车位绑定成功"})
}

// UnbindParkingFromUser 管理员解除车位与用户的绑定
// @Summary 管理员解除车位与用户的绑定
// @Description 管理员根据用户 ID 和车位 ID 解除车位与指定用户的绑定
// @Tags admin
// @Accept json
// @Produce json
// @Param input body models.UnbindParkingRequest true "解除绑定车位请求体"
// @Security BearerAuth
// @Success 200 {object} models.BindParkingResponse "解除绑定成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 404 {object} ErrorResponse "用户或车位不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /admin/unbind-parking [post]
func (c *AdminController) UnbindParkingFromUser(ctx *gin.Context) {
	var req models.UnbindParkingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := c.parkingService.UnbindParkingFromUser(ctx.Request.Context(), req.UserID, req.ParkingID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) || errors.Is(err, models.ErrParkingNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, models.BindParkingResponse{Message: "车位解除绑定成功"})
}

// GetParkingBindUser 查询车位绑定的用户信息
// @Summary 查询车位绑定的用户信息
// @Description 管理员根据车位 ID 查询车位绑定的用户信息
// @Tags admin
// @Produce json
// @Param parkingID path uint true "车位 ID"
// @Security BearerAuth
// @Success 200 {object} models.ParkingBindUserResponse "车位绑定用户信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权访问"
// @Failure 404 {object} ErrorResponse "车位不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /admin/parking/{parkingID}/bind-user [get]
func (c *AdminController) GetParkingBindUser(ctx *gin.Context) {
	parkingIDStr := ctx.Param("parkingID")
	parkingID, err := strconv.ParseUint(parkingIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的车位 ID"})
		return
	}

	response, err := c.parkingService.GetParkingBindUser(ctx.Request.Context(), uint(parkingID))
	if err != nil {
		if err == models.ErrParkingNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "车位不存在"})
		} else if err == models.ErrUserNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "用户不存在"})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}
