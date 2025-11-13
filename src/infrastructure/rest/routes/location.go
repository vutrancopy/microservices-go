package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/location"
	"github.com/gin-gonic/gin"
)

func LocationRoutes(router *gin.RouterGroup, controller location.ILocationController) {
	locationGroup := router.Group("/location")
	{
		// Favorite locations endpoints
		locationGroup.POST("/favorite", controller.CreateFavorite)
		locationGroup.GET("/favorite/user/:userId", controller.GetUserFavorites)
		locationGroup.GET("/favorite/:id", controller.GetFavoriteByID)
		locationGroup.PUT("/favorite/:id", controller.UpdateFavorite)
		locationGroup.DELETE("/favorite/:id", controller.DeleteFavorite)
		locationGroup.PUT("/favorite/:id/primary", controller.SetAsPrimary)
	}
}
