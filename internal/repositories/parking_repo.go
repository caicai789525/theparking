// internal/repositories/parking_repo.go
package repositories

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
	"modules/internal/models"
	"modules/pkg/logger"
	"time"

	"gorm.io/gorm"
)

type ParkingRepository interface {
	CreateSpot(ctx context.Context, spot *models.ParkingSpot) error
	GetSpotByID(ctx context.Context, id uint) (*models.ParkingSpot, error)
	GetUserSpots(ctx context.Context, userID uint) ([]*models.ParkingSpot, error)
	UpdateSpot(ctx context.Context, spot *models.ParkingSpot) error
	DeleteSpot(ctx context.Context, id uint) error
	ListSpots(ctx context.Context, filter SpotFilter) ([]*models.ParkingSpot, error)
	CreateRecord(ctx context.Context, record *models.ParkingRecord) error
	GetOngoingRecord(ctx context.Context, license string) (*models.ParkingRecord, error)
	UpdateStatus(ctx context.Context, spotID uint, status models.ParkingStatus) error
	UpdateSpotExpiry(ctx context.Context, spotID uint, expiresAt *time.Time) error
	OccupySpot(ctx context.Context, spotID uint, license string, userID *uint) (*models.ParkingRecord, error)
	ReleaseSpot(ctx context.Context, recordID uint) (*models.ParkingRecord, error)
	UpdateRecord(ctx context.Context, record *models.ParkingRecord) (*models.ParkingRecord, error)
	GetParkingByID(ctx context.Context, parkingID uint) (*models.ParkingRecord, error)
	UpdateParking(ctx context.Context, parking *models.ParkingRecord) error
	GetParkingSpotByID(ctx context.Context, parkingID uint) (*models.ParkingSpot, error)
}

type parkingRepo struct {
	db *gorm.DB
}

func NewParkingRepo(db *gorm.DB) ParkingRepository {
	return &parkingRepo{db: db}
}

func (r *parkingRepo) GetUserSpots(ctx context.Context, userID uint) ([]*models.ParkingSpot, error) {
	var spots []*models.ParkingSpot
	err := r.db.WithContext(ctx).Where("owner_id = ?", userID).Find(&spots).Error
	return spots, err
}

func (r *parkingRepo) CreateSpot(ctx context.Context, spot *models.ParkingSpot) error {
	return r.db.WithContext(ctx).Create(spot).Error
}

func (r *parkingRepo) GetSpotByID(ctx context.Context, id uint) (*models.ParkingSpot, error) {
	var spot models.ParkingSpot
	err := r.db.WithContext(ctx).First(&spot, id).Error
	return &spot, err
}

func (r *parkingRepo) UpdateSpot(ctx context.Context, spot *models.ParkingSpot) error {
	return r.db.WithContext(ctx).Save(spot).Error
}

type SpotFilter struct {
	Type      models.ParkingType
	Status    models.ParkingStatus
	OwnerID   uint
	UpdatedAt *time.Time
}

func (r *parkingRepo) OccupySpot(
	ctx context.Context,
	spotID uint,
	license string,
	userID *uint,
) (*models.ParkingRecord, error) { // 修改返回类型
	var record *models.ParkingRecord // 用于保存创建的记录

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取并锁定车位
		var spot models.ParkingSpot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&spot, spotID).Error; err != nil {
			return err
		}

		if spot.Status != "Idle" {
			return errors.New("停车位不可用")
		}

		// 创建停车记录
		newRecord := &models.ParkingRecord{
			SpotID:    spotID,
			UserID:    userID,
			License:   license,
			EntryTime: time.Now(),
		}

		if err := tx.Create(newRecord).Error; err != nil {
			return err
		}

		// 更新车位状态
		if err := tx.Model(&spot).Updates(map[string]interface{}{
			"status":  models.Occupied,
			"license": license,
		}).Error; err != nil {
			return err
		}

		record = newRecord // 保存创建的记录
		return nil
	})

	return record, err // 返回记录和错误
}

