// internal/services/auth_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"log"
	"modules/config"
	"modules/internal/models"
	"modules/internal/repositories"
	"modules/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo repositories.UserRepository
	Cfg      *config.Config
}

func NewAuthService(userRepo repositories.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		Cfg:      cfg,
	}
}

// 用户注册
func (s *AuthService) Register(ctx context.Context, username, password, email string) error {
	// 检查用户名是否已存在
	if _, err := s.userRepo.GetUserByUsername(ctx, username); err == nil {
		return errors.New("用户名已存在")
	}

	user := &models.User{
		Username: username,
		Password: password,
		Email:    email,
	}

	if err := user.HashPassword(); err != nil {
		return err
	}

	return s.userRepo.CreateUser(ctx, user)
}

// 用户登录
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("查询用户失败: %w", err)
	}

	if err := user.CheckPassword(password); err != nil {
		return "", fmt.Errorf("密码验证失败: %w", err)
	}

	var roles []models.Role
	// 检查 user.Roles 是否为空或 nil
	if len(user.Roles) == 0 {
		// 处理角色为空的情况，你可以选择返回空切片或报错
		roles = []models.Role{}
	} else {
		if err := json.Unmarshal(user.Roles, &roles); err != nil {
			return "", fmt.Errorf("反序列化用户角色失败: %w", err)
		}
	}

	now := time.Now()
	// 生成 JWT
	claims := utils.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.Cfg.JWT.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "parking_system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Cfg.JWT.Secret))
}

// AdminLogin 管理员登录
func (s *AuthService) AdminLogin(ctx context.Context, username, password string) (string, error) {
	// 从数据库获取用户信息
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		log.Printf("查询用户失败: %v", err)
		return "", fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户是否为管理员
	var roles []models.Role
	if err := json.Unmarshal(user.Roles, &roles); err != nil {
		log.Printf("反序列化用户角色失败: %v, user.Roles: %s", err, string(user.Roles))
		return "", fmt.Errorf("反序列化用户角色失败: %w", err)
	}

	isAdmin := false
	for _, role := range roles {
		if role == models.Admin {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		log.Printf("该用户不是管理员, username: %s, roles: %v", username, roles)
		return "", errors.New("该用户不是管理员")
	}

	// 验证密码
	if err := user.CheckPassword(password); err != nil {
		log.Printf("密码验证失败, username: %s", username)
		return "", fmt.Errorf("密码验证失败: %w", err)
	}

	// 生成 JWT
	now := time.Now()
	claims := utils.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.Cfg.JWT.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "parking_system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Cfg.JWT.Secret))
}
