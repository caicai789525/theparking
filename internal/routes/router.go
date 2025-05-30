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

func SetupRouter(router *gin.Engine, deps *RouterDependencies) {
	// 使用传进来的 router，而不是重新创建 r := gin.Default()

	// Swagger docs 路由
	setupSwaggerRoutes(router)

	// 公共路由
	setupPublicRoutes(router, deps)

	// 需要身份验证的路由
	setupAuthRoutes(router, deps)

	// 管理员路由
	setupAdminRoutes(router, deps)

	// 报表路由
	setupReportRoutes(router, deps)
}

func setupSwaggerRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func setupPublicRoutes(router *gin.Engine, deps *RouterDependencies) {
	public := router.Group("/")
	{
		public.POST("/auth/register", deps.AuthController.Register)
		public.POST("/auth/login", deps.AuthController.Login)
		// 其他无需认证的接口...
	}
}

func setupAuthRoutes(router *gin.Engine, deps *RouterDependencies) {
	authGroup := router.Group("")
	authGroup.Use(middleware.JWTAuthMiddleware(deps.Cfg))
	{
		authGroup.POST("/vehicles", deps.VehicleService.BindVehicle)
		authGroup.POST("/lease", deps.LeaseService.CreateLease)
		authGroup.GET("/parking/my-spots", deps.ParkingService.GetUserSpots)

		parking := authGroup.Group("/parking")
		{
			parking.GET("/spots", deps.ParkingService.ListSpots)
			parking.POST("/entry", deps.ParkingService.Entry)
			parking.POST("/exit/:id", deps.ParkingService.Exit)
			parking.POST("/rent", deps.VehicleService.PublishForRent)
		}

		owner := authGroup.Group("/owner").Use(middleware.RoleCheck(models.Owner))
		{
			owner.POST("/purchase", deps.OwnerService.PurchaseSpot)
			owner.POST("/spots", deps.ParkingService.CreateSpot)
		}
	}
}

func setupAdminRoutes(router *gin.Engine, deps *RouterDependencies) {
	admin := router.Group("/admin").Use(middleware.RoleCheck(models.Admin))
	{
		admin.PUT("/spots/:id/status", deps.AdminService.UpdateSpotStatus)
		admin.GET("/stats", deps.AdminService.GetSystemStats)
		admin.POST("/spots", deps.ParkingService.CreateSpot)
	}
	router.POST("/admin/login", deps.AuthController.AdminLogin)
}

func setupReportRoutes(router *gin.Engine, deps *RouterDependencies) {
	report := router.Group("/reports")
	report.Use(middleware.JWTAuthMiddleware(deps.Cfg))
	{
		report.GET("/daily", deps.ReportService.GetDailyReport)
	}
}
