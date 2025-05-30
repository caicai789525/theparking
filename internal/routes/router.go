package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"modules/config"
	_ "modules/docs" // 注意：必须导入以启用 Swagger 文档生成

	"modules/internal/controllers"
	"modules/internal/middleware"
	"modules/internal/models"
)

// RouterDependencies 定义路由所需的依赖项
type RouterDependencies struct {
	AuthController *controllers.AuthController
	ParkingService *controllers.ParkingController
	AdminService   *controllers.AdminController
	LeaseService   *controllers.LeaseController
	ReportService  *controllers.ReportController
	VehicleService *controllers.VehicleController
	OwnerService   *controllers.OwnerController
	Cfg            *config.Config
}

// SetupRouter 初始化整个路由系统
func SetupRouter(router *gin.Engine, deps *RouterDependencies) {
	// 使用传进来的 router，而不是重新创建 r := gin.Default()

	// 配置 Swagger 文档路由
	setupSwaggerRoutes(router)

	// 配置公共路由，无需认证即可访问
	setupPublicRoutes(router, deps)

	// 配置需要身份验证的路由
	setupAuthRoutes(router, deps)

	// 配置管理员专属路由
	setupAdminRoutes(router, deps)

	// 配置报表相关路由
	setupReportRoutes(router, deps)
}

// setupSwaggerRoutes 配置 Swagger 文档的访问路由
func setupSwaggerRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// setupPublicRoutes 配置公共路由组，包含注册、用户登录等接口
func setupPublicRoutes(router *gin.Engine, deps *RouterDependencies) {
	public := router.Group("/")
	{
		// 用户注册接口
		public.POST("/auth/register", deps.AuthController.Register)
		// 用户登录接口
		public.POST("/auth/login", deps.AuthController.UserLogin)
		// 其他无需认证的接口...
	}
}

// setupAuthRoutes 配置需要身份验证的路由组
func setupAuthRoutes(router *gin.Engine, deps *RouterDependencies) {
	authGroup := router.Group("")
	// 应用 JWT 身份验证中间件
	authGroup.Use(middleware.JWTAuthMiddleware(deps.Cfg))
	{
		// 绑定车辆信息接口
		authGroup.POST("/vehicles", deps.VehicleService.BindVehicle)
		// 创建租赁记录接口
		authGroup.POST("/lease", deps.LeaseService.CreateLease)
		// 获取用户停车位信息接口
		authGroup.GET("/parking/my-spots", deps.ParkingService.GetUserSpots)

		parking := authGroup.Group("/parking")
		{
			// 列出所有停车位信息接口
			parking.GET("/spots", deps.ParkingService.ListSpots)
			// 车辆进入停车场接口
			parking.POST("/entry", deps.ParkingService.Entry)
			// 车辆离开停车场接口
			parking.POST("/exit/:id", deps.ParkingService.Exit)
			// 发布车辆出租信息接口
			parking.POST("/rent", deps.VehicleService.PublishForRent)
		}

		owner := authGroup.Group("/owner").Use(middleware.RoleCheck(models.Owner))
		{
			// 业主购买停车位接口
			owner.POST("/purchase", deps.OwnerService.PurchaseSpot)
			// 业主创建停车位接口
			owner.POST("/spots", deps.ParkingService.CreateSpot)
		}
	}
}

// setupAdminRoutes 配置管理员专属路由组
func setupAdminRoutes(router *gin.Engine, deps *RouterDependencies) {
	admin := router.Group("/admin").Use(middleware.RoleCheck(models.Admin))
	{
		// 更新停车位状态接口
		admin.PUT("/spots/:id/status", deps.AdminService.UpdateSpotStatus)
		// 获取系统统计信息接口
		admin.GET("/stats", deps.AdminService.GetSystemStats)
		// 管理员创建停车位接口
		admin.POST("/spots", deps.ParkingService.CreateSpot)
	}
	// 管理员登录接口
	router.POST("/admin/login", deps.AuthController.AdminLogin)
}

// setupReportRoutes 配置报表相关路由组
func setupReportRoutes(router *gin.Engine, deps *RouterDependencies) {
	report := router.Group("/reports")
	// 应用 JWT 身份验证中间件
	report.Use(middleware.JWTAuthMiddleware(deps.Cfg))
	{
		// 获取每日报表信息接口
		report.GET("/daily", deps.ReportService.GetDailyReport)
	}
}
