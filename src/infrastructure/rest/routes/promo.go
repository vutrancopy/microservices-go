package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/promo"
	"github.com/gin-gonic/gin"
)

func PromoRoutes(router *gin.RouterGroup, controller promo.IPromoController) {
	promoGroup := router.Group("/promo")
	{
		// Public/User endpoints
		promoGroup.GET("/active", controller.GetAllActivePromos)
		promoGroup.POST("/validate", controller.ValidatePromoCode)
		promoGroup.POST("/apply", controller.ApplyPromoCode)
		promoGroup.GET("/history/:userId", controller.GetUserPromoHistory)

		// Admin endpoints (should be protected with auth + admin middleware)
		promoGroup.POST("", controller.CreatePromoCode)
		promoGroup.PUT("/:id", controller.UpdatePromoCode)
		promoGroup.DELETE("/:id", controller.DeactivatePromoCode)
	}
}
