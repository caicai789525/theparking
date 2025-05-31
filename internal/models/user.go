// internal/models/user.go
package models

import (
	"errors"
	"github.com/goccy/go-json"
	"golang.org/x/crypto/bcrypt"
	"modules/internal/utils"
	"time"
)

type Role string

const (
	Admin  Role = "admin"
	Owner  Role = "owner"
	Renter Role = "renter"
)

var (
	ErrUserNotFound    = errors.New("用户不存在")
	ErrParkingNotFound = errors.New("车位不存在")
)

type JSONBytes []byte

func (r JSONBytes) Unmarshal(dst *[]Role) error {
	return json.Unmarshal(r, dst)
}

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;size:50;not null"`
	Password  string    `gorm:"size:100;not null"`
	Email     string    `gorm:"uniqueIndex;size:100;not null"`
	Phone     string    `gorm:"size:20"`
	Roles     JSONBytes `gorm:"type:json"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	IsActive  bool      `gorm:"default:true"` // 新增用户活跃状态字段
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// HashPassword 对密码进行哈希处理
func (u *User) HashPassword() error {
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 检查密码是否正确
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
