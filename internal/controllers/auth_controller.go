package controllers

import (
	"errors"
	"log"
	"modules/internal/models"
	"modules/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *services.AuthService
}

// RegisterResponse 用户注册响应
type RegisterResponse struct {
	// 用户ID
	ID uint `json:"id"`
	// 用户名
	Username string `json:"username"`
	// 邮箱
	Email string `json:"email"`
}

// 定义 SuccessResponse 结构体
type SuccessResponse struct {
	Message string `json:"message"`
}

// LoginResponse 用户登录响应
type LoginResponse struct {
	// JWT Token
	Token string `json:"token"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	// 用户名
	Username string `json:"username" binding:"required"`
	// 密码
	Password string `json:"password" binding:"required"`
	// 邮箱
	Email string `json:"email" binding:"required,email"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	// 用户名
	Username string `json:"username" binding:"required"`
	// 密码
	Password string `json:"password" binding:"required"`
}

// 成功消息响应结构
type MessageResponse struct {
	Message string `json:"message"`
}

// 登录成功返回token结构
type TokenResponse struct {
	Token string `json:"token"`
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册一个新用户
// @Tags auth
// @Accept json
// @Produce json
// @Example {"username": "user1", "password": "pass123", "email": "user@example.com"}
// @Param input body RegisterRequest true "注册信息"
// @Success 201 {object} MessageResponse "注册成功消息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 409 {object} ErrorResponse "用户名或邮箱已存在"
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := c.service.Register(ctx, req.Username, req.Password, req.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			// 用户不存在，正常注册，不应该返回错误
			// 继续执行创建用户逻辑，这里代码逻辑已在 service 里处理
		} else if errors.Is(err, errors.New("用户名已存在")) {
			ctx.JSON(http.StatusConflict, ErrorResponse{Error: "用户名已存在"})
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusCreated, SuccessResponse{Message: "注册成功"})
}

// UserLogin 用户登录
// @Summary 用户登录
// @Description 用户登录并返回 JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Example {"username": "user1", "password": "pass123"}
// @Param input body LoginRequest true "登录信息"
// @Success 200 {object} LoginResponse "登录成功，返回token"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "认证失败，用户名或密码错误"
// @Router /login [post]
func (c *AuthController) UserLogin(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数绑定失败: %v", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "请求参数错误"})
		return
	}

	token, err := c.service.Login(ctx, req.Username, req.Password, false)
	if err != nil {
		log.Printf("用户登录失败: %v", err)
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "认证失败，用户名或密码错误"})
		return
	}

	ctx.JSON(http.StatusOK, LoginResponse{Token: token})
}
