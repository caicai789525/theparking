// internal/repositories/lease_repo.go
package repositories

import (
	"context"
	"modules/internal/models"
	"time"

	"gorm.io/gorm"
)

type LeaseTx interface {
	CreateLease(ctx context.Context, lease *models.LeaseOrder) error
}

type LeaseRepository interface {
	CreateLease(ctx context.Context, lease *models.LeaseOrder) error
	GetUserLeases(ctx context.Context, userID uint, status models.LeaseStatus) ([]*models.LeaseOrder, error)
	UpdateLeaseStatus(ctx context.Context, leaseID uint, status models.LeaseStatus) error
	GetExpiringLeases(ctx context.Context, before time.Time) ([]*models.LeaseOrder, error)
	Transaction(ctx context.Context, fn func(tx LeaseTx) error) error
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

// Transaction 实现事务方法
func (r *leaseRepo) Transaction(ctx context.Context, fn func(tx LeaseTx) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &leaseTxRepo{db: tx}
		return fn(txRepo)
	})
}

// leaseTxRepo 实现 LeaseTx 接口
type leaseTxRepo struct {
	db *gorm.DB
}

// CreateLease 实现 LeaseTx 接口的 CreateLease 方法
func (t *leaseTxRepo) CreateLease(ctx context.Context, lease *models.LeaseOrder) error {
	return t.db.WithContext(ctx).Create(lease).Error
}
