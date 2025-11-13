package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/document"
	"github.com/gin-gonic/gin"
)

func DocumentRoutes(router *gin.RouterGroup, controller document.IDocumentController) {
	documentGroup := router.Group("/document")
	{
		// Driver endpoints
		documentGroup.POST("", controller.UploadDocument)
		documentGroup.GET("/driver/:driverId", controller.GetDriverDocuments)
		documentGroup.GET("/:id", controller.GetDocumentByID)
		documentGroup.DELETE("/:id", controller.DeleteDocument)
		documentGroup.GET("/driver/:driverId/summary", controller.GetDocumentSummary)

		// Admin endpoints (should be protected with auth + admin middleware)
		documentGroup.GET("/pending", controller.GetPendingDocuments)
		documentGroup.POST("/:id/approve", controller.ApproveDocument)
		documentGroup.POST("/:id/reject", controller.RejectDocument)
	}
}
