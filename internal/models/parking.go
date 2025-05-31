// internal/models/parking.go
package models

import (
	"time"
)

type ParkingType string

const (
	Permanent ParkingType = "permanent"
	ShortTerm ParkingType = "short_term"
	Temporary ParkingType = "temporary"
)

type ParkingStatus string

const (
	Idle     ParkingStatus = "idle"
	Occupied ParkingStatus = "occupied"
	Faulty   ParkingStatus = "faulty"
)

// BindParkingRequest 管理员绑定车位给用户的请求结构体
type BindParkingRequest struct {
	UserID    uint `json:"user_id" binding:"required"`
	ParkingID uint `json:"parking_id" binding:"required"`
}

// UserInfoResponse 用户信息响应结构体
type UserInfoResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	// 可根据实际需求添加更多字段
}

// ParkingSpot 车位信息
type ParkingSpot struct {
	// 创建时间
	CreatedAt string `json:"createdAt"`
	// 过期时间，修改为 VARCHAR 类型
	ExpiresAt string `json:"expiresAt" description:"添加过期时间字段" gorm:"type:varchar(255)"`
	// 每小时费率
	HourlyRate float64 `json:"hourlyRate"`
	// 车位ID
	ID uint `json:"id" gorm:"primaryKey"`
	// 车牌号
	License string `json:"license"`
	// 每月费率
	MonthlyRate float64 `json:"monthlyRate"`
	// 备注
	Notes string `json:"notes"`
	// 业主ID
	OwnerID uint `json:"ownerID"`
	// 车位状态
	Status string `json:"status" gorm:"type:enum('idle', 'occupied', 'faulty')"`
	// 车位类型
	Type string `json:"type" gorm:"type:enum('permanent', 'short_term', 'temporary')"`
	// 更新时间
	UpdatedAt string `json:"updatedAt"`
}

// ParkingRecord 停车记录
type ParkingRecord struct {
	// 记录ID
	ID uint `gorm:"primaryKey"`
	// 车位ID
	SpotID uint `gorm:"not null"`
	// 用户ID
	UserID *uint // 关联用户（如果是业主）
	// 车牌号
	License string `gorm:"type:varchar(100);not null"`
	// 入场时间
	EntryTime time.Time `gorm:"not null"`
	// 出场时间
	ExitTime *time.Time
	// 总费用
	TotalCost float64 `gorm:"type:decimal(10,2)"`
	// 是否完成
	IsCompleted bool `gorm:"default:false"`
	// 车辆ID
	VehicleID *uint // 添加关联车辆ID
}

type UnbindParkingRequest struct {
	UserID    uint `json:"user_id" binding:"required"`
	ParkingID uint `json:"parking_id" binding:"required"`
}

type ParkingBindUserResponse struct {
	ParkingID uint   `json:"parking_id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username,omitempty"`
}

// BindParkingResponse 绑定车位响应结构体
type BindParkingResponse struct {
	Message string `json:"message"`
}
