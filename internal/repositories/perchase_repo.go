// internal/repositories/purchase_repo.go
package repositories

import (
	"context"
	"gorm.io/gorm"
	"modules/internal/models"
)

type PurchaseRepository interface {
	Create(ctx context.Context, record *models.PurchaseRecord) error
	GetByUserID(ctx context.Context, userID uint) ([]*models.PurchaseRecord, error)
}

type purchaseRepo struct {
	db *gorm.DB
}

func NewPurchaseRepo(db *gorm.DB) PurchaseRepository {
	return &purchaseRepo{db: db}
}

func (r *purchaseRepo) Create(ctx context.Context, record *models.PurchaseRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *purchaseRepo) GetByUserID(ctx context.Context, userID uint) ([]*models.PurchaseRecord, error) {
	var records []*models.PurchaseRecord
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error
	return records, err
}
