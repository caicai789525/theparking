// internal/models/lease.go
package models

import (
	"time"
)

type LeaseStatus string

const (
	LeaseActive  LeaseStatus = "active"
	LeaseExpired LeaseStatus = "expired"
)

type LeaseOrder struct {
	ID         uint
	UserID     uint
	SpotID     uint
	StartDate  time.Time
	EndDate    time.Time
	TotalPrice float64 `gorm:"type:decimal(10,2)"`
	Status     LeaseStatus
	AutoRenew  bool `gorm:"default:false"`
	CreatedAt  time.Time
}
