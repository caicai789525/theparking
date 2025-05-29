// internal/models/report.go
package models

import (
	"time"
)

type DailyReport struct {
	Date         time.Time `gorm:"type:date"`
	TotalIncome  float64   `gorm:"type:decimal(10,2)"`
	TemporaryCnt int
	ShortTermCnt int
	PermanentCnt int
}

type MaintenanceRecord struct {
	ID          uint      `gorm:"primaryKey"`
	SpotID      uint      `gorm:"not null"`
	Description string    `gorm:"type:text"`
	ReportedBy  uint      `gorm:"not null"`
	Status      string    `gorm:"type:varchar(20);default:'pending'"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	ResolvedAt  *time.Time
}
