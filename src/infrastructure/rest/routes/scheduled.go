package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/scheduled"
	"github.com/gin-gonic/gin"
)

func ScheduledRoutes(router *gin.RouterGroup, controller scheduled.IScheduledController) {
	scheduledGroup := router.Group("/scheduled")
	{
		// User endpoints
		scheduledGroup.POST("", controller.CreateScheduledRide)
		scheduledGroup.GET("/:id", controller.GetScheduledRideByID)
		scheduledGroup.GET("/user/:userId", controller.GetUserScheduledRides)
		scheduledGroup.GET("/upcoming", controller.GetUpcomingScheduledRides)
		scheduledGroup.POST("/:id/cancel", controller.CancelScheduledRide)
		scheduledGroup.PUT("/:id", controller.UpdateScheduledRide)
	}
}
