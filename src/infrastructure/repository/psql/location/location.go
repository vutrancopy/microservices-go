package location

import (
	"encoding/json"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/location"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LocationRepository handles favorite location data access
type LocationRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

type LocationRepositoryInterface interface {
	Create(location *location.FavoriteLocation) (*location.FavoriteLocation, error)
	GetByID(id int) (*location.FavoriteLocation, error)
	GetUserFavorites(userID int) (*[]location.FavoriteLocation, error)
	Update(id int, updates map[string]interface{}) (*location.FavoriteLocation, error)
	Delete(id int) error
	SetAsPrimary(id int, userID int, locationType location.LocationType) error
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *gorm.DB, logger *logger.Logger) *LocationRepository {
	return &LocationRepository{
		DB:     db,
		Logger: logger,
	}
}

func (r *LocationRepository) Create(loc *location.FavoriteLocation) (*location.FavoriteLocation, error) {
	// If this is set as primary, unset other primaries of same type
	if loc.IsPrimary {
		r.DB.Model(&location.FavoriteLocation{}).
			Where("user_id = ? AND type = ?", loc.UserID, loc.Type).
			Update("is_primary", false)
	}

	result := r.DB.Create(loc)

	if result.Error != nil {
		var gormErr domainErrors.GormErr
		if err := json.Unmarshal([]byte(result.Error.Error()), &gormErr); err == nil {
			if gormErr.Number == 1062 {
				return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			}
		}
		r.Logger.Error("Error creating favorite location", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return loc, nil
}

func (r *LocationRepository) GetByID(id int) (*location.FavoriteLocation, error) {
	var loc location.FavoriteLocation
	result := r.DB.First(&loc, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &loc, nil
}

func (r *LocationRepository) GetUserFavorites(userID int) (*[]location.FavoriteLocation, error) {
	var locations []location.FavoriteLocation
	result := r.DB.Where("user_id = ?", userID).
		Order("is_primary DESC, created_at DESC").
		Find(&locations)

	if result.Error != nil {
		r.Logger.Error("Error getting user favorite locations", 
			zap.Int("userID", userID), 
			zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &locations, nil
}

func (r *LocationRepository) Update(id int, updates map[string]interface{}) (*location.FavoriteLocation, error) {
	var loc location.FavoriteLocation
	if err := r.DB.First(&loc, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	// If updating to primary, unset other primaries of same type
	if isPrimary, ok := updates["is_primary"].(bool); ok && isPrimary {
		r.DB.Model(&location.FavoriteLocation{}).
			Where("user_id = ? AND type = ? AND id != ?", loc.UserID, loc.Type, id).
			Update("is_primary", false)
	}

	result := r.DB.Model(&loc).Updates(updates)
	if result.Error != nil {
		r.Logger.Error("Error updating favorite location", zap.Int("id", id), zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &loc, nil
}

func (r *LocationRepository) Delete(id int) error {
	result := r.DB.Delete(&location.FavoriteLocation{}, id)
	if result.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if result.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

func (r *LocationRepository) SetAsPrimary(id int, userID int, locationType location.LocationType) error {
	// Start transaction
	tx := r.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Unset all primaries of this type for user
	if err := tx.Model(&location.FavoriteLocation{}).
		Where("user_id = ? AND type = ?", userID, locationType).
		Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	// Set this one as primary
	if err := tx.Model(&location.FavoriteLocation{}).
		Where("id = ?", id).
		Update("is_primary", true).Error; err != nil {
		tx.Rollback()
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return tx.Commit().Error
}
