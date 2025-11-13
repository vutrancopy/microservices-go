package routes

import (
	"net/http"

	"github.com/gbrayhan/microservices-go/src/infrastructure/di"
	"github.com/gin-gonic/gin"
)

func ApplicationRouter(router *gin.Engine, appContext *di.ApplicationContext) {
	v1 := router.Group("/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Ride Hailing API is running",
			"version": "1.0.0",
		})
	})

	// Authentication & User Management
	AuthRoutes(v1, appContext.AuthController)
	UserRoutes(v1, appContext.UserController)
	
	// Ride Hailing Core Services
	VehicleRoutes(v1, appContext.VehicleController)
	RideRoutes(v1, appContext.RideController)
	PaymentRoutes(v1, appContext.PaymentController)
	
	// Enhanced Features
	PricingRoutes(v1, appContext.PricingController)
	PromoRoutes(v1, appContext.PromoController)
	LocationRoutes(v1, appContext.LocationController)
	DocumentRoutes(v1, appContext.DocumentController)
	ScheduledRoutes(v1, appContext.ScheduledController)
	WalletRoutes(v1, appContext.WalletController)
	
	// Legacy (will be removed in future versions)
	MedicineRoutes(v1, appContext.MedicineController)
}
