// corn/corn.go
package cron

import (
	"context"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"modules/internal/repositories"
	"modules/internal/services"
	"modules/pkg/logger"
)

func StartCronJobs(db *gorm.DB) {
	c := cron.New()

	// 初始化仓库
	parkingRepo := repositories.NewParkingRepo(db)
	userRepo := repositories.NewUserRepo(db)
	leaseRepo := repositories.NewLeaseRepo(db)
	reportRepo := repositories.NewReportRepo(db)

	// 每天凌晨1点执行
	c.AddFunc("0 1 * * *", func() {
		ctx := context.Background()

		// 初始化租赁服务（移除了支付和通知依赖）
		leaseService := services.NewLeaseService(
			leaseRepo,
			parkingRepo,
		)

		// 过期租赁处理
		if err := leaseService.CheckLeaseExpirations(ctx); err != nil {
			logger.Log.Error("处理过期租赁失败", zap.Error(err))
		}

		// 初始化报表服务
		reportService := services.NewReportService(reportRepo, parkingRepo)

		// 生成日报表
		if _, err := reportService.GenerateDailyReport(ctx, 1); err != nil {
			logger.Log.Error("生成日报表失败", zap.Error(err))
		}
	})

	// 每小时检查车位状态
	c.AddFunc("@hourly", func() {
		ctx := context.Background()
		// 初始化停车服务
		parkingService := services.NewParkingService(
			parkingRepo,
			userRepo,
		)

		if err := parkingService.CheckFaultySpots(ctx); err != nil {
			logger.Log.Error("检查故障车位失败", zap.Error(err))
		}
	})

	c.Start()
}
