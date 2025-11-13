package location

import (
	"net/http"
	"strconv"

	"github.com/gbrayhan/microservices-go/src/application/usecases/location"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainLocation "github.com/gbrayhan/microservices-go/src/domain/location"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ILocationController interface {
	CreateFavorite(ctx *gin.Context)
	GetUserFavorites(ctx *gin.Context)
	GetFavoriteByID(ctx *gin.Context)
	UpdateFavorite(ctx *gin.Context)
	DeleteFavorite(ctx *gin.Context)
	SetAsPrimary(ctx *gin.Context)
}

type LocationController struct {
	locationUseCase location.ILocationUseCase
	logger          *logger.Logger
}

func NewLocationController(locationUseCase location.ILocationUseCase, logger *logger.Logger) *LocationController {
	return &LocationController{
		locationUseCase: locationUseCase,
		logger:          logger,
	}
}

type CreateFavoriteRequest struct {
	Label     string  `json:"label" binding:"required"`
	Type      string  `json:"type" binding:"required"`
	Address   string  `json:"address" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	IsPrimary bool    `json:"isPrimary"`
}

func (c *LocationController) CreateFavorite(ctx *gin.Context) {
	var req CreateFavoriteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotAuthenticated))
		return
	}

	c.logger.Info("Creating favorite location", 
		zap.Int("userID", userID.(int)),
		zap.String("label", req.Label))

	favoriteLocation := &domainLocation.FavoriteLocation{
		UserID:    userID.(int),
		Label:     req.Label,
		Type:      domainLocation.LocationType(req.Type),
		Address:   req.Address,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		IsPrimary: req.IsPrimary,
	}

	created, err := c.locationUseCase.CreateFavorite(favoriteLocation)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Favorite location created successfully",
		"data":    created,
	})
}

func (c *LocationController) GetUserFavorites(ctx *gin.Context) {
	userIDParam := ctx.Param("userId")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting user favorites", zap.Int("userID", userID))

	favorites, err := c.locationUseCase.GetUserFavorites(userID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    favorites,
	})
}

func (c *LocationController) GetFavoriteByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting favorite by ID", zap.Int("id", id))

	favorite, err := c.locationUseCase.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    favorite,
	})
}

func (c *LocationController) UpdateFavorite(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Updating favorite location", zap.Int("id", id))

	updated, err := c.locationUseCase.UpdateFavorite(id, updates)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Favorite location updated successfully",
		"data":    updated,
	})
}

func (c *LocationController) DeleteFavorite(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Deleting favorite location", zap.Int("id", id))

	err = c.locationUseCase.DeleteFavorite(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Favorite location deleted successfully",
	})
}

func (c *LocationController) SetAsPrimary(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotAuthenticated))
		return
	}

	c.logger.Info("Setting favorite as primary", zap.Int("id", id))

	err = c.locationUseCase.SetAsPrimary(id, userID.(int))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Favorite location set as primary successfully",
	})
}
