package repositories

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"strings"

	"gorm.io/gorm"

	"modules/internal/models"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	// 定义 CreateUser 方法
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	CheckUserExists(ctx context.Context, username, email string) (bool, error)
}

type userRepo struct {
	db *gorm.DB
}

func (r *userRepo) CreateUser(ctx context.Context, user *models.User) error {
	// 使用 GORM 的 Create 方法在数据库中创建用户记录
	// WithContext 方法将上下文传递给数据库操作，支持请求取消、超时控制等
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		// 如果创建过程中出现错误，返回该错误
		return result.Error
	}
	if result.RowsAffected == 0 {
		// 如果没有影响任何行，说明创建失败，返回自定义错误
		return errors.New("用户创建失败，未插入任何记录")
	}
	return nil
}
func NewUserRepo(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// GetUserByUsername 根据用户名查询用户信息
func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	// 输入验证，检查用户名是否为空
	if strings.TrimSpace(username) == "" {
		zap.L().Error("查询用户时，用户名为空")
		return nil, errors.New("用户名不能为空")
	}

	var user models.User
	// 去除用户名前后空格并转换为小写，进行不区分大小写的查询
	lowerUsername := strings.ToLower(strings.TrimSpace(username))
	zap.L().Info("开始查询用户", zap.String("username", lowerUsername))

	err := r.db.WithContext(ctx).Where("LOWER(username) = ?", lowerUsername).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			zap.L().Info("未找到用户", zap.String("username", lowerUsername))
			return nil, models.ErrUserNotFound
		}
		zap.L().Error("查询用户时发生数据库错误", zap.String("username", lowerUsername), zap.Error(err))
		return nil, err
	}

	zap.L().Info("成功找到用户", zap.String("username", user.Username))
	return &user, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) CheckUserExists(ctx context.Context, username, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
