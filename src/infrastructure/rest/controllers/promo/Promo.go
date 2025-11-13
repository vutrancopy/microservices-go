package promo

import (
	"net/http"
	"strconv"

	"github.com/gbrayhan/microservices-go/src/application/usecases/promo"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainPromo "github.com/gbrayhan/microservices-go/src/domain/promo"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IPromoController interface {
	ValidatePromoCode(ctx *gin.Context)
	ApplyPromoCode(ctx *gin.Context)
	GetAllActivePromos(ctx *gin.Context)
	GetUserPromoHistory(ctx *gin.Context)
	CreatePromoCode(ctx *gin.Context)
	UpdatePromoCode(ctx *gin.Context)
	DeactivatePromoCode(ctx *gin.Context)
}

type PromoController struct {
	promoUseCase promo.IPromoUseCase
	logger       *logger.Logger
}

func NewPromoController(promoUseCase promo.IPromoUseCase, logger *logger.Logger) *PromoController {
	return &PromoController{
		promoUseCase: promoUseCase,
		logger:       logger,
	}
}

type ValidatePromoRequest struct {
	Code      string  `json:"code" binding:"required"`
	UserID    int     `json:"userId" binding:"required"`
	RideFare  float64 `json:"rideFare" binding:"required,gt=0"`
}

type ApplyPromoRequest struct {
	PromoCodeID    int     `json:"promoCodeId" binding:"required"`
	UserID         int     `json:"userId" binding:"required"`
	RideID         int     `json:"rideId" binding:"required"`
	DiscountAmount float64 `json:"discountAmount" binding:"required,gte=0"`
}

func (c *PromoController) ValidatePromoCode(ctx *gin.Context) {
	var req ValidatePromoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Validating promo code", 
		zap.String("code", req.Code),
		zap.Int("userID", req.UserID))

	promoCode, discount, err := c.promoUseCase.ValidatePromoCode(req.Code, req.UserID, req.RideFare)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"promoCode":      promoCode,
			"discount":       discount,
			"originalFare":   req.RideFare,
			"finalFare":      req.RideFare - discount,
		},
	})
}

func (c *PromoController) ApplyPromoCode(ctx *gin.Context) {
	var req ApplyPromoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Applying promo code", 
		zap.Int("promoCodeID", req.PromoCodeID),
		zap.Int("rideID", req.RideID))

	err := c.promoUseCase.ApplyPromoCode(req.PromoCodeID, req.UserID, req.RideID, req.DiscountAmount)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Promo code applied successfully",
	})
}

func (c *PromoController) GetAllActivePromos(ctx *gin.Context) {
	c.logger.Info("Getting all active promo codes")

	promos, err := c.promoUseCase.GetAllActivePromos()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    promos,
	})
}

func (c *PromoController) GetUserPromoHistory(ctx *gin.Context) {
	userIDParam := ctx.Param("userId")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting user promo history", zap.Int("userID", userID))

	history, err := c.promoUseCase.GetUserPromoUsageHistory(userID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

func (c *PromoController) CreatePromoCode(ctx *gin.Context) {
	var promoCode domainPromo.PromoCode
	if err := ctx.ShouldBindJSON(&promoCode); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Creating promo code", zap.String("code", promoCode.Code))

	createdPromo, err := c.promoUseCase.CreatePromoCode(&promoCode)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Promo code created successfully",
		"data":    createdPromo,
	})
}

func (c *PromoController) UpdatePromoCode(ctx *gin.Context) {
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

	c.logger.Info("Updating promo code", zap.Int("id", id))

	updatedPromo, err := c.promoUseCase.UpdatePromoCode(id, updates)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Promo code updated successfully",
		"data":    updatedPromo,
	})
}

func (c *PromoController) DeactivatePromoCode(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Deactivating promo code", zap.Int("id", id))

	err = c.promoUseCase.DeactivatePromoCode(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Promo code deactivated successfully",
	})
}
