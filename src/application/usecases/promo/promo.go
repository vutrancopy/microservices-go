package promo

import (
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/promo"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	promoRepo "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/promo"
	"go.uber.org/zap"
)

type IPromoUseCase interface {
	GetByCode(code string) (*promo.PromoCode, error)
	ValidatePromoCode(code string, userID int, rideFare float64) (*promo.PromoCode, float64, error)
	ApplyPromoCode(promoCodeID int, userID int, rideID int, discountAmount float64) error
	CreatePromoCode(promoCode *promo.PromoCode) (*promo.PromoCode, error)
	UpdatePromoCode(id int, updates map[string]interface{}) (*promo.PromoCode, error)
	DeactivatePromoCode(id int) error
	GetAllActivePromos() (*[]promo.PromoCode, error)
	GetUserPromoUsageHistory(userID int) (*[]promo.PromoUsage, error)
	GetAllPromos() (*[]promo.PromoCode, error)
}

type PromoUseCase struct {
	promoRepo *promoRepo.PromoRepository
	logger    *logger.Logger
}

func NewPromoUseCase(
	promoRepo *promoRepo.PromoRepository,
	logger *logger.Logger,
) *PromoUseCase {
	return &PromoUseCase{
		promoRepo: promoRepo,
		logger:    logger,
	}
}

func (u *PromoUseCase) GetByCode(code string) (*promo.PromoCode, error) {
	u.logger.Info("Getting promo code", zap.String("code", code))

	promoCode, err := u.promoRepo.GetByCode(code)
	if err != nil {
		u.logger.Error("Error getting promo code", zap.Error(err))
		return nil, err
	}

	return promoCode, nil
}

func (u *PromoUseCase) ValidatePromoCode(code string, userID int, rideFare float64) (*promo.PromoCode, float64, error) {
	u.logger.Info("Validating promo code", 
		zap.String("code", code), 
		zap.Int("userID", userID),
		zap.Float64("rideFare", rideFare))

	// Get promo code
	promoCode, err := u.promoRepo.GetByCode(code)
	if err != nil {
		u.logger.Error("Promo code not found", zap.Error(err))
		return nil, 0, err
	}

	// Check if promo is valid
	if err := promoCode.IsValid(); err != nil {
		u.logger.Info("Promo code is not valid", zap.Error(err))
		return nil, 0, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Check user usage count
	userUsageCount, err := u.promoRepo.GetUserUsageCount(userID, promoCode.ID)
	if err != nil {
		u.logger.Error("Error getting user promo usage count", zap.Error(err))
		return nil, 0, err
	}

	// Check if this is user's first ride
	isFirstRide, err := u.isUserFirstRide(userID)
	if err != nil {
		u.logger.Error("Error checking if user has first ride", zap.Error(err))
		return nil, 0, err
	}

	// Check if user can use this promo
	if err := promoCode.CanUserUse(userID, userUsageCount, isFirstRide); err != nil {
		u.logger.Info("User cannot use this promo code", zap.Error(err))
		return nil, 0, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Calculate discount
	discount, err := promoCode.CalculateDiscount(rideFare)
	if err != nil {
		u.logger.Info("Error calculating discount", zap.Error(err))
		return nil, 0, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	u.logger.Info("Promo code validated successfully", 
		zap.String("code", code),
		zap.Float64("discount", discount))

	return promoCode, discount, nil
}

func (u *PromoUseCase) ApplyPromoCode(promoCodeID int, userID int, rideID int, discountAmount float64) error {
	u.logger.Info("Applying promo code", 
		zap.Int("promoCodeID", promoCodeID),
		zap.Int("userID", userID),
		zap.Int("rideID", rideID),
		zap.Float64("discountAmount", discountAmount))

	// Create promo usage record
	usage := &promo.PromoUsage{
		PromoCodeID:    promoCodeID,
		UserID:         userID,
		RideID:         rideID,
		DiscountAmount: discountAmount,
		UsedAt:         time.Now(),
	}

	if err := u.promoRepo.CreateUsage(usage); err != nil {
		u.logger.Error("Error creating promo usage", zap.Error(err))
		return err
	}

	// Increment promo code usage count
	if err := u.promoRepo.IncrementUsage(promoCodeID); err != nil {
		u.logger.Error("Error incrementing promo usage", zap.Error(err))
		return err
	}

	u.logger.Info("Promo code applied successfully")
	return nil
}

func (u *PromoUseCase) CreatePromoCode(promoCode *promo.PromoCode) (*promo.PromoCode, error) {
	u.logger.Info("Creating new promo code", zap.String("code", promoCode.Code))

	// Validate promo code
	if promoCode.Code == "" || promoCode.Type == "" {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	if promoCode.Value <= 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	if promoCode.ValidFrom.IsZero() || promoCode.ValidUntil.IsZero() {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	if promoCode.ValidFrom.After(promoCode.ValidUntil) {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	createdPromo, err := u.promoRepo.Create(promoCode)
	if err != nil {
		u.logger.Error("Error creating promo code", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Promo code created successfully", zap.Int("id", createdPromo.ID))
	return createdPromo, nil
}

func (u *PromoUseCase) UpdatePromoCode(id int, updates map[string]interface{}) (*promo.PromoCode, error) {
	u.logger.Info("Updating promo code", zap.Int("id", id))

	updatedPromo, err := u.promoRepo.Update(id, updates)
	if err != nil {
		u.logger.Error("Error updating promo code", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Promo code updated successfully", zap.Int("id", id))
	return updatedPromo, nil
}

func (u *PromoUseCase) DeactivatePromoCode(id int) error {
	u.logger.Info("Deactivating promo code", zap.Int("id", id))

	updates := map[string]interface{}{
		"active": false,
	}

	_, err := u.promoRepo.Update(id, updates)
	if err != nil {
		u.logger.Error("Error deactivating promo code", zap.Error(err))
		return err
	}

	u.logger.Info("Promo code deactivated successfully", zap.Int("id", id))
	return nil
}

func (u *PromoUseCase) GetAllActivePromos() (*[]promo.PromoCode, error) {
	u.logger.Info("Getting all active promo codes")

	promos, err := u.promoRepo.GetAllActive()
	if err != nil {
		u.logger.Error("Error getting active promo codes", zap.Error(err))
		return nil, err
	}

	return promos, nil
}

func (u *PromoUseCase) GetUserPromoUsageHistory(userID int) (*[]promo.PromoUsage, error) {
	u.logger.Info("Getting user promo usage history", zap.Int("userID", userID))

	usages, err := u.promoRepo.GetUserUsageHistory(userID)
	if err != nil {
		u.logger.Error("Error getting user promo usage history", zap.Error(err))
		return nil, err
	}

	return usages, nil
}

func (u *PromoUseCase) GetAllPromos() (*[]promo.PromoCode, error) {
	u.logger.Info("Getting all promo codes")

	promos, err := u.promoRepo.GetAllPromos()
	if err != nil {
		u.logger.Error("Error getting all promo codes", zap.Error(err))
		return nil, err
	}

	return promos, nil
}

// Helper method to check if user has completed any rides
func (u *PromoUseCase) isUserFirstRide(userID int) (bool, error) {
	// For simplicity, we return false (not first ride)
	// In a real implementation, this would query the ride repository
	// to check if the user has any completed rides
	// TODO: Implement proper first ride detection
	return false, nil
}
