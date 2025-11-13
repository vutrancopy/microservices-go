package scheduled

import (
	"net/http"
	"strconv"

	"github.com/gbrayhan/microservices-go/src/application/usecases/scheduled"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainScheduled "github.com/gbrayhan/microservices-go/src/domain/scheduled"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IScheduledController interface {
	CreateScheduledRide(ctx *gin.Context)
	GetScheduledRideByID(ctx *gin.Context)
	GetUserScheduledRides(ctx *gin.Context)
	GetUpcomingScheduledRides(ctx *gin.Context)
	CancelScheduledRide(ctx *gin.Context)
	UpdateScheduledRide(ctx *gin.Context)
}

type ScheduledController struct {
	scheduledUseCase scheduled.IScheduledRideUseCase
	logger           *logger.Logger
}

func NewScheduledController(scheduledUseCase scheduled.IScheduledRideUseCase, logger *logger.Logger) *ScheduledController {
	return &ScheduledController{
		scheduledUseCase: scheduledUseCase,
		logger:           logger,
	}
}

type CreateScheduledRideRequest struct {
	VehicleType       string  `json:"vehicleType" binding:"required"`
	PickupLatitude    float64 `json:"pickupLatitude" binding:"required"`
	PickupLongitude   float64 `json:"pickupLongitude" binding:"required"`
	PickupAddress     string  `json:"pickupAddress" binding:"required"`
	DropoffLatitude   float64 `json:"dropoffLatitude" binding:"required"`
	DropoffLongitude  float64 `json:"dropoffLongitude" binding:"required"`
	DropoffAddress    string  `json:"dropoffAddress" binding:"required"`
	ScheduledTime     string  `json:"scheduledTime" binding:"required"`
	EstimatedDistance float64 `json:"estimatedDistance"`
	EstimatedDuration int     `json:"estimatedDuration"`
	EstimatedFare     float64 `json:"estimatedFare"`
	Notes             string  `json:"notes"`
	PromoCode         string  `json:"promoCode"`
}

type CancelScheduledRideRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (c *ScheduledController) CreateScheduledRide(ctx *gin.Context) {
	var req CreateScheduledRideRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotAuthenticated))
		return
	}

	c.logger.Info("Creating scheduled ride", 
		zap.Int("riderID", userID.(int)),
		zap.String("scheduledTime", req.ScheduledTime))

	scheduledRide := &domainScheduled.ScheduledRide{
		RiderID:           userID.(int),
		PickupLatitude:    req.PickupLatitude,
		PickupLongitude:   req.PickupLongitude,
		PickupAddress:     req.PickupAddress,
		DropoffLatitude:   req.DropoffLatitude,
		DropoffLongitude:  req.DropoffLongitude,
		DropoffAddress:    req.DropoffAddress,
		EstimatedDistance: req.EstimatedDistance,
		EstimatedDuration: req.EstimatedDuration,
		EstimatedFare:     req.EstimatedFare,
		Notes:             req.Notes,
		PromoCode:         req.PromoCode,
	}

	created, err := c.scheduledUseCase.CreateScheduledRide(scheduledRide)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Scheduled ride created successfully",
		"data":    created,
	})
}

func (c *ScheduledController) GetScheduledRideByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting scheduled ride by ID", zap.Int("id", id))

	ride, err := c.scheduledUseCase.GetByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ride,
	})
}

func (c *ScheduledController) GetUserScheduledRides(ctx *gin.Context) {
	userIDParam := ctx.Param("userId")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting user scheduled rides", zap.Int("userID", userID))

	rides, err := c.scheduledUseCase.GetUserScheduledRides(userID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rides,
	})
}

func (c *ScheduledController) GetUpcomingScheduledRides(ctx *gin.Context) {
	limitParam := ctx.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitParam)

	c.logger.Info("Getting upcoming scheduled rides", zap.Int("limit", limit))

	rides, err := c.scheduledUseCase.GetUpcomingScheduledRides(limit)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rides,
	})
}

func (c *ScheduledController) CancelScheduledRide(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var req CancelScheduledRideRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotAuthenticated))
		return
	}

	c.logger.Info("Cancelling scheduled ride", zap.Int("id", id))

	err = c.scheduledUseCase.CancelScheduledRide(id, req.Reason, userID.(int))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Scheduled ride cancelled successfully",
	})
}

func (c *ScheduledController) UpdateScheduledRide(ctx *gin.Context) {
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

	c.logger.Info("Updating scheduled ride", zap.Int("id", id))

	updated, err := c.scheduledUseCase.UpdateScheduledRide(id, updates)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Scheduled ride updated successfully",
		"data":    updated,
	})
}
