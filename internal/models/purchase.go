// internal/models/purchase.go
package models

import "time"

type PurchaseRecord struct {
	ID            uint         `gorm:"primaryKey"`
	UserID        uint         `gorm:"not null"`
	SpotID        uint         `gorm:"not null"`
	PurchasePrice float64      `gorm:"type:decimal(10,2)"`
	PurchaseDate  time.Time    `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	User          *User        `gorm:"foreignKey:UserID"`
	Spot          *ParkingSpot `gorm:"foreignKey:SpotID"`
}
