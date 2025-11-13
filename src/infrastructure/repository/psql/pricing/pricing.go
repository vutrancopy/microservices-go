package pricing

import (
	"encoding/json"

	
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/pricing"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PricingRepository handles pricing configuration data access
type PricingRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

type PricingRepositoryInterface interface {
	GetByVehicleType(vehicleType string) (*pricing.PricingConfig, error)
	Create(config *pricing.PricingConfig) (*pricing.PricingConfig, error)
	Update(id int, updates map[string]interface{}) (*pricing.PricingConfig, error)
	GetAll() (*[]pricing.PricingConfig, error)
	GetActive() (*[]pricing.PricingConfig, error)
}

// NewPricingRepository creates a new pricing repository
func NewPricingRepository(db *gorm.DB, logger *logger.Logger) *PricingRepository {
	return &PricingRepository{
		DB:     db,
		Logger: logger,
	}
}

func (r *PricingRepository) GetByVehicleType(vehicleType string) (*pricing.PricingConfig, error) {
	var config pricing.PricingConfig
	result := r.DB.Where("vehicle_type = ? AND active = ?", vehicleType, true).First(&config)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		r.Logger.Error("Error getting pricing config by vehicle type",
			zap.String("vehicleType", vehicleType),
			zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &config, nil
}

func (r *PricingRepository) Create(config *pricing.PricingConfig) (*pricing.PricingConfig, error) {
	result := r.DB.Create(config)

	if result.Error != nil {
		var gormErr domainErrors.GormErr
		if err := json.Unmarshal([]byte(result.Error.Error()), &gormErr); err == nil {
			if gormErr.Number == 1062 {
				return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			}
		}
		r.Logger.Error("Error creating pricing config", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return config, nil
}

func (r *PricingRepository) Update(id int, updates map[string]interface{}) (*pricing.PricingConfig, error) {
	var config pricing.PricingConfig
	if err := r.DB.First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	result := r.DB.Model(&config).Updates(updates)
	if result.Error != nil {
		r.Logger.Error("Error updating pricing config", zap.Int("id", id), zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &config, nil
}

func (r *PricingRepository) GetAll() (*[]pricing.PricingConfig, error) {
	var configs []pricing.PricingConfig
	result := r.DB.Order("vehicle_type ASC").Find(&configs)

	if result.Error != nil {
		r.Logger.Error("Error getting all pricing configs", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &configs, nil
}

func (r *PricingRepository) GetActive() (*[]pricing.PricingConfig, error) {
	var configs []pricing.PricingConfig
	result := r.DB.Where("active = ?", true).Order("vehicle_type ASC").Find(&configs)

	if result.Error != nil {
		r.Logger.Error("Error getting active pricing configs", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &configs, nil
}

