package scheduled

import (
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/scheduled"
	scheduledRepo "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/scheduled"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

type IScheduledRideUseCase interface {
	CreateScheduledRide(ride *scheduled.ScheduledRide) (*scheduled.ScheduledRide, error)
	GetByID(id int) (*scheduled.ScheduledRide, error)
	GetUserScheduledRides(userID int) (*[]scheduled.ScheduledRide, error)
	GetUpcomingScheduledRides(limit int) (*[]scheduled.ScheduledRide, error)
	CancelScheduledRide(id int, reason string, userID int) error
	ConfirmScheduledRide(id int, driverID int) error
	UpdateScheduledRide(id int, updates map[string]interface{}) (*scheduled.ScheduledRide, error)
	ProcessScheduledRides() error
}

type ScheduledRideUseCase struct {
	scheduledRepo *scheduledRepo.ScheduledRideRepository
	logger        *logger.Logger
}

func NewScheduledRideUseCase(
	scheduledRepo *scheduledRepo.ScheduledRideRepository,
	logger *logger.Logger,
) *ScheduledRideUseCase {
	return &ScheduledRideUseCase{
		scheduledRepo: scheduledRepo,
		logger:        logger,
	}
}

func (u *ScheduledRideUseCase) CreateScheduledRide(ride *scheduled.ScheduledRide) (*scheduled.ScheduledRide, error) {
	u.logger.Info("Creating scheduled ride", 
		zap.Int("riderID", ride.RiderID),
		zap.Time("scheduledTime", ride.ScheduledTime))

	// Validate scheduled ride
	if err := ride.Validate(); err != nil {
		u.logger.Error("Invalid scheduled ride", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Set initial status
	ride.Status = scheduled.ScheduledStatusPending

	createdRide, err := u.scheduledRepo.Create(ride)
	if err != nil {
		u.logger.Error("Error creating scheduled ride", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Scheduled ride created successfully", zap.Int("id", createdRide.ID))
	return createdRide, nil
}

func (u *ScheduledRideUseCase) GetByID(id int) (*scheduled.ScheduledRide, error) {
	u.logger.Info("Getting scheduled ride by ID", zap.Int("id", id))

	ride, err := u.scheduledRepo.GetByID(id)
	if err != nil {
		u.logger.Error("Error getting scheduled ride", zap.Error(err))
		return nil, err
	}

	return ride, nil
}

func (u *ScheduledRideUseCase) GetUserScheduledRides(userID int) (*[]scheduled.ScheduledRide, error) {
	u.logger.Info("Getting user scheduled rides", zap.Int("userID", userID))

	rides, err := u.scheduledRepo.GetUserScheduledRides(userID)
	if err != nil {
		u.logger.Error("Error getting user scheduled rides", zap.Error(err))
		return nil, err
	}

	return rides, nil
}

func (u *ScheduledRideUseCase) GetUpcomingScheduledRides(limit int) (*[]scheduled.ScheduledRide, error) {
	u.logger.Info("Getting upcoming scheduled rides", zap.Int("limit", limit))

	rides, err := u.scheduledRepo.GetUpcomingScheduledRides(limit)
	if err != nil {
		u.logger.Error("Error getting upcoming scheduled rides", zap.Error(err))
		return nil, err
	}

	return rides, nil
}

func (u *ScheduledRideUseCase) CancelScheduledRide(id int, reason string, userID int) error {
	u.logger.Info("Cancelling scheduled ride", 
		zap.Int("id", id),
		zap.String("reason", reason),
		zap.Int("userID", userID))

	// Get scheduled ride
	ride, err := u.scheduledRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Verify ownership (only rider can cancel)
	if ride.RiderID != userID {
		return domainErrors.NewAppErrorWithType(domainErrors.NotAuthorized)
	}

	// Cancel ride
	if err := ride.Cancel(reason); err != nil {
		u.logger.Error("Cannot cancel scheduled ride", zap.Error(err))
		return domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Update in database
	updates := map[string]interface{}{
		"status":              ride.Status,
		"cancellation_reason": ride.CancellationReason,
		"cancelled_at":        ride.CancelledAt,
	}

	_, err = u.scheduledRepo.Update(id, updates)
	if err != nil {
		u.logger.Error("Error cancelling scheduled ride", zap.Error(err))
		return err
	}

	u.logger.Info("Scheduled ride cancelled successfully", zap.Int("id", id))
	return nil
}

func (u *ScheduledRideUseCase) ConfirmScheduledRide(id int, driverID int) error {
	u.logger.Info("Confirming scheduled ride", 
		zap.Int("id", id),
		zap.Int("driverID", driverID))

	// Get scheduled ride
	ride, err := u.scheduledRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Confirm ride
	if err := ride.Confirm(driverID); err != nil {
		u.logger.Error("Cannot confirm scheduled ride", zap.Error(err))
		return domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Update in database
	updates := map[string]interface{}{
		"status":       ride.Status,
		"driver_id":    ride.DriverID,
		"confirmed_at": ride.ConfirmedAt,
	}

	_, err = u.scheduledRepo.Update(id, updates)
	if err != nil {
		u.logger.Error("Error confirming scheduled ride", zap.Error(err))
		return err
	}

	u.logger.Info("Scheduled ride confirmed successfully", zap.Int("id", id))
	return nil
}

func (u *ScheduledRideUseCase) UpdateScheduledRide(id int, updates map[string]interface{}) (*scheduled.ScheduledRide, error) {
	u.logger.Info("Updating scheduled ride", zap.Int("id", id))

	updatedRide, err := u.scheduledRepo.Update(id, updates)
	if err != nil {
		u.logger.Error("Error updating scheduled ride", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Scheduled ride updated successfully", zap.Int("id", id))
	return updatedRide, nil
}

func (u *ScheduledRideUseCase) ProcessScheduledRides() error {
	u.logger.Info("Processing scheduled rides")

	// Get rides ready to execute
	readyRides, err := u.scheduledRepo.GetReadyToExecute()
	if err != nil {
		return err
	}

	u.logger.Info("Found rides ready to execute", zap.Int("count", len(*readyRides)))

	// TODO: For each ready ride, create actual ride
	// This would typically call RideUseCase.CreateRide()
	// For now, just log

	// Check and mark expired rides
	expiredRides, err := u.scheduledRepo.GetExpiredRides()
	if err != nil {
		return err
	}

	for _, ride := range *expiredRides {
		if err := u.scheduledRepo.MarkAsExpired(ride.ID); err != nil {
			u.logger.Error("Error marking scheduled ride as expired", 
				zap.Int("id", ride.ID), 
				zap.Error(err))
		}
	}

	u.logger.Info("Scheduled rides processing completed", 
		zap.Int("expired", len(*expiredRides)))

	return nil
}
