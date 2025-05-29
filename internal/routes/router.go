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
	AuthService    *controllers.AuthController
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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 公共路由
	auth := router.Group("/auth")
	{
		auth.POST("/register", deps.AuthService.Register)
		auth.POST("/login", deps.AuthService.Login)
	}

	// 需要身份验证
	authGroup := router.Group("")
	authGroup.Use(middleware.JWTAuthMiddleware("secretkey"))
	{
		authGroup.POST("/vehicles", deps.VehicleService.BindVehicle)

		authGroup.POST("/lease", deps.LeaseService.CreateLease)

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

	admin := router.Group("/admin").Use(middleware.RoleCheck(models.Admin))
	{
		admin.PUT("/spots/:id/status", deps.AdminService.UpdateSpotStatus)
		admin.GET("/stats", deps.AdminService.GetSystemStats)
		admin.POST("/spots", deps.ParkingService.CreateSpot)
	}

	router.POST("/admin/login", deps.AuthService.AdminLogin)

	report := router.Group("/reports").Use(middleware.JWTAuthMiddleware("secretkey"))
	{
		report.GET("/daily", deps.ReportService.GetDailyReport)
	}
}
