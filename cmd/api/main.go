package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"log"
	"modules/config"
	"modules/internal/controllers"
	"modules/internal/middleware"
	"modules/internal/repositories"
	"modules/internal/routes"
	"modules/internal/services"
	"modules/pkg/database"
	"modules/pkg/logger"

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
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)

	viper.SetConfigFile("root/theparking/config/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// 读取端口配置
	port := viper.GetString("port")
	if port == "" {
		port = "8080" // 默认端口
	}

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

	router.Use(middleware.JWTAuthMiddleware(cfg))

	// 注册路由时传入 router
	routes.SetupRouter(router, &routes.RouterDependencies{
		AuthService:    ctrls.AuthController,
		ParkingService: ctrls.ParkingController,
		AdminService:   ctrls.AdminController,
		LeaseService:   ctrls.LeaseController,
		ReportService:  ctrls.ReportController,
		VehicleService: ctrls.VehicleController,
		OwnerService:   ctrls.OwnerController,
		Cfg:            ctrls.Cfg,
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Log.Fatal("服务启动失败", zap.Error(err))
	}
}
