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

type ParkingSpot struct {
	ID          uint          `gorm:"primaryKey"`
	Type        ParkingType   `gorm:"type:varchar(20);not null"`
	Status      ParkingStatus `gorm:"type:varchar(20);not null;default:'idle'"`
	OwnerID     *uint
	License     string  `gorm:"type:varchar(100)"`
	HourlyRate  float64 `gorm:"type:decimal(10,2)"`
	MonthlyRate float64 `gorm:"type:decimal(10,2)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	ExpiresAt   *time.Time `gorm:"index"` // 添加过期时间字段
	Notes       string
}

type ParkingRecord struct {
	ID          uint      `gorm:"primaryKey"`
	SpotID      uint      `gorm:"not null"`
	UserID      *uint     // 关联用户（如果是业主）
	License     string    `gorm:"type:varchar(100);not null"`
	EntryTime   time.Time `gorm:"not null"`
	ExitTime    *time.Time
	TotalCost   float64 `gorm:"type:decimal(10,2)"`
	IsCompleted bool    `gorm:"default:false"`
	VehicleID   *uint   // 添加关联车辆ID
}
