package scheduled

import (
	"encoding/json"
	"time"

	
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/scheduled"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ScheduledRideRepository handles scheduled ride data access
type ScheduledRideRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

type ScheduledRideRepositoryInterface interface {
	Create(ride *scheduled.ScheduledRide) (*scheduled.ScheduledRide, error)
	GetByID(id int) (*scheduled.ScheduledRide, error)
	GetUserScheduledRides(userID int) (*[]scheduled.ScheduledRide, error)
	GetUpcomingScheduledRides(limit int) (*[]scheduled.ScheduledRide, error)
	Update(id int, updates map[string]interface{}) (*scheduled.ScheduledRide, error)
	Delete(id int) error
	GetReadyToExecute() (*[]scheduled.ScheduledRide, error)
	GetExpiredRides() (*[]scheduled.ScheduledRide, error)
	MarkAsExpired(id int) error
}

// NewScheduledRideRepository creates a new scheduled ride repository
func NewScheduledRideRepository(db *gorm.DB, logger *logger.Logger) *ScheduledRideRepository {
	return &ScheduledRideRepository{
		DB:     db,
		Logger: logger,
	}
}

func (r *ScheduledRideRepository) Create(ride *scheduled.ScheduledRide) (*scheduled.ScheduledRide, error) {
	result := r.DB.Create(ride)

	if result.Error != nil {
		var gormErr domainErrors.GormErr
		if err := json.Unmarshal([]byte(result.Error.Error()), &gormErr); err == nil {
			if gormErr.Number == 1062 {
				return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			}
		}
		r.Logger.Error("Error creating scheduled ride", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return ride, nil
}

func (r *ScheduledRideRepository) GetByID(id int) (*scheduled.ScheduledRide, error) {
	var ride scheduled.ScheduledRide
	result := r.DB.First(&ride, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &ride, nil
}

func (r *ScheduledRideRepository) GetUserScheduledRides(userID int) (*[]scheduled.ScheduledRide, error) {
	var rides []scheduled.ScheduledRide
	result := r.DB.Where("rider_id = ?", userID).
		Order("scheduled_time DESC").
		Find(&rides)

	if result.Error != nil {
		r.Logger.Error("Error getting user scheduled rides", 
			zap.Int("userID", userID), 
			zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &rides, nil
}

func (r *ScheduledRideRepository) GetUpcomingScheduledRides(limit int) (*[]scheduled.ScheduledRide, error) {
	var rides []scheduled.ScheduledRide
	now := time.Now()

	query := r.DB.Where("status IN (?, ?) AND scheduled_time > ?", 
		scheduled.ScheduledStatusPending, 
		scheduled.ScheduledStatusConfirmed, 
		now).
		Order("scheduled_time ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&rides)

	if result.Error != nil {
		r.Logger.Error("Error getting upcoming scheduled rides", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &rides, nil
}

func (r *ScheduledRideRepository) Update(id int, updates map[string]interface{}) (*scheduled.ScheduledRide, error) {
	var ride scheduled.ScheduledRide
	if err := r.DB.First(&ride, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	result := r.DB.Model(&ride).Updates(updates)
	if result.Error != nil {
		r.Logger.Error("Error updating scheduled ride", zap.Int("id", id), zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &ride, nil
}

func (r *ScheduledRideRepository) Delete(id int) error {
	result := r.DB.Delete(&scheduled.ScheduledRide{}, id)
	if result.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if result.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

func (r *ScheduledRideRepository) GetReadyToExecute() (*[]scheduled.ScheduledRide, error) {
	var rides []scheduled.ScheduledRide
	// Get rides that should be executed (10 minutes before scheduled time)
	executeTime := time.Now().Add(10 * time.Minute)

	result := r.DB.Where("status IN (?, ?) AND scheduled_time <= ?", 
		scheduled.ScheduledStatusPending, 
		scheduled.ScheduledStatusConfirmed, 
		executeTime).
		Order("scheduled_time ASC").
		Find(&rides)

	if result.Error != nil {
		r.Logger.Error("Error getting ready to execute scheduled rides", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &rides, nil
}

func (r *ScheduledRideRepository) GetExpiredRides() (*[]scheduled.ScheduledRide, error) {
	var rides []scheduled.ScheduledRide
	// Get rides that are 30 minutes past scheduled time and not started
	expireTime := time.Now().Add(-30 * time.Minute)

	result := r.DB.Where("status IN (?, ?) AND scheduled_time < ?", 
		scheduled.ScheduledStatusPending, 
		scheduled.ScheduledStatusConfirmed, 
		expireTime).
		Find(&rides)

	if result.Error != nil {
		r.Logger.Error("Error getting expired scheduled rides", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &rides, nil
}

func (r *ScheduledRideRepository) MarkAsExpired(id int) error {
	result := r.DB.Model(&scheduled.ScheduledRide{}).
		Where("id = ?", id).
		Update("status", scheduled.ScheduledStatusExpired)

	if result.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if result.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

