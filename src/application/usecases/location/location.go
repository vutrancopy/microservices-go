package location

import (
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/location"
	locationRepo "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/location"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

type ILocationUseCase interface {
	CreateFavorite(location *location.FavoriteLocation) (*location.FavoriteLocation, error)
	GetUserFavorites(userID int) (*[]location.FavoriteLocation, error)
	GetByID(id int) (*location.FavoriteLocation, error)
	UpdateFavorite(id int, updates map[string]interface{}) (*location.FavoriteLocation, error)
	DeleteFavorite(id int) error
	SetAsPrimary(id int, userID int) error
}

type LocationUseCase struct {
	locationRepo *locationRepo.LocationRepository
	logger       *logger.Logger
}

func NewLocationUseCase(
	locationRepo *locationRepo.LocationRepository,
	logger *logger.Logger,
) *LocationUseCase {
	return &LocationUseCase{
		locationRepo: locationRepo,
		logger:       logger,
	}
}

func (u *LocationUseCase) CreateFavorite(loc *location.FavoriteLocation) (*location.FavoriteLocation, error) {
	u.logger.Info("Creating favorite location", 
		zap.Int("userID", loc.UserID),
		zap.String("label", loc.Label))

	// Validate location
	if err := loc.Validate(); err != nil {
		u.logger.Error("Invalid favorite location", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	createdLocation, err := u.locationRepo.Create(loc)
	if err != nil {
		u.logger.Error("Error creating favorite location", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Favorite location created successfully", zap.Int("id", createdLocation.ID))
	return createdLocation, nil
}

func (u *LocationUseCase) GetUserFavorites(userID int) (*[]location.FavoriteLocation, error) {
	u.logger.Info("Getting user favorite locations", zap.Int("userID", userID))

	locations, err := u.locationRepo.GetUserFavorites(userID)
	if err != nil {
		u.logger.Error("Error getting user favorite locations", zap.Error(err))
		return nil, err
	}

	return locations, nil
}

func (u *LocationUseCase) GetByID(id int) (*location.FavoriteLocation, error) {
	u.logger.Info("Getting favorite location by ID", zap.Int("id", id))

	loc, err := u.locationRepo.GetByID(id)
	if err != nil {
		u.logger.Error("Error getting favorite location", zap.Error(err))
		return nil, err
	}

	return loc, nil
}

func (u *LocationUseCase) UpdateFavorite(id int, updates map[string]interface{}) (*location.FavoriteLocation, error) {
	u.logger.Info("Updating favorite location", zap.Int("id", id))

	updatedLocation, err := u.locationRepo.Update(id, updates)
	if err != nil {
		u.logger.Error("Error updating favorite location", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Favorite location updated successfully", zap.Int("id", id))
	return updatedLocation, nil
}

func (u *LocationUseCase) DeleteFavorite(id int) error {
	u.logger.Info("Deleting favorite location", zap.Int("id", id))

	err := u.locationRepo.Delete(id)
	if err != nil {
		u.logger.Error("Error deleting favorite location", zap.Error(err))
		return err
	}

	u.logger.Info("Favorite location deleted successfully", zap.Int("id", id))
	return nil
}

func (u *LocationUseCase) SetAsPrimary(id int, userID int) error {
	u.logger.Info("Setting favorite location as primary", 
		zap.Int("id", id),
		zap.Int("userID", userID))

	// Get location to check its type
	loc, err := u.locationRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Verify ownership
	if loc.UserID != userID {
		return domainErrors.NewAppErrorWithType(domainErrors.NotAuthorized)
	}

	err = u.locationRepo.SetAsPrimary(id, userID, loc.Type)
	if err != nil {
		u.logger.Error("Error setting favorite location as primary", zap.Error(err))
		return err
	}

	u.logger.Info("Favorite location set as primary successfully", zap.Int("id", id))
	return nil
}
