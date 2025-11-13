package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/payment"
	"github.com/gin-gonic/gin"
)

func PaymentRoutes(router *gin.RouterGroup, controller payment.IPaymentController) {
	paymentGroup := router.Group("/payment")
	{
		// CRUD operations
		paymentGroup.POST("", controller.CreatePayment)
		paymentGroup.GET("", controller.GetAllPayments)
		paymentGroup.GET("/:id", controller.GetPaymentByID)
		paymentGroup.PUT("/:id", controller.UpdatePayment)
		paymentGroup.DELETE("/:id", controller.DeletePayment)

		// Query by relationship
		paymentGroup.GET("/ride/:rideId", controller.GetPaymentByRide)
		paymentGroup.GET("/rider/:riderId", controller.GetPaymentsByRider)
		paymentGroup.GET("/driver/:driverId", controller.GetPaymentsByDriver)
		paymentGroup.GET("/pending", controller.GetPendingPayments)

		// Payment operations
		paymentGroup.POST("/:id/process", controller.ProcessPayment)
		paymentGroup.POST("/:id/fail", controller.FailPayment)
		paymentGroup.POST("/:id/refund", controller.RefundPayment)

		// Driver earnings
		paymentGroup.GET("/driver/:driverId/earnings", controller.GetDriverEarnings)

		// Search
		paymentGroup.POST("/search", controller.SearchPayments)
	}
}
