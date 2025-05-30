package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
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

	"gorm.io/gorm"
)

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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("進入了 CORS Middleware, method:", c.Request.Method)

		// 获取 Origin
		origin := c.GetHeader("Origin")
		fmt.Println("Request Origin:", origin) // 确保 Origin 正确输出

		// 只允许特定的域名，这里设定为 http://localhost:63344，您可以根据需要调整

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// 处理 OPTIONS 请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// 继续请求处理
		c.Next()
	}
}

func initializeControllers(db *gorm.DB, cfg *config.Config) *ControllerDependencies {
	// Repos
	userRepo := repositories.NewUserRepo(db)
	fmt.Printf("UserRepo initialized: %+v\n", userRepo)
	parkingRepo := repositories.NewParkingRepo(db)
	fmt.Printf("ParkingRepo initialized: %+v\n", parkingRepo)
	purchaseRepo := repositories.NewPurchaseRepo(db)
	fmt.Printf("PurchaseRepo initialized: %+v\n", purchaseRepo)
	reportRepo := repositories.NewReportRepo(db)
	fmt.Printf("ReportRepo initialized: %+v\n", reportRepo)
	leaseRepo := repositories.NewLeaseRepo(db)
	fmt.Printf("LeaseRepo initialized: %+v\n", leaseRepo)
	vehicleRepo := repositories.NewVehicleRepo(db)
	fmt.Printf("VehicleRepo initialized: %+v\n", vehicleRepo)

	// Services
	authService := services.NewAuthService(userRepo, cfg)
	fmt.Printf("AuthService initialized: %+v\n", authService)
	parkingService := services.NewParkingService(parkingRepo, userRepo)
	fmt.Printf("ParkingService initialized: %+v\n", parkingService)
	ownerService := services.NewOwnerService(parkingRepo, userRepo, purchaseRepo)
	fmt.Printf("OwnerService initialized: %+v\n", ownerService)
	reportService := services.NewReportService(reportRepo, parkingRepo)
	fmt.Printf("ReportService initialized: %+v\n", reportService)
	leaseService := services.NewLeaseService(leaseRepo, parkingRepo)
	fmt.Printf("LeaseService initialized: %+v\n", leaseService)
	vehicleService := services.NewVehicleService(vehicleRepo, userRepo, parkingRepo, leaseService)
	fmt.Printf("VehicleService initialized: %+v\n", vehicleService)

	// Controllers
	authController := controllers.NewAuthController(authService)
	fmt.Printf("AuthController initialized: %+v\n", authController)
	parkingController := controllers.NewParkingController(parkingService)
	fmt.Printf("ParkingController initialized: %+v\n", parkingController)
	adminController := controllers.NewAdminController(parkingService, reportService)
	fmt.Printf("AdminController initialized: %+v\n", adminController)
	leaseController := controllers.NewLeaseController(leaseService)
	fmt.Printf("LeaseController initialized: %+v\n", leaseController)
	reportController := controllers.NewReportController(reportService)
	fmt.Printf("ReportController initialized: %+v\n", reportController)
	vehicleController := controllers.NewVehicleController(vehicleService)
	fmt.Printf("VehicleController initialized: %+v\n", vehicleController)
	ownerController := controllers.NewOwnerController(ownerService)
	fmt.Printf("OwnerController initialized: %+v\n", ownerController)

	return &ControllerDependencies{
		AuthController:    authController,
		ParkingController: parkingController,
		AdminController:   adminController,
		LeaseController:   leaseController,
		ReportController:  reportController,
		VehicleController: vehicleController,
		OwnerController:   ownerController,
		Cfg:               cfg,
	}
}

// @title 停车系统 API
// @version 1.0
// @description 用于管理车位、用户、租赁的接口服务。
// @termsOfService http://swagger.io/terms/

// @contact.name 技术支持
// @contact.email support@example.com

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 使用 JWT 格式的令牌进行授权。格式：Bearer <token>

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		// 处理配置加载失败的情况
		logger.Log.Fatal("加载配置失败", zap.Error(err))
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
	// 设置日志文件路径到配置中
	cfg.LogFilePath = logPath

	// 使用配置重新初始化日志记录器
	logger.InitLogger(cfg)
	defer func() {
		if logger.Log != nil {
			logger.Log.Sync()
		}
	}()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("数据库连接失败", zap.Error(err))
	}

	database.Migrate(db)

	ctrls := initializeControllers(db, cfg)

	// 只创建一个 Gin 引擎
	router := gin.Default()

	// 挂载 CORS 中间件
	router.Use(CORSMiddleware())

	//router.Use(middleware.JWTAuthMiddleware(cfg))

	// 注册路由时传入 router
	deps := &routes.RouterDependencies{
		AuthController: ctrls.AuthController,
		ParkingService: ctrls.ParkingController,
		AdminService:   ctrls.AdminController,
		LeaseService:   ctrls.LeaseController,
		ReportService:  ctrls.ReportController,
		VehicleService: ctrls.VehicleController,
		OwnerService:   ctrls.OwnerController,
		Cfg:            ctrls.Cfg,
	}

	// 打印依赖注入信息，确认 ParkingService 正确注入
	fmt.Printf("ParkingService: %+v\n", deps.ParkingService)
	routes.SetupRouter(router, deps)
	// 打印信息，确认 SetupRouter 调用完成
	fmt.Println("SetupRouter call completed")

	// 打印所有注册的路由，用于调试
	for _, route := range router.Routes() {
		logger.Log.Info("Registered Route",
			zap.String("Method", route.Method),
			zap.String("Path", route.Path),
			zap.String("Handler", route.Handler))
	}
	for _, route := range router.Routes() {
		fmt.Printf("Method: %s, Path: %s, Handler: %s\n", route.Method, route.Path, route.Handler)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 直接从环境变量获取端口号，若未设置则使用默认值 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Log.Info("服务将启动在端口", zap.String("port", port))
	if err := router.Run(":" + port); err != nil {
		logger.Log.Fatal("服务启动失败", zap.Error(err))
	}
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	logrus.SetOutput(logFile)
	logrus.Info("This is an info log")
	logrus.Error("This is an error log")
}
