// internal/models/user.go
package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"time"
)

type Role string

const (
	Admin  Role = "admin"
	Owner  Role = "owner"
	Renter Role = "renter"
)

type User struct {
	ID        uint           `gorm:"primaryKey"`
	Username  string         `gorm:"uniqueIndex;size:50;not null"`
	Password  string         `gorm:"size:100;not null"`
	Email     string         `gorm:"uniqueIndex;size:100;not null"`
	Phone     string         `gorm:"size:20"`
	Roles     datatypes.JSON `gorm:"type:json"` //
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	Vehicles  []Vehicle      `gorm:"foreignKey:UserID"`
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 密码加密
func (u *User) HashPassword() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

// 密码验证
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
