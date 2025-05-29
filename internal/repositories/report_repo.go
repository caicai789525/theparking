// internal/repositories/report_repo.go
package repositories

import (
	"context"
	"modules/internal/models"
	"time"

	"gorm.io/gorm"
)

type ReportRepository interface {
	GetDailyReports(ctx context.Context, start, end time.Time) ([]*models.DailyReport, error)
	GetSpotUtilization(ctx context.Context) (map[models.ParkingType]float64, error)
	GetUserActivities(ctx context.Context, userID uint) ([]*models.ParkingRecord, error)
	// 维护记录
	CreateMaintenance(ctx context.Context, record *models.MaintenanceRecord) error
	ResolveMaintenance(ctx context.Context, id uint) error
	GetMaintenanceRecords(ctx context.Context, status string) ([]*models.MaintenanceRecord, error)
}

type reportRepo struct {
	db *gorm.DB
}

func NewReportRepo(db *gorm.DB) ReportRepository {
	return &reportRepo{db: db}
}

func (r *reportRepo) GetDailyReports(ctx context.Context, start, end time.Time) ([]*models.DailyReport, error) {
	var reports []*models.DailyReport

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			DATE(exit_time) AS date,
			COALESCE(SUM(total_cost), 0) AS total_income,
			COUNT(CASE WHEN spots.type = 'temporary' THEN 1 END) AS temporary_cnt,
			COUNT(CASE WHEN spots.type = 'short_term' THEN 1 END) AS short_term_cnt,
			COUNT(CASE WHEN spots.type = 'permanent' THEN 1 END) AS permanent_cnt
		FROM parking_records
		JOIN parking_spots ON parking_spots.id = parking_records.spot_id
		WHERE exit_time BETWEEN ? AND ?
		GROUP BY DATE(exit_time)
		ORDER BY date DESC
	`, start, end).Scan(&reports).Error

	return reports, err
}

func (r *reportRepo) GetSpotUtilization(ctx context.Context) (map[models.ParkingType]float64, error) {
	var stats []struct {
		Type  models.ParkingType
		Total int
		Idle  int
	}

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			type,
			COUNT(*) AS total,
			COUNT(CASE WHEN status = 'idle' THEN 1 END) AS idle
		FROM parking_spots
		GROUP BY type
	`).Scan(&stats).Error

	result := make(map[models.ParkingType]float64)
	for _, s := range stats {
		if s.Total > 0 {
			result[s.Type] = (1 - float64(s.Idle)/float64(s.Total)) * 100
		}
	}
	return result, err
}

func (r *reportRepo) HasPendingMaintenance(ctx context.Context, spotID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.MaintenanceRecord{}).
		Where("spot_id = ? AND status = ?", spotID, "pending").
		Count(&count).Error
	return count > 0, err
}

func (r *reportRepo) CreateMaintenance(ctx context.Context, record *models.MaintenanceRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}
func (r *reportRepo) GetUserActivities(ctx context.Context, userID uint) ([]*models.ParkingRecord, error) {
	var records []*models.ParkingRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("entry_time DESC").
		Find(&records).Error
	return records, err
}

func (r *reportRepo) ResolveMaintenance(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.MaintenanceRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": time.Now(),
		}).Error
}

func (r *reportRepo) GetMaintenanceRecords(ctx context.Context, status string) ([]*models.MaintenanceRecord, error) {
	var records []*models.MaintenanceRecord
	query := r.db.WithContext(ctx)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Find(&records).Error
	return records, err
}
