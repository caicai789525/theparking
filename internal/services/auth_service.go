// internal/services/auth_service.go
package services

import (
	"context"
	"errors"
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

// 用户登录
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", errors.New("用户不存在")
	}

	if err := user.CheckPassword(password); err != nil {
		return "", errors.New("密码错误")
	}

	var roles []models.Role
	if err := json.Unmarshal(user.Roles, &roles); err != nil {
		return "", err
	}

	// 生成 JWT
	claims := utils.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Cfg.JWT.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "parking_system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Cfg.JWT.Secret))
}

// AdminLogin 管理员登录
func (s *AuthService) AdminLogin(ctx context.Context, username, password string) (string, error) {
	// 假设管理员用户名是 "admin"，密码是 "admin123"
	if username != "admin" || password != "admin123" {
		return "", errors.New("用户名或密码错误")
	}

	// 生成 JWT
	claims := utils.Claims{
		UserID:   0,
		Username: "admin",
		Roles:    []models.Role{models.Admin}, // 假设 Admin 是 models.Role 类型
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Cfg.JWT.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "parking_system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Cfg.JWT.Secret))
}
