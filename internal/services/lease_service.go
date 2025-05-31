// internal/services/lease_service.go
package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"modules/internal/models"
	"modules/internal/repositories"
	"modules/pkg/logger"
	"time"
)

// CreateLease 创建租赁订单
func (s *LeaseService) CreateLease(
	ctx context.Context,
	userID uint,
	spotID uint,
	period int, // 租赁时长（月数）
	rate float64, // 租赁费率
) (*models.LeaseOrder, error) {
	// 参数校验
	if period <= 0 {
		err := fmt.Errorf("租赁时长必须为正整数")
		logger.Log.Error("创建租赁订单失败",
			zap.Uint("userID", userID),
			zap.Uint("spotID", spotID),
			zap.Int("period", period),
			zap.Error(err))
		return nil, err
	}

	if rate <= 0 {
		err := fmt.Errorf("租赁费率必须为正数")
		logger.Log.Error("创建租赁订单失败",
			zap.Uint("userID", userID),
			zap.Uint("spotID", spotID),
			zap.Float64("rate", rate),
			zap.Error(err))
		return nil, err
	}

	logger.Log.Info("Starting to create lease",
		zap.Uint("userID", userID),
		zap.Uint("spotID", spotID),
		zap.Int("period", period),
		zap.Float64("rate", rate))

	// 计算总价时使用传入的费率
	totalPrice := rate * float64(period)

	startDate := time.Now()
	endDate := startDate.AddDate(0, period, 0)

	// 检查结束时间是否早于开始时间
	if endDate.Before(startDate) {
		err := fmt.Errorf("结束时间早于开始时间")
		logger.Log.Error("创建租赁订单失败",
			zap.Uint("userID", userID),
			zap.Uint("spotID", spotID),
			zap.Time("startDate", startDate),
			zap.Time("endDate", endDate),
			zap.Error(err))
		return nil, err
	}

	lease := &models.LeaseOrder{
		UserID:     userID,
		SpotID:     spotID,
		StartDate:  startDate,
		EndDate:    endDate,
		TotalPrice: totalPrice,
		Status:     models.LeaseActive,
	}

	// 使用事务创建租赁订单和更新车位信息
	err := s.leaseRepo.Transaction(ctx, func(tx repositories.LeaseTx) error {
		if err := tx.CreateLease(ctx, lease); err != nil {
			return fmt.Errorf("创建租赁订单失败: %w", err)
		}

		// 更新车位到期时间
		if err := s.parkingRepo.UpdateSpotExpiry(ctx, spotID, &endDate); err != nil {
			return fmt.Errorf("更新车位到期时间失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Log.Error("创建租赁订单失败",
			zap.Uint("userID", userID),
			zap.Uint("spotID", spotID),
			zap.Error(err))
		return nil, err
	}

	logger.Log.Info("Lease created successfully",
		zap.Uint("userID", userID),
		zap.Uint("spotID", spotID))

	return lease, nil
}

type LeaseService struct {
	leaseRepo   repositories.LeaseRepository
	parkingRepo repositories.ParkingRepository
}

func NewLeaseService(
	lr repositories.LeaseRepository,
	pr repositories.ParkingRepository,
) *LeaseService {
	return &LeaseService{
		leaseRepo:   lr,
		parkingRepo: pr,
	}
}

func (s *LeaseService) CheckLeaseExpirations(ctx context.Context) error {
	expiringLeases, err := s.leaseRepo.GetExpiringLeases(ctx, time.Now().Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("获取到期租赁失败: %w", err)
	}

	for _, lease := range expiringLeases {
		// 直接标记租赁过期
		if err := s.leaseRepo.UpdateLeaseStatus(ctx, lease.ID, models.LeaseExpired); err != nil {
			logger.Log.Error("更新租赁状态失败",
				zap.Uint("leaseID", lease.ID),
				zap.Error(err))
			continue
		}

		// 释放关联车位
		if err := s.parkingRepo.UpdateSpotExpiry(ctx, lease.SpotID, nil); err != nil {
			logger.Log.Error("释放车位失败",
				zap.Uint("spotID", lease.SpotID),
				zap.Error(err))
		}
	}
	return nil
}
