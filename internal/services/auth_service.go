// internal/services/auth_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"modules/config"
	"modules/internal/models"
	"modules/internal/repositories"
	"modules/internal/utils"
	"regexp"
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

// 邮箱正则表达式
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// 用户注册
func (s *AuthService) Register(ctx context.Context, username, password, email string) error {
	// 输入验证
	if username == "" {
		return errors.New("用户名不能为空")
	}
	if password == "" {
		return errors.New("密码不能为空")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}

	// 检查用户名是否已存在
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("查询用户名失败: %v", err)
			return fmt.Errorf("查询用户名失败: %w", err)
		}
	} else if user != nil {
		return errors.New("用户名已存在")
	}

	user = &models.User{
		Username: username,
		Password: password,
		Email:    email,
	}

	if err := user.HashPassword(); err != nil {
		log.Printf("密码哈希处理失败: %v", err)
		return err
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		log.Printf("创建用户记录失败: %v", err)
		return fmt.Errorf("创建用户记录失败: %w", err)
	}

	return nil
}

// Login 用户/管理员通用登录方法
func (s *AuthService) Login(ctx context.Context, username, password string, checkAdmin bool) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("用户不存在")
		}
		return "", fmt.Errorf("查询用户失败: %w", err)
	}

	if err := user.CheckPassword(password); err != nil {
		return "", errors.New("密码错误")
	}

	var roles []models.Role
	if len(user.Roles) == 0 {
		roles = []models.Role{}
	} else {
		if err := user.Roles.Unmarshal(&roles); err != nil {
			return "", fmt.Errorf("反序列化用户角色失败: %w", err)
		}
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
			return "", errors.New("非管理员用户，无权访问")
		}
	}

	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = string(role)
	}

	return utils.GenerateJWT(s.Cfg.JWT.Secret, user.ID, user.Username, roleStrings, s.Cfg.JWT.ExpiresIn)
}

// AdminLogin 管理员登录方法，复用 Login 方法
func (s *AuthService) AdminLogin(ctx context.Context, username, password string) (string, error) {
	return s.Login(ctx, username, password, true)
}
