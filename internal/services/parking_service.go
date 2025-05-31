// internal/services/parking_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"modules/internal/models"
	"modules/internal/repositories"
	"modules/pkg/logger"
	"time"
)

type ParkingService struct {
	parkingRepo repositories.ParkingRepository
	userRepo    repositories.UserRepository
	Notes       string `gorm:"type:text"`
}

func NewParkingService(
	pr repositories.ParkingRepository,
	ur repositories.UserRepository,
) *ParkingService {
	return &ParkingService{
		parkingRepo: pr,
		userRepo:    ur,
	}
}

// 计算停车费用
func (s *ParkingService) CalculateFee(record *models.ParkingRecord, spot *models.ParkingSpot) float64 {
	if record.ExitTime == nil {
		return 0
	}

	duration := record.ExitTime.Sub(record.EntryTime).Hours()

	// 将 spot.Type 与 string 类型的枚举值进行比较
	switch string(spot.Type) {
	case string(models.ShortTerm):
		// 当 ExpiresAt 不为空字符串时进行解析
		if spot.ExpiresAt != "" {
			// 假设 ExpiresAt 格式为 RFC3339，可根据实际情况调整
			expiresAt, err := time.Parse(time.RFC3339, spot.ExpiresAt)
			if err == nil && time.Now().After(expiresAt) {
				return duration * spot.HourlyRate
			}
		}
		return 0
	case string(models.Temporary):
		return duration * spot.HourlyRate
	default:
		return 0
	}
}

// 处理车辆入场
func (s *ParkingService) ProcessEntry(ctx context.Context, license string, userID *uint) (*models.ParkingRecord, error) {
	// 检查是否有进行中的记录
	existing, err := s.parkingRepo.GetOngoingRecord(ctx, license)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询进行中记录失败: %w", err)
	}
	if existing != nil {
		return nil, errors.New("该车辆已有进行中的停车记录")
	}

	// 自动分配临时车位
	spot, err := s.findAvailableSpot(ctx, models.Temporary)
	if err != nil {
		return nil, fmt.Errorf("分配车位失败: %w", err)
	}

	// 修正：显式接收两个返回值并处理错误
	record, err := s.parkingRepo.OccupySpot(ctx, spot.ID, license, userID)
	if err != nil {
		return nil, fmt.Errorf("占用车位失败: %w", err)
	}
	return record, nil
}

// 处理车辆出场
func (s *ParkingService) ProcessExit(ctx context.Context, recordID uint) (*models.ParkingRecord, error) {
	// 修正：接收 ReleaseSpot 的两个返回值
	record, err := s.parkingRepo.ReleaseSpot(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("释放车位失败: %w", err)
	}

	// 获取关联车位信息
	spot, err := s.parkingRepo.GetSpotByID(ctx, record.SpotID)
	if err != nil {
		return nil, fmt.Errorf("获取车位信息失败: %w", err)
	}

	// 计算并更新费用
	record.TotalCost = s.CalculateFee(record, spot)
	updatedRecord, err := s.parkingRepo.UpdateRecord(ctx, record)
	if err != nil {
		return nil, fmt.Errorf("更新记录失败: %w", err)
	}
	return updatedRecord, nil
}

// 业主车辆特殊入场处理
func (s *ParkingService) ProcessOwnerEntry(ctx context.Context, userID uint, license string) (*models.ParkingRecord, error) {
	spots, err := s.parkingRepo.ListSpots(ctx, repositories.SpotFilter{
		OwnerID: userID,
		Status:  models.Idle,
	})
	if err != nil {
		return nil, fmt.Errorf("查询业主车位失败: %w", err)
	}
	if len(spots) == 0 {
		return nil, errors.New("业主没有可用车位")
	}

	// 修正：接收 OccupySpot 的两个返回值
	record, err := s.parkingRepo.OccupySpot(ctx, spots[0].ID, license, &userID)
	if err != nil {
		return nil, fmt.Errorf("占用车位失败: %w", err)
	}
	return record, nil
}

// 更新车位状态
func (s *ParkingService) UpdateSpotStatus(
	ctx context.Context,
	spotID uint,
	status models.ParkingStatus,
	notes string, // 添加 notes 参数
) (*models.ParkingSpot, error) {
	spot, err := s.parkingRepo.GetSpotByID(ctx, spotID)
	if err != nil {
		return nil, fmt.Errorf("获取车位失败: %w", err)
	}

	// 将 models.ParkingStatus 类型的 status 转换为 string 类型
	spot.Status = string(status)
	spot.Notes = notes // 假设 models.ParkingSpot 有 Notes 字段

	if err := s.parkingRepo.UpdateSpot(ctx, spot); err != nil {
		return nil, fmt.Errorf("更新状态失败: %w", err)
	}

	return spot, nil
}

