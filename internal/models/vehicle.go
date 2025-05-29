// internal/models/vehicle.go
package models

import "time"

type Vehicle struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   `gorm:"not null;index"`
	LicensePlate string `gorm:"type:varchar(20);uniqueIndex;not null"`
	Brand        string `gorm:"type:varchar(50)"`
	Model        string `gorm:"type:varchar(50)"`
	IsDefault    bool   `gorm:"default:false"`
	CreatedAt    time.Time
}

//migrations/005_init_vehicles.up.sql
