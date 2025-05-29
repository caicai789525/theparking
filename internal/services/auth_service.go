// internal/services/auth_service.go
package services

import (
	"context"
	"errors"
	"modules/config"
	"modules/internal/models"
	"modules/internal/repositories"
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

	// 生成JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(s.Cfg.JWT.ExpiresIn).Unix(),
	})

	return token.SignedString([]byte(s.Cfg.JWT.Secret))
}
