package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"modules/internal/models"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	// 定义 CreateUser 方法
	CreateUser(ctx context.Context, user *models.User) error
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

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}
