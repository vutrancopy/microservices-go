package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/pricing"
	"github.com/gin-gonic/gin"
)

func PricingRoutes(router *gin.RouterGroup, controller pricing.IPricingController) {
	pricingGroup := router.Group("/pricing")
	{
		// Public endpoints
		pricingGroup.GET("", controller.GetAllPricings)
		pricingGroup.GET("/vehicle/:vehicleType", controller.GetPricingByVehicleType)
		pricingGroup.POST("/estimate", controller.CalculateFareEstimate)
		pricingGroup.POST("/cancellation-fee", controller.GetCancellationFee)

		// Admin endpoints (should be protected with auth + admin middleware)
		pricingGroup.PUT("/:id", controller.UpdatePricing)
	}
}
