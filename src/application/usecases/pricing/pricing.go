package pricing

import (
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/pricing"
	pricingRepo "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/pricing"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

type IPricingUseCase interface {
	GetPricingByVehicleType(vehicleType common.VehicleType) (*pricing.PricingConfig, error)
	CalculateFareEstimate(request *pricing.FareEstimateRequest) (*pricing.FareBreakdown, error)
	GetCancellationFee(vehicleType common.VehicleType, minutesSinceBooking int) (float64, error)
	UpdatePricingConfig(config *pricing.PricingConfig) (*pricing.PricingConfig, error)
	GetAllPricingConfigs() (*[]pricing.PricingConfig, error)
	GetActivePricingConfigs() (*[]pricing.PricingConfig, error)
	CreatePricingConfig(config *pricing.PricingConfig) (*pricing.PricingConfig, error)
}

type PricingUseCase struct {
	pricingRepo *pricingRepo.PricingRepository
	logger      *logger.Logger
}

func NewPricingUseCase(
	pricingRepo *pricingRepo.PricingRepository,
	logger *logger.Logger,
) *PricingUseCase {
	return &PricingUseCase{
		pricingRepo: pricingRepo,
		logger:      logger,
	}
}

func (u *PricingUseCase) GetPricingByVehicleType(vehicleType common.VehicleType) (*pricing.PricingConfig, error) {
	u.logger.Info("Getting pricing for vehicle type", zap.String("vehicleType", string(vehicleType)))

	config, err := u.pricingRepo.GetByVehicleType(string(vehicleType))
	if err != nil {
		u.logger.Error("Error getting pricing config", zap.Error(err))
		return nil, err
	}

	return config, nil
}

func (u *PricingUseCase) CalculateFareEstimate(request *pricing.FareEstimateRequest) (*pricing.FareBreakdown, error) {
	u.logger.Info("Calculating fare estimate", 
		zap.String("vehicleType", string(request.VehicleType)),
		zap.Float64("distance", request.DistanceKm))

	// Get pricing config for vehicle type
	config, err := u.pricingRepo.GetByVehicleType(string(request.VehicleType))
	if err != nil {
		u.logger.Error("Error getting pricing config for fare calculation", zap.Error(err))
		return nil, err
	}

	// Apply surge if provided
	if request.SurgeMultiplier > 1.0 {
		config.ApplySurge(request.SurgeMultiplier)
	}

	// Calculate fare (promo discount will be applied by promo usecase)
	fareBreakdown := config.CalculateFare(
		request.DistanceKm,
		request.EstimatedMinutes,
		request.WaitingMinutes,
		0, // No promo discount yet
	)

	u.logger.Info("Fare calculated", 
		zap.Float64("total", fareBreakdown.Total),
		zap.Float64("baseFare", fareBreakdown.BaseFare))

	return fareBreakdown, nil
}

func (u *PricingUseCase) GetCancellationFee(vehicleType common.VehicleType, minutesSinceBooking int) (float64, error) {
	u.logger.Info("Calculating cancellation fee", 
		zap.String("vehicleType", string(vehicleType)),
		zap.Int("minutesSinceBooking", minutesSinceBooking))

	config, err := u.pricingRepo.GetByVehicleType(string(vehicleType))
	if err != nil {
		u.logger.Error("Error getting pricing config for cancellation fee", zap.Error(err))
		return 0, err
	}

	fee := config.GetCancellationFee(minutesSinceBooking)
	u.logger.Info("Cancellation fee calculated", zap.Float64("fee", fee))

	return fee, nil
}

func (u *PricingUseCase) UpdatePricingConfig(config *pricing.PricingConfig) (*pricing.PricingConfig, error) {
	u.logger.Info("Updating pricing config", zap.Int("id", config.ID))

	updates := map[string]interface{}{
		"base_fare":         config.BaseFare,
		"cost_per_km":       config.CostPerKm,
		"cost_per_minute":   config.CostPerMinute,
		"minimum_fare":      config.MinimumFare,
		"service_fee":       config.ServiceFee,
		"booking_fee":       config.BookingFee,
		"cancellation_fee":  config.CancellationFee,
		"waiting_time_rate": config.WaitingTimeRate,
		"surge_multiplier":  config.SurgeMultiplier,
		"active":            config.Active,
	}

	updatedConfig, err := u.pricingRepo.Update(config.ID, updates)
	if err != nil {
		u.logger.Error("Error updating pricing config", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Pricing config updated successfully", zap.Int("id", updatedConfig.ID))
	return updatedConfig, nil
}

func (u *PricingUseCase) GetAllPricingConfigs() (*[]pricing.PricingConfig, error) {
	u.logger.Info("Getting all pricing configs")

	configs, err := u.pricingRepo.GetAll()
	if err != nil {
		u.logger.Error("Error getting all pricing configs", zap.Error(err))
		return nil, err
	}

	return configs, nil
}

func (u *PricingUseCase) GetActivePricingConfigs() (*[]pricing.PricingConfig, error) {
	u.logger.Info("Getting active pricing configs")

	configs, err := u.pricingRepo.GetActive()
	if err != nil {
		u.logger.Error("Error getting active pricing configs", zap.Error(err))
		return nil, err
	}

	return configs, nil
}

func (u *PricingUseCase) CreatePricingConfig(config *pricing.PricingConfig) (*pricing.PricingConfig, error) {
	u.logger.Info("Creating new pricing config", zap.String("vehicleType", string(config.VehicleType)))

	// Validate pricing config
	if config.BaseFare < 0 || config.CostPerKm < 0 || config.CostPerMinute < 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	createdConfig, err := u.pricingRepo.Create(config)
	if err != nil {
		u.logger.Error("Error creating pricing config", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Pricing config created successfully", zap.Int("id", createdConfig.ID))
	return createdConfig, nil
}
