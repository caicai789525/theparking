// internal/services/report_service.go
package services

import (
	"context"
	"fmt"
	"modules/internal/models"
	"modules/internal/repositories"
	"time"
)

type ReportService struct {
	reportRepo  repositories.ReportRepository
	parkingRepo repositories.ParkingRepository // 新增停车仓库依赖
}

func NewReportService(
	rr repositories.ReportRepository,
	pr repositories.ParkingRepository, // ✅ 需要添加 parkingRepo
) *ReportService {
	return &ReportService{
		reportRepo:  rr,
		parkingRepo: pr, // 需要停车仓库来获取车位数据
	}
}

func (s *ReportService) GenerateDailyReport(ctx context.Context, days int) (*models.DailyReport, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	reports, err := s.reportRepo.GetDailyReports(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("获取日报表数据失败: %w", err)
	}

	var total models.DailyReport
	for _, r := range reports {
		total.TotalIncome += r.TotalIncome
		total.TemporaryCnt += r.TemporaryCnt
		total.ShortTermCnt += r.ShortTermCnt
		total.PermanentCnt += r.PermanentCnt
	}
	return &total, nil
}

func (s *ReportService) GetSpotStats(ctx context.Context) (map[string]interface{}, error) {
	// 获取车位利用率
	utilization, err := s.reportRepo.GetSpotUtilization(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取利用率数据失败: %w", err)
	}

	// 获取所有车位
	spots, err := s.parkingRepo.ListSpots(ctx, repositories.SpotFilter{})
	if err != nil {
		return nil, fmt.Errorf("获取车位列表失败: %w", err)
	}

	// 统计总数和可用数
	var total, available int
	for _, spot := range spots {
		total++
		if spot.Status == "Idle" {
			available++
		}
	}

	return map[string]interface{}{
		"total_spots":       total,
		"available_spots":   available,
		"utilization_rates": utilization,
	}, nil
}
