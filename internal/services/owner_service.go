// internal/services/owner_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"modules/internal/models"
	"modules/internal/repositories"
	"time"

	"gorm.io/gorm"
)

type OwnerService struct {
	parkingRepo  repositories.ParkingRepository
	userRepo     repositories.UserRepository
	purchaseRepo repositories.PurchaseRepository // 新增
}

func NewOwnerService(
	pr repositories.ParkingRepository,
	ur repositories.UserRepository,
	pur repositories.PurchaseRepository, // 新增
) *OwnerService {
	return &OwnerService{
		parkingRepo:  pr,
		userRepo:     ur,
		purchaseRepo: pur, // 新增
	}
}

// 购置永久车位核心逻辑
func (s *OwnerService) PurchasePermanentSpot(
	ctx context.Context,
	userID uint,
	spotID uint,
	price float64,
) (*models.ParkingSpot, error) {
	// 1. 验证车位是否存在且类型可转换
	spot, err := s.parkingRepo.GetSpotByID(ctx, spotID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("车位不存在")
		}
		return nil, fmt.Errorf("查询车位失败: %w", err)
	}

	// 2. 仅允许转换临时或短租车位
	// 将 spot.Type 转换为 ParkingType 类型
	spotType := models.ParkingType(spot.Type)
	if spotType != models.Temporary && spotType != models.ShortTerm {
		return nil, errors.New("该车位类型不可购置为永久车位")
	}

	// 3. 更新车位信息
	if err := s.parkingRepo.UpdateSpot(ctx, &models.ParkingSpot{
		ID: spotID,
		// 将 models.Permanent 转换为 string 类型
		Type: string(models.Permanent),
		// 将 models.Idle 转换为 string 类型
		Status: string(models.Idle),
		// 直接使用 userID，因为 OwnerID 是 uint 类型
		OwnerID: userID,
	}); err != nil {
		return nil, fmt.Errorf("更新车位失败: %w", err)
	}

	// 4. 创建购置记录（需实现 PurchaseRecord 模型）
	record := &models.PurchaseRecord{
		UserID:        userID, // 使用 uint 类型的 userID
		SpotID:        spotID,
		PurchasePrice: price,
		PurchaseDate:  time.Now(),
	}
	if err := s.purchaseRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("创建购置记录失败: %w", err)
	}

	return spot, nil
}
