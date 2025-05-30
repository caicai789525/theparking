package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"modules/config"
	"modules/internal/controllers"
	"modules/internal/repositories"
	"modules/internal/routes"
	"modules/internal/services"
	"modules/pkg/database"
	"modules/pkg/logger"
	"os"
	"path/filepath"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		// 处理配置加载失败的情况
		log.Fatalf("加载配置失败: %v", err)
	}

	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前工作目录失败: %v", err)
	}

	// 构建日志文件路径
	logDir := filepath.Join(dir, "log")
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}
	logPath := filepath.Join(logDir, "parking.log")
	// 将日志文件路径设置到配置中
	cfg.LogFilePath = logPath

	// 使用配置初始化日志记录器
	logger.InitLogger(cfg)
	if logger.Log == nil {
		log.Fatal("日志记录器初始化失败，日志记录器为 nil")
	}
	defer func() {
		if logger.Log != nil {
			logger.Log.Sync()
		}
	}()

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)
	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("数据库连接失败", zap.Error(err))
	}

	// 执行数据库迁移
	database.Migrate(db)

	// 初始化 UserRepository
	userRepo := repositories.NewUserRepo(db)

	// 初始化 AuthService，传入 UserRepository 和配置
	authService := services.NewAuthService(userRepo, cfg)

	// 初始化控制器
	ctrls := initializeControllers(db, cfg)

	// 创建 Gin 引擎
	router := gin.Default()

	// 挂载 CORS 中间件
	router.Use(CORSMiddleware())

	// 初始化路由依赖，注入 authService
	deps := &routes.RouterDependencies{
		AuthService:    authService,
		AuthController: ctrls.AuthController,
		ParkingService: ctrls.ParkingController,
		AdminService:   ctrls.AdminController,
		LeaseService:   ctrls.LeaseController,
		ReportService:  ctrls.ReportController,
		VehicleService: ctrls.VehicleController,
		OwnerService:   ctrls.OwnerController,
		Cfg:            ctrls.Cfg,
	}

	// 设置路由
	routes.SetupRouter(router, deps)

	// 打印所有注册的路由，用于调试
	for _, route := range router.Routes() {
		logger.Log.Info("Registered Route",
			zap.String("Method", route.Method),
			zap.String("Path", route.Path),
			zap.String("Handler", route.Handler))
		logrus.Infof("Method: %s, Path: %s, Handler: %s", route.Method, route.Path, route.Handler)
	}

	// 健康检查接口
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 获取端口号
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}
	logger.Log.Info("服务将启动在端口", zap.String("port", port))
	logrus.Infof("服务将启动在端口 %s", port)
	if err := router.Run(":" + port); err != nil {
		logger.Log.Fatal("服务启动失败", zap.Error(err))
		logrus.Fatalf("服务启动失败: %v", err)
	}
}

// initializeControllers 初始化控制器
func initializeControllers(db *gorm.DB, cfg *config.Config) *ControllerDependencies {
	// Repos
	userRepo := repositories.NewUserRepo(db)
	parkingRepo := repositories.NewParkingRepo(db)
	purchaseRepo := repositories.NewPurchaseRepo(db)
	reportRepo := repositories.NewReportRepo(db)
	leaseRepo := repositories.NewLeaseRepo(db)
	vehicleRepo := repositories.NewVehicleRepo(db)

	// Services
	authService := services.NewAuthService(userRepo, cfg)
	parkingService := services.NewParkingService(parkingRepo, userRepo)
	ownerService := services.NewOwnerService(parkingRepo, userRepo, purchaseRepo)
	reportService := services.NewReportService(reportRepo, parkingRepo)
	leaseService := services.NewLeaseService(leaseRepo, parkingRepo)
	vehicleService := services.NewVehicleService(vehicleRepo, userRepo, parkingRepo, leaseService)

	// Controllers
	return &ControllerDependencies{
		AuthController:    controllers.NewAuthController(authService),
		ParkingController: controllers.NewParkingController(parkingService),
		AdminController:   controllers.NewAdminController(parkingService, reportService),
		LeaseController:   controllers.NewLeaseController(leaseService),
		ReportController:  controllers.NewReportController(reportService),
		VehicleController: controllers.NewVehicleController(vehicleService),
		OwnerController:   controllers.NewOwnerController(ownerService),
		Cfg:               cfg,
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ControllerDependencies 控制器依赖
type ControllerDependencies struct {
	AuthController    *controllers.AuthController
	ParkingController *controllers.ParkingController
	AdminController   *controllers.AdminController
	LeaseController   *controllers.LeaseController
	ReportController  *controllers.ReportController
	VehicleController *controllers.VehicleController
	OwnerController   *controllers.OwnerController
	Cfg               *config.Config
}
