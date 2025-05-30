// internal/services/vehicle_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"modules/internal/models"
	"modules/internal/repositories"
)

type VehicleService struct {
	repo         repositories.VehicleRepository
	userRepo     repositories.UserRepository
	parkingRepo  repositories.ParkingRepository
	leaseService *LeaseService
}

func NewVehicleService(
	vr repositories.VehicleRepository,
	ur repositories.UserRepository,
	pr repositories.ParkingRepository,
	ls *LeaseService,
) *VehicleService {
	return &VehicleService{
		repo:         vr,
		userRepo:     ur,
		parkingRepo:  pr,
		leaseService: ls,
	}
}

func (s *VehicleService) PublishSpotForRent(ctx context.Context, userID, spotID uint, rate float64, period int) (*models.LeaseOrder, error) {
	spot, err := s.parkingRepo.GetSpotByID(ctx, spotID)
	if err != nil {
		return nil, fmt.Errorf("获取车位信息失败: %w", err)
	}

	// 判断 OwnerID 是否为 0
	if spot.OwnerID == 0 {
		return nil, errors.New("车位无业主，无法出租")
	}

	// 判断操作的用户是否为车位业主
	if spot.OwnerID != userID {
		return nil, errors.New("无权操作该车位")
	}

	return s.leaseService.CreateLease(ctx, userID, spotID, period, rate)
}

func (s *VehicleService) BindVehicle(ctx context.Context, userID uint, license, brand, model string) (*models.Vehicle, error) {
	vehicle := &models.Vehicle{
		UserID:       userID,
		LicensePlate: license,
		Brand:        brand,
		Model:        model,
	}

	if err := s.repo.AddVehicle(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("绑定车辆失败: %w", err)
	}
	return vehicle, nil
}
