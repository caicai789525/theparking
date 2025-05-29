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
	// 计算总价时使用传入的费率
	totalPrice := rate * float64(period)

	lease := &models.LeaseOrder{
		UserID:     userID,
		SpotID:     spotID,
		StartDate:  time.Now(),
		EndDate:    time.Now().AddDate(0, period, 0),
		TotalPrice: totalPrice,
		Status:     models.LeaseActive,
	}

	if err := s.leaseRepo.CreateLease(ctx, lease); err != nil {
		return nil, fmt.Errorf("创建租赁订单失败: %w", err)
	}

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
