// internal/repositories/lease_repo.go
package repositories

import (
	"context"
	"modules/internal/models"
	"time"

	"gorm.io/gorm"
)

type LeaseRepository interface {
	CreateLease(ctx context.Context, lease *models.LeaseOrder) error
	GetUserLeases(ctx context.Context, userID uint, status models.LeaseStatus) ([]*models.LeaseOrder, error)
	UpdateLeaseStatus(ctx context.Context, leaseID uint, status models.LeaseStatus) error
	GetExpiringLeases(ctx context.Context, before time.Time) ([]*models.LeaseOrder, error)
}

type leaseRepo struct {
	db *gorm.DB
}

func NewLeaseRepo(db *gorm.DB) LeaseRepository {
	return &leaseRepo{db: db}
}

func (r *leaseRepo) CreateLease(ctx context.Context, lease *models.LeaseOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查车位是否可租用
		var spot models.ParkingSpot
		if err := tx.First(&spot, lease.SpotID).Error; err != nil {
			return err
		}

		// 注释掉类型检查逻辑
		// if spot.Type != string(models.ShortTerm) {
		//     return errors.New("车位类型不是短租类型")
		// }

		// 更新车位到期时间
		if err := tx.Model(&spot).Update("expires_at", lease.EndDate).Error; err != nil {
			return err
		}

		return tx.Create(lease).Error
	})
}

func (r *leaseRepo) GetUserLeases(ctx context.Context, userID uint, status models.LeaseStatus) ([]*models.LeaseOrder, error) {
	var leases []*models.LeaseOrder
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&leases).Error
	return leases, err
}

func (r *leaseRepo) UpdateLeaseStatus(ctx context.Context, leaseID uint, status models.LeaseStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.LeaseOrder{}).
		Where("id = ?", leaseID).
		Update("status", status).
		Error
}

func (r *leaseRepo) GetExpiringLeases(ctx context.Context, before time.Time) ([]*models.LeaseOrder, error) {
	var leases []*models.LeaseOrder
	err := r.db.WithContext(ctx).
		Where("end_date < ? AND status = ?", before, models.LeaseActive).
		Find(&leases).Error
	return leases, err
}