func (r *parkingRepo) ReleaseSpot(ctx context.Context, recordID uint) (*models.ParkingRecord, error) {
	var record models.ParkingRecord
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用正确的锁语法
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&record, recordID).Error; err != nil {
			return err
		}

		if record.IsCompleted {
			return errors.New("停车记录已完成")
		}

		// 更新出场时间
		exitTime := time.Now()
		record.ExitTime = &exitTime
		record.IsCompleted = true

		// 更新车位状态
		if err := tx.Model(&models.ParkingSpot{}).
			Where("id = ?", record.SpotID).
			Updates(map[string]interface{}{
				"status":  models.Idle,
				"license": "",
			}).Error; err != nil {
			return err
		}

		return tx.Save(&record).Error
	})

	return &record, err
}
func (r *parkingRepo) DeleteSpot(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ParkingSpot{}, id).Error
}

func (r *parkingRepo) ListSpots(ctx context.Context, filter SpotFilter) ([]*models.ParkingSpot, error) {
	query := r.db.WithContext(ctx).Model(&models.ParkingSpot{})

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.OwnerID != 0 {
		query = query.Where("owner_id = ?", filter.OwnerID)
	}

	var spots []*models.ParkingSpot
	err := query.Find(&spots).Error
	return spots, err
}

func (r *parkingRepo) CreateRecord(ctx context.Context, record *models.ParkingRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetParkingSpotByID 根据车位 ID 获取车位信息
func (r *parkingRepo) GetParkingSpotByID(ctx context.Context, parkingID uint) (*models.ParkingSpot, error) {
	var parkingSpot models.ParkingSpot
	err := r.db.WithContext(ctx).First(&parkingSpot, parkingID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrParkingNotFound
		}
		return nil, err
	}
	return &parkingSpot, nil
}

func (r *parkingRepo) GetOngoingRecord(ctx context.Context, license string) (*models.ParkingRecord, error) {
	var record models.ParkingRecord
	err := r.db.WithContext(ctx).
		Where("license = ? AND is_completed = ?", license, false).
		First(&record).Error
	return &record, err
}

func (r *parkingRepo) UpdateRecord(
	ctx context.Context,
	record *models.ParkingRecord,
) (*models.ParkingRecord, error) { // 添加返回的 ParkingRecord
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return nil, err
	}
	return record, nil // 返回更新后的记录
}

func (r *parkingRepo) UpdateStatus(ctx context.Context, spotID uint, status models.ParkingStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.ParkingSpot{}).
		Where("id = ?", spotID).
		Update("status", status).
		Error
}

func (r *parkingRepo) UpdateSpotExpiry(
	ctx context.Context,
	spotID uint,
	expiresAt *time.Time,
) error {
	// 记录调试日志
	logger.Log.Debug("开始更新车位过期时间",
		zap.Uint("spotID", spotID),
		zap.Any("expiresAt", expiresAt))

	// 执行数据库更新操作
	result := r.db.WithContext(ctx).
		Model(&models.ParkingSpot{}).   // 指定操作模型
		Where("id = ?", spotID).        // 定位具体记录
		Update("expires_at", expiresAt) // 更新过期时间字段

	// 错误处理流程
	if result.Error != nil {
		logger.Log.Error("数据库更新失败",
			zap.Uint("spotID", spotID),
			zap.Error(result.Error))
		return fmt.Errorf("更新车位信息失败: %w", result.Error)
	}

	// 检查实际更新的记录数
	if result.RowsAffected == 0 {
		logger.Log.Warn("未找到目标车位",
			zap.Uint("spotID", spotID))
		return errors.New("指定的车位不存在")
	}

	// 记录成功日志
	logger.Log.Info("成功更新车位过期时间",
		zap.Uint("spotID", spotID),
		zap.Any("newExpiry", expiresAt))

	return nil
}

func (r *parkingRepo) Transaction(
	ctx context.Context,
	fn func(repo ParkingRepository) error,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&parkingRepo{db: tx})
	})
}

// 修正方法接收者类型和返回类型
func (r *parkingRepo) GetParkingByID(ctx context.Context, parkingID uint) (*models.ParkingRecord, error) {
	var parking models.ParkingRecord
	err := r.db.WithContext(ctx).First(&parking, parkingID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 记录不存在，返回自定义错误
			return nil, models.ErrParkingNotFound
		}
		return nil, err
	}
	return &parking, nil
}

// 修正方法接收者类型
func (r *parkingRepo) UpdateParking(ctx context.Context, parking *models.ParkingRecord) error {
	return r.db.WithContext(ctx).Save(parking).Error
}
