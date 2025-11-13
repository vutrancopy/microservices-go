package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/vehicle"
	"github.com/gin-gonic/gin"
)

func VehicleRoutes(router *gin.RouterGroup, controller vehicle.IVehicleController) {
	vehicleGroup := router.Group("/vehicle")
	{
		// CRUD operations
		vehicleGroup.POST("", controller.CreateVehicle)
		vehicleGroup.GET("", controller.GetAllVehicles)
		vehicleGroup.GET("/:id", controller.GetVehicleByID)
		vehicleGroup.PUT("/:id", controller.UpdateVehicle)
		vehicleGroup.DELETE("/:id", controller.DeleteVehicle)

		// Driver-specific
		vehicleGroup.GET("/driver/:driverId", controller.GetVehiclesByDriver)

		// Status management
		vehicleGroup.PUT("/:id/activate", controller.ActivateVehicle)
		vehicleGroup.PUT("/:id/deactivate", controller.DeactivateVehicle)

		// Search
		vehicleGroup.POST("/search", controller.SearchVehicles)
	}
}
