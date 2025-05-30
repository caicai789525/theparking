// internal/services/auth_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
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

// Login 用户/管理员通用登录方法
func (s *AuthService) Login(ctx context.Context, username, password string, checkAdmin bool) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("查询用户失败: %w", err)
	}

	if err := user.CheckPassword(password); err != nil {
		return "", fmt.Errorf("密码验证失败: %w", err)
	}

	var roles []models.Role
	if err := json.Unmarshal(user.Roles, &roles); err != nil {
		return "", fmt.Errorf("反序列化用户角色失败: %w", err)
	}

	if checkAdmin {
		isAdmin := false
		for _, role := range roles {
			if role == models.Admin {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			return "", errors.New("非管理员用户")
		}
	}

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
