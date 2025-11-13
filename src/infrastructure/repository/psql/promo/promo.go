package promo

import (
	"encoding/json"
	"time"

	
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/promo"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PromoRepository handles promo code data access
type PromoRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

type PromoRepositoryInterface interface {
	GetByCode(code string) (*promo.PromoCode, error)
	GetByID(id int) (*promo.PromoCode, error)
	Create(promoCode *promo.PromoCode) (*promo.PromoCode, error)
	Update(id int, updates map[string]interface{}) (*promo.PromoCode, error)
	Delete(id int) error
	GetAllActive() (*[]promo.PromoCode, error)
	IncrementUsage(id int) error
	CreateUsage(usage *promo.PromoUsage) error
	GetUserUsageCount(userID int, promoCodeID int) (int, error)
	GetUserUsageHistory(userID int) (*[]promo.PromoUsage, error)
	GetAllPromos() (*[]promo.PromoCode, error)
}

// NewPromoRepository creates a new promo repository
func NewPromoRepository(db *gorm.DB, logger *logger.Logger) *PromoRepository {
	return &PromoRepository{
		DB:     db,
		Logger: logger,
	}
}

func (r *PromoRepository) GetByCode(code string) (*promo.PromoCode, error) {
	var promoCode promo.PromoCode
	result := r.DB.Where("code = ?", promo.NormalizeCode(code)).First(&promoCode)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		r.Logger.Error("Error getting promo code", zap.String("code", code), zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &promoCode, nil
}

func (r *PromoRepository) GetByID(id int) (*promo.PromoCode, error) {
	var promoCode promo.PromoCode
	result := r.DB.First(&promoCode, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &promoCode, nil
}

func (r *PromoRepository) Create(promoCode *promo.PromoCode) (*promo.PromoCode, error) {
	promoCode.Code = promo.NormalizeCode(promoCode.Code)
	result := r.DB.Create(promoCode)

	if result.Error != nil {
		var gormErr domainErrors.GormErr
		if err := json.Unmarshal([]byte(result.Error.Error()), &gormErr); err == nil {
			if gormErr.Number == 1062 {
				return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			}
		}
		r.Logger.Error("Error creating promo code", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return promoCode, nil
}

func (r *PromoRepository) Update(id int, updates map[string]interface{}) (*promo.PromoCode, error) {
	var promoCode promo.PromoCode
	if err := r.DB.First(&promoCode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	result := r.DB.Model(&promoCode).Updates(updates)
	if result.Error != nil {
		r.Logger.Error("Error updating promo code", zap.Int("id", id), zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &promoCode, nil
}

func (r *PromoRepository) Delete(id int) error {
	result := r.DB.Delete(&promo.PromoCode{}, id)
	if result.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if result.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

func (r *PromoRepository) GetAllActive() (*[]promo.PromoCode, error) {
	var promoCodes []promo.PromoCode
	now := time.Now()

	result := r.DB.Where("active = ? AND valid_from <= ? AND valid_until >= ?", 
		true, now, now).Find(&promoCodes)

	if result.Error != nil {
		r.Logger.Error("Error getting active promo codes", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &promoCodes, nil
}

func (r *PromoRepository) IncrementUsage(id int) error {
	result := r.DB.Model(&promo.PromoCode{}).Where("id = ?", id).
		UpdateColumn("current_usage", gorm.Expr("current_usage + 1"))

	if result.Error != nil {
		r.Logger.Error("Error incrementing promo usage", zap.Int("id", id), zap.Error(result.Error))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return nil
}

func (r *PromoRepository) CreateUsage(usage *promo.PromoUsage) error {
	result := r.DB.Create(usage)

	if result.Error != nil {
		r.Logger.Error("Error creating promo usage", zap.Error(result.Error))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return nil
}

func (r *PromoRepository) GetUserUsageCount(userID int, promoCodeID int) (int, error) {
	var count int64
	result := r.DB.Model(&promo.PromoUsage{}).
		Where("user_id = ? AND promo_code_id = ?", userID, promoCodeID).
		Count(&count)

	if result.Error != nil {
		r.Logger.Error("Error getting user promo usage count", 
			zap.Int("userID", userID), 
			zap.Int("promoCodeID", promoCodeID), 
			zap.Error(result.Error))
		return 0, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return int(count), nil
}

func (r *PromoRepository) GetUserUsageHistory(userID int) (*[]promo.PromoUsage, error) {
	var usages []promo.PromoUsage
	result := r.DB.Where("user_id = ?", userID).
		Order("used_at DESC").
		Find(&usages)

	if result.Error != nil {
		r.Logger.Error("Error getting user promo usage history", 
			zap.Int("userID", userID), 
			zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &usages, nil
}

func (r *PromoRepository) GetAllPromos() (*[]promo.PromoCode, error) {
	var promoCodes []promo.PromoCode
	result := r.DB.Order("created_at DESC").Find(&promoCodes)

	if result.Error != nil {
		r.Logger.Error("Error getting all promo codes", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &promoCodes, nil
}

