// internal/services/auth_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5" // 导入 JWT 包
	"gorm.io/gorm"
	"log"
	"modules/config"
	"modules/internal/models"
	"modules/internal/repositories"
	"modules/internal/utils"
	"regexp"
	"time"
)

type AuthService struct {
	userRepo repositories.UserRepository
	Cfg      *config.Config
}

// Claims 定义 JWT 声明结构
type Claims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
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
	// 检查用户是否已存在
	exists, err := s.userRepo.CheckUserExists(ctx, username, email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("用户名或邮箱已存在")
	}

	// 创建用户对象
	user := &models.User{
		Username: username,
		Password: password,
		Email:    email,
	}

	// 对密码进行哈希处理
	if err := user.HashPassword(); err != nil {
		return err
	}

	// 将用户信息保存到数据库
	return s.userRepo.CreateUser(ctx, user)
}

// GenerateToken 生成 JWT 令牌
func (s *AuthService) GenerateToken(userID uint, username string, roles []string) (string, error) {
	expiresIn, err := time.ParseDuration(s.Cfg.JWT.ExpiresIn)
	if err != nil {
		return "", fmt.Errorf("解析令牌过期时间失败: %w", err)
	}
	return GenerateJWT(s.Cfg.JWT.Secret, userID, username, roles, expiresIn)
}

// GenerateJWT 生成 JWT 令牌
func GenerateJWT(secret string, userID uint, username string, roles []string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken 验证 JWT 令牌
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.Cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// Login 用户/管理员通用登录方法
func (s *AuthService) Login(ctx context.Context, username, password string, checkAdmin bool) (string, error) {
	if s.userRepo == nil {
		return "", errors.New("userRepo is not initialized")
	}
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

	// 将 s.Cfg.JWT.ExpiresIn 转换为 time.Duration 类型
	expiresIn, err := time.ParseDuration(s.Cfg.JWT.ExpiresIn)
	if err != nil {
		log.Printf("解析令牌过期时间失败，使用默认值 24h: %v", err)
		expiresIn = 24 * time.Hour
	}

	return utils.GenerateJWT(s.Cfg.JWT.Secret, user.ID, user.Username, roleStrings, expiresIn)
}

// AdminLogin 管理员登录方法，复用 Login 方法
func (s *AuthService) AdminLogin(ctx context.Context, username, password string) (string, error) {
	return s.Login(ctx, username, password, true)
}
