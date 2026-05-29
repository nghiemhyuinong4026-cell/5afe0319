package routes

import (
	"vehicle-management-system/config"
	"vehicle-management-system/controllers"
	"vehicle-management-system/middleware"
	"vehicle-management-system/models"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// 初始化控制器
	authController := controllers.NewAuthController(&cfg.JWT)
	vehicleController := controllers.NewVehicleController()
	driverController := controllers.NewDriverController()
	requestController := controllers.NewRequestController()
	approvalController := controllers.NewApprovalController()
	dispatchController := controllers.NewDispatchController()
	tripController := controllers.NewTripController()
	auditController := controllers.NewAuditController()

	// 公开路由
	router.POST("/api/login", authController.Login)

	// 需要认证的路由
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		// 车辆列表
		api.GET("/vehicles", vehicleController.List)

		// 司机列表
		api.GET("/drivers", driverController.List)

		// 用车申请相关
		requests := api.Group("/requests")
		{
			requests.POST("", requestController.Create)
			requests.GET("/my", requestController.MyRequests)
			requests.GET("/:id", requestController.GetByID)
		}

		// 审批相关 - 仅manager角色
		approvals := api.Group("/approvals")
		approvals.Use(middleware.RoleMiddleware(string(models.RoleManager)))
		{
			approvals.GET("/pending", approvalController.ListPendingApproval)
			approvals.POST("/:id/approve", approvalController.Approve)
			approvals.POST("/:id/reject", approvalController.Reject)
		}

		// 派车相关 - 仅dispatcher角色
		dispatches := api.Group("/dispatches")
		dispatches.Use(middleware.RoleMiddleware(string(models.RoleDispatcher)))
		{
			dispatches.GET("/pending", dispatchController.ListPendingDispatch)
			dispatches.POST("", dispatchController.CreateDispatch)
		}

		// 行程相关 - 仅driver角色
		trips := api.Group("/trips")
		trips.Use(middleware.RoleMiddleware(string(models.RoleDriver)))
		{
			trips.GET("/my", tripController.MyTrips)
			trips.POST("/:id/start", tripController.StartTrip)
			trips.POST("/:id/complete", tripController.CompleteTrip)
		}

		// 审计日志 - 所有已认证用户都可以查看
		api.GET("/audit-logs", auditController.List)
	}

	return router
}