// 检查并恢复故障车位
func (s *ParkingService) CheckFaultySpots(ctx context.Context) error {
	threshold := time.Now().Add(-24 * time.Hour)
	// 修正：使用正确的字段 UpdatedAt 替代 UpdatedBefore
	spots, err := s.parkingRepo.ListSpots(ctx, repositories.SpotFilter{
		Status:    models.Faulty,
		UpdatedAt: &threshold,
	})
	if err != nil {
		return fmt.Errorf("查询故障车位失败: %w", err)
	}

	for _, spot := range spots {
		if err := s.parkingRepo.UpdateStatus(ctx, spot.ID, models.Idle); err != nil {
			logger.Log.Error("恢复车位状态失败",
				zap.Uint("spotID", spot.ID),
				zap.Error(err))
			continue
		}
		logger.Log.Info("成功恢复车位状态",
			zap.Uint("spotID", spot.ID))
	}
	return nil
}

// 获取车位列表
func (s *ParkingService) ListSpots(ctx context.Context) ([]*models.ParkingSpot, error) {
	return s.parkingRepo.ListSpots(ctx, repositories.SpotFilter{})
}

// 私有方法：查找可用车位
func (s *ParkingService) findAvailableSpot(
	ctx context.Context,
	spotType models.ParkingType,
) (*models.ParkingSpot, error) {
	spots, err := s.parkingRepo.ListSpots(ctx, repositories.SpotFilter{
		Type:   spotType,
		Status: models.Idle,
	})
	if err != nil {
		return nil, fmt.Errorf("查询可用车位失败: %w", err)
	}
	if len(spots) == 0 {
		return nil, errors.New("当前没有可用车位")
	}
	return spots[0], nil
}

// 创建停车位
func (s *ParkingService) CreateSpot(
	ctx context.Context,
	spot *models.ParkingSpot,
) (*models.ParkingSpot, error) {
	// 设置默认费率
	if spot.HourlyRate == 0 {
		switch spot.Type {
		// 将 models.Temporary 转换为 string 类型
		case string(models.Temporary):
			spot.HourlyRate = 5 // 默认临时车位费率
		// 将 models.ShortTerm 转换为 string 类型
		case string(models.ShortTerm):
			spot.MonthlyRate = 300 // 默认短租月费
		}
	}

	if err := s.parkingRepo.CreateSpot(ctx, spot); err != nil {
		return nil, fmt.Errorf("创建车位失败: %w", err)
	}
	return spot, nil
}

// GetUserSpots 查询用户自己的车位
func (s *ParkingService) GetUserSpots(ctx context.Context, userID uint) ([]*models.ParkingSpot, error) {
	return s.parkingRepo.GetUserSpots(ctx, userID)
}

// BindParkingToUser 管理员将车位绑定给用户
func (s *ParkingService) BindParkingToUser(ctx context.Context, userID, parkingID uint) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return models.ErrUserNotFound
	}

	// 检查车位是否存在
	parking, err := s.parkingRepo.GetParkingByID(ctx, parkingID)
	if err != nil {
		return err
	}
	if parking == nil {
		return models.ErrParkingNotFound
	}

	// 检查车位是否已被绑定
	// 先判断 parking.UserID 是否为 nil，不为 nil 时再解引用比较
	if parking.UserID != nil && *parking.UserID != 0 {
		return models.ErrParkingAlreadyBound
	}

	// 绑定车位给用户
	// 将 uint 类型的 userID 转换为 *uint 类型
	userIDPtr := &userID
	parking.UserID = userIDPtr
	return s.parkingRepo.UpdateParking(ctx, parking)
}

// GetUserInfo 根据用户 ID 查询用户信息
func (s *ParkingService) GetUserInfo(ctx context.Context, userID uint) (*models.UserInfoResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, models.ErrUserNotFound
	}

	return &models.UserInfoResponse{
		ID:       user.ID,
		Username: user.Username, // 假设 User 结构体有 Username 字段
	}, nil
}

// UnbindParkingFromUser 管理员解除车位与用户的绑定
func (s *ParkingService) UnbindParkingFromUser(ctx context.Context, userID, parkingID uint) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return models.ErrUserNotFound
	}

	// 检查车位是否存在
	parking, err := s.parkingRepo.GetParkingByID(ctx, parkingID)
	if err != nil {
		return err
	}
	if parking == nil {
		return models.ErrParkingNotFound
	}

	// 检查车位是否已绑定给该用户
	if parking.UserID == nil || *parking.UserID != userID {
		return errors.New("车位未绑定给该用户")
	}

	// 解除车位与用户的绑定
	parking.UserID = nil
	return s.parkingRepo.UpdateParking(ctx, parking)
}

// GetParkingBindUser 查询车位绑定的用户信息
func (s *ParkingService) GetParkingBindUser(ctx context.Context, parkingID uint) (*models.ParkingBindUserResponse, error) {
	// 检查车位是否存在
	spot, err := s.parkingRepo.GetParkingSpotByID(ctx, parkingID)
	if err != nil {
		return nil, err
	}
	if spot == nil {
		return nil, models.ErrParkingNotFound
	}

	var username string
	// 不再使用指针操作
	if spot.OwnerID != 0 {
		user, err := s.userRepo.GetUserByID(ctx, spot.OwnerID)
		if err != nil {
			return nil, err
		}
		if user != nil {
			username = user.Username
		}
	}

	return &models.ParkingBindUserResponse{
		ParkingID: parkingID,
		UserID:    spot.OwnerID, // 直接赋值 uint 类型
		Username:  username,
	}, nil
}
