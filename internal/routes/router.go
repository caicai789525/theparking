package routes

import (
	"fmt"
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

// SetupRouter 配置路由
func SetupRouter(router *gin.Engine, deps *RouterDependencies) {
	setupSwaggerRoutes(router)
	setupPublicRoutes(router, deps)
	setupAuthRoutes(router, deps)
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
		authGroup.POST("/vehicles", func(ctx *gin.Context) {
			fmt.Println("Accessing /vehicles route")
			deps.VehicleService.BindVehicle(ctx)
		})
		// 创建租赁记录接口
		authGroup.POST("/lease", func(ctx *gin.Context) {
			fmt.Println("Accessing /lease route")
			deps.LeaseService.CreateLease(ctx)
		})
		// 获取用户停车位信息接口，添加日志确认执行
		authGroup.GET("/parking/myspots", func(ctx *gin.Context) {
			// 临时添加日志，确认路由被访问
			fmt.Println("Accessing /parking/my-spots route")
			deps.ParkingService.GetUserSpots(ctx)
		})

		parking := authGroup.Group("/parking")
		{
			// 列出所有停车位信息接口
			parking.GET("/spots", func(ctx *gin.Context) {
				fmt.Println("Accessing /parking/spots route")
				deps.ParkingService.ListSpots(ctx)
			})
			// 车辆进入停车场接口
			parking.POST("/entry", func(ctx *gin.Context) {
				fmt.Println("Accessing /parking/entry route")
				deps.ParkingService.Entry(ctx)
			})
			// 车辆离开停车场接口
			parking.POST("/exit/:id", func(ctx *gin.Context) {
				fmt.Println("Accessing /parking/exit/:id route")
				deps.ParkingService.Exit(ctx)
			})
			// 发布车辆出租信息接口
			parking.POST("/rent", func(ctx *gin.Context) {
				fmt.Println("Accessing /parking/rent route")
				deps.VehicleService.PublishForRent(ctx)
			})
		}

		owner := authGroup.Group("/owner").Use(middleware.RoleCheck(models.Owner))
		{
			// 业主购买停车位接口
			owner.POST("/purchase", func(ctx *gin.Context) {
				fmt.Println("Accessing /owner/purchase route")
				deps.OwnerService.PurchaseSpot(ctx)
			})
			// 业主创建停车位接口
			owner.POST("/spots", func(ctx *gin.Context) {
				fmt.Println("Accessing /owner/spots route")
				deps.ParkingService.CreateSpot(ctx)
			})
		}
	}
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
