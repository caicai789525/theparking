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
	"modules/internal/services"
)

type RouterDependencies struct {
	AuthService    *services.AuthService
	AuthController *controllers.AuthController
	ParkingService *controllers.ParkingController
	AdminService   *controllers.AdminController
	LeaseService   *controllers.LeaseController
	ReportService  *controllers.ReportController
	VehicleService *controllers.VehicleController
	OwnerService   *controllers.OwnerController
	Cfg            *config.Config
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
		// 管理员登录接口
		public.POST("/admin/login", deps.AdminService.AdminLogin)
		// 其他无需认证的接口...
	}
}

// setupAuthRoutes 配置需要身份验证的路由组
func setupAuthRoutes(router *gin.Engine, deps *RouterDependencies) {
	authGroup := router.Group("/")
	// 应用 JWT 身份验证中间件
	authGroup.Use(middleware.JWTAuthMiddleware(deps.Cfg, deps.AuthService))
	{
		// 绑定车辆信息接口
		authGroup.POST("/vehicles", deps.VehicleService.BindVehicle)
		// 获取用户停车位信息接口
		authGroup.GET("/parking/my-spots", deps.ParkingService.GetUserSpots)
		// 创建租赁记录接口
		authGroup.POST("/lease", deps.LeaseService.CreateLease)

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

// setupReportRoutes 配置报表相关路由组
func setupReportRoutes(router *gin.Engine, deps *RouterDependencies) {
	report := router.Group("/reports")
	// 应用 JWT 身份验证中间件
	report.Use(middleware.JWTAuthMiddleware(deps.Cfg, deps.AuthService))
	{
		// 获取每日报表信息接口
		report.GET("/daily", deps.ReportService.GetDailyReport)
	}
}

// setupAdminRoutes 配置管理员相关路由组
func setupAdminRoutes(router *gin.Engine, deps *RouterDependencies) {
	adminGroup := router.Group("/admin")
	// 应用 JWT 身份验证和管理员角色检查中间件
	adminGroup.Use(middleware.JWTAuthMiddleware(deps.Cfg, deps.AuthService))
	adminGroup.Use(middleware.RoleCheck(models.Admin))
	{
		// 更新车位状态接口
		adminGroup.PUT("/spots/:id/status", deps.AdminService.UpdateSpotStatus)
		// 获取系统统计数据接口
		adminGroup.GET("/stats", deps.AdminService.GetSystemStats)
	}
}

// SetupRouter 配置路由
func SetupRouter(router *gin.Engine, deps *RouterDependencies) {
	setupSwaggerRoutes(router)
	setupPublicRoutes(router, deps)
	setupAuthRoutes(router, deps)
	setupReportRoutes(router, deps)
	setupAdminRoutes(router, deps) // 新增管理员路由
}
