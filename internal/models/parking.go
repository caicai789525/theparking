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

// ParkingSpot 车位信息
type ParkingSpot struct {
	// 创建时间
	CreatedAt string `json:"createdAt"`
	// 过期时间
	ExpiresAt string `json:"expiresAt" description:"添加过期时间字段"`
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
