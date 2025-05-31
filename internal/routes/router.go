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

// applyAuthMiddleware 应用 JWT 身份验证中间件
func applyAuthMiddleware(group *gin.RouterGroup, deps *RouterDependencies) {
	group.Use(middleware.JWTAuthMiddleware(deps.Cfg, deps.AuthService))
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

// setupVehicleRoutes 配置车辆相关路由组
func setupVehicleRoutes(authGroup *gin.RouterGroup, deps *RouterDependencies) {
	vehicles := authGroup.Group("/vehicles")
	{
		// 绑定车辆信息接口
		vehicles.POST("", deps.VehicleService.BindVehicle)
		// 查询自己的车辆接口
		vehicles.GET("", deps.VehicleService.GetUserVehicles)
		// 删除车辆接口
		vehicles.DELETE("/:id", deps.VehicleService.RemoveVehicle)
	}
}

// setupParkingRoutes 配置停车场相关路由组
func setupParkingRoutes(authGroup *gin.RouterGroup, deps *RouterDependencies) {
	parking := authGroup.Group("/parking")
	{
		// 列出所有停车位信息接口
		parking.GET("/spots", deps.ParkingService.ListSpots)
		// 获取用户停车位信息接口
		parking.GET("/my-spots", deps.ParkingService.GetUserSpots)
		// 车辆进入停车场接口
		parking.POST("/entry", deps.ParkingService.Entry)
		// 车辆离开停车场接口
		parking.POST("/exit/:id", deps.ParkingService.Exit)
		// 发布车辆出租信息接口
		parking.POST("/rent", deps.VehicleService.PublishForRent)
	}
}

// setupLeaseRoutes 配置租赁相关路由组
func setupLeaseRoutes(authGroup *gin.RouterGroup, deps *RouterDependencies) {
	// 创建租赁记录接口
	authGroup.POST("/lease", deps.LeaseService.CreateLease)
}

// setupOwnerRoutes 配置业主相关路由组
func setupOwnerRoutes(authGroup *gin.RouterGroup, deps *RouterDependencies) {
	owner := authGroup.Group("/owner").Use(middleware.RoleCheck(models.Owner))
	{
		// 业主购买停车位接口
		owner.POST("/purchase", deps.OwnerService.PurchaseSpot)
		// 业主创建停车位接口
		owner.POST("/spots", deps.ParkingService.CreateSpot)
	}
}

// setupAuthRoutes 配置需要身份验证的路由组
func setupAuthRoutes(router *gin.Engine, deps *RouterDependencies) {
	authGroup := router.Group("/")
	applyAuthMiddleware(authGroup, deps)

	setupVehicleRoutes(authGroup, deps)
	setupParkingRoutes(authGroup, deps)
	setupLeaseRoutes(authGroup, deps)
	setupOwnerRoutes(authGroup, deps)
}

// setupReportRoutes 配置报表相关路由组
func setupReportRoutes(router *gin.Engine, deps *RouterDependencies) {
	report := router.Group("/reports")
	applyAuthMiddleware(report, deps)
	{
		// 获取每日报表信息接口
		report.GET("/daily", deps.ReportService.GetDailyReport)
	}
}

// setupAdminRoutes 配置管理员相关路由组
func setupAdminRoutes(router *gin.Engine, deps *RouterDependencies) {
	adminGroup := router.Group("/admin")
	applyAuthMiddleware(adminGroup, deps)
	adminGroup.Use(middleware.RoleCheck(models.Admin))
	{
		// 更新车位状态接口
		adminGroup.PUT("/spots/:id/status", deps.AdminService.UpdateSpotStatus)
		// 获取系统统计数据接口
		adminGroup.GET("/stats", deps.AdminService.GetSystemStats)
		adminGroup.POST("/bind-parking", deps.AdminService.BindParkingToUser)
		// 解除车位与用户绑定接口
		adminGroup.DELETE("/unbind-parking/:parkingID", deps.AdminService.UnbindParkingFromUser)
		adminGroup.GET("/users/:userID", deps.AdminService.GetUserInfo)
		// 查询车位绑定用户信息接口
		adminGroup.GET("parking/:parkingID/bind-user", deps.AdminService.GetParkingBindUser)
	}
}

// SetupRouter 配置路由
func SetupRouter(router *gin.Engine, deps *RouterDependencies) {
	setupSwaggerRoutes(router)
	setupPublicRoutes(router, deps)
	setupAuthRoutes(router, deps)
	setupReportRoutes(router, deps)
	setupAdminRoutes(router, deps)
}
