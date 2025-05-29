package controllers

import (
	"modules/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *services.AuthService
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=1,max=50"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Email    string `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
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

	if err := c.service.Register(ctx, req.Username, req.Password, req.Email); err != nil {
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, MessageResponse{Message: "注册成功"})
}

// Login 用户登录
// @Summary 用户登录
// @Description 登录并返回 JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Example {"username": "user1", "password": "pass123"}
// @Param input body LoginRequest true "登录信息"
// @Success 200 {object} TokenResponse "登录成功，返回token"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "认证失败，用户名或密码错误"
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	token, err := c.service.Login(ctx, req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, TokenResponse{Token: token})
}
