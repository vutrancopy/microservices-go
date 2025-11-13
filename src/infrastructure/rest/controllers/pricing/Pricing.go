package pricing

import (
	"net/http"
	"strconv"

	"github.com/gbrayhan/microservices-go/src/application/usecases/pricing"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainPricing "github.com/gbrayhan/microservices-go/src/domain/pricing"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IPricingController interface {
	GetPricingByVehicleType(ctx *gin.Context)
	CalculateFareEstimate(ctx *gin.Context)
	GetCancellationFee(ctx *gin.Context)
	GetAllPricings(ctx *gin.Context)
	UpdatePricing(ctx *gin.Context)
}

type PricingController struct {
	pricingUseCase pricing.IPricingUseCase
	logger         *logger.Logger
}

func NewPricingController(pricingUseCase pricing.IPricingUseCase, logger *logger.Logger) *PricingController {
	return &PricingController{
		pricingUseCase: pricingUseCase,
		logger:         logger,
	}
}

// FareEstimateRequest represents fare estimation request
type FareEstimateRequest struct {
	VehicleType      string  `json:"vehicleType" binding:"required"`
	DistanceKm       float64 `json:"distanceKm" binding:"required,gt=0"`
	EstimatedMinutes int     `json:"estimatedMinutes" binding:"required,gt=0"`
	WaitingMinutes   int     `json:"waitingMinutes"`
	SurgeMultiplier  float64 `json:"surgeMultiplier"`
	PromoCode        string  `json:"promoCode"`
}

// CancellationFeeRequest represents cancellation fee request
type CancellationFeeRequest struct {
	VehicleType          string `json:"vehicleType" binding:"required"`
	MinutesSinceBooking  int    `json:"minutesSinceBooking" binding:"required,gte=0"`
}

func (c *PricingController) GetPricingByVehicleType(ctx *gin.Context) {
	vehicleType := ctx.Param("vehicleType")

	c.logger.Info("Getting pricing for vehicle type", zap.String("vehicleType", vehicleType))

	// Validate vehicle type
	vType := common.VehicleType(vehicleType)
	if vType != common.VehicleTypeSedan && vType != common.VehicleTypeSUV && 
	   vType != common.VehicleTypeVan && vType != common.VehicleTypeBike {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	pricingConfig, err := c.pricingUseCase.GetPricingByVehicleType(vType)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pricingConfig,
	})
}

func (c *PricingController) CalculateFareEstimate(ctx *gin.Context) {
	var req FareEstimateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Calculating fare estimate", 
		zap.String("vehicleType", req.VehicleType),
		zap.Float64("distance", req.DistanceKm))

	// Validate vehicle type
	vType := common.VehicleType(req.VehicleType)
	if vType != common.VehicleTypeSedan && vType != common.VehicleTypeSUV && 
	   vType != common.VehicleTypeVan && vType != common.VehicleTypeBike {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	// Create estimate request
	estimateReq := &domainPricing.FareEstimateRequest{
		VehicleType:      vType,
		DistanceKm:       req.DistanceKm,
		EstimatedMinutes: req.EstimatedMinutes,
		WaitingMinutes:   req.WaitingMinutes,
		SurgeMultiplier:  req.SurgeMultiplier,
		PromoCode:        req.PromoCode,
	}

	fareBreakdown, err := c.pricingUseCase.CalculateFareEstimate(estimateReq)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fareBreakdown,
	})
}

func (c *PricingController) GetCancellationFee(ctx *gin.Context) {
	var req CancellationFeeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	vType := common.VehicleType(req.VehicleType)
	fee, err := c.pricingUseCase.GetCancellationFee(vType, req.MinutesSinceBooking)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"cancellationFee": fee,
			"vehicleType":     req.VehicleType,
			"minutesSinceBooking": req.MinutesSinceBooking,
		},
	})
}

func (c *PricingController) GetAllPricings(ctx *gin.Context) {
	c.logger.Info("Getting all pricing configs")

	configs, err := c.pricingUseCase.GetActivePricingConfigs()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

func (c *PricingController) UpdatePricing(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var pricingConfig domainPricing.PricingConfig
	if err := ctx.ShouldBindJSON(&pricingConfig); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	pricingConfig.ID = id

	updatedConfig, err := c.pricingUseCase.UpdatePricingConfig(&pricingConfig)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pricing config updated successfully",
		"data":    updatedConfig,
	})
}
