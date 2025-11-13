package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/ride"
	"github.com/gin-gonic/gin"
)

func RideRoutes(router *gin.RouterGroup, controller ride.IRideController) {
	rideGroup := router.Group("/ride")
	{
		// CRUD operations
		rideGroup.POST("", controller.CreateRide)
		rideGroup.GET("", controller.GetAllRides)
		rideGroup.GET("/:id", controller.GetRideByID)
		rideGroup.PUT("/:id", controller.UpdateRide)
		rideGroup.DELETE("/:id", controller.DeleteRide)

		// Query by user type
		rideGroup.GET("/rider/:riderId", controller.GetRidesByRider)
		rideGroup.GET("/driver/:driverId", controller.GetRidesByDriver)
		rideGroup.GET("/pending", controller.GetPendingRides)

		// Ride lifecycle management
		rideGroup.POST("/:id/assign-driver", controller.AssignDriver)
		rideGroup.POST("/:id/match-driver", controller.MatchDriverAutomatically)
		rideGroup.POST("/:id/start", controller.StartRide)
		rideGroup.POST("/:id/complete", controller.CompleteRide)
		rideGroup.POST("/:id/cancel", controller.CancelRide)

		// Rating system
		rideGroup.POST("/:id/rate-driver", controller.RateDriver)
		rideGroup.POST("/:id/rate-rider", controller.RateRider)

		// Search
		rideGroup.POST("/search", controller.SearchRides)
	}
}
