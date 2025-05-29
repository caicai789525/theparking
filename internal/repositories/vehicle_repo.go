// internal/repositories/vehicle_repo.go
package repositories

import (
	"context"
	"errors"
	"modules/internal/models"

	"gorm.io/gorm"
)

type VehicleRepository interface {
	AddVehicle(ctx context.Context, vehicle *models.Vehicle) error
	RemoveVehicle(ctx context.Context, userID, vehicleID uint) error
	GetUserVehicles(ctx context.Context, userID uint) ([]*models.Vehicle, error)
	GetVehicleByLicense(ctx context.Context, license string) (*models.Vehicle, error)
}

type vehicleRepo struct {
	db *gorm.DB
}

func NewVehicleRepo(db *gorm.DB) VehicleRepository {
	return &vehicleRepo{db: db}
}

func (r *vehicleRepo) AddVehicle(ctx context.Context, vehicle *models.Vehicle) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查是否已存在
		var count int64
		if err := tx.Model(&models.Vehicle{}).
			Where("license_plate = ?", vehicle.LicensePlate).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("车辆已存在")
		}

		// 如果是第一辆车则设为默认
		var existingCount int64
		tx.Model(&models.Vehicle{}).Where("user_id = ?", vehicle.UserID).Count(&existingCount)
		if existingCount == 0 {
			vehicle.IsDefault = true
		}

		return tx.Create(vehicle).Error
	})
}

func (r *vehicleRepo) RemoveVehicle(ctx context.Context, userID, vehicleID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 验证车辆所有权
		var vehicle models.Vehicle
		if err := tx.Where("id = ? AND user_id = ?", vehicleID, userID).First(&vehicle).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("车辆不存在或无权操作")
			}
			return err
		}

		// 删除车辆
		if err := tx.Delete(&vehicle).Error; err != nil {
			return err
		}

		// 如果删除的是默认车辆，设置新的默认车辆
		if vehicle.IsDefault {
			return tx.Model(&models.Vehicle{}).
				Where("user_id = ?", userID).
				Order("created_at ASC"). // 选择最早添加的车辆
				Limit(1).
				Update("is_default", true).Error
		}
		return nil
	})
}

func (r *vehicleRepo) GetUserVehicles(ctx context.Context, userID uint) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepo) GetVehicleByLicense(ctx context.Context, license string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	err := r.db.WithContext(ctx).
		Where("license_plate = ?", license).
		First(&vehicle).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("车辆不存在")
	}
	return &vehicle, err
}
