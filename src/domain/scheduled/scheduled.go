package scheduled

import (
	"errors"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain/common"
)

// ScheduledRideStatus represents the status of a scheduled ride
type ScheduledRideStatus string

const (
	ScheduledStatusPending   ScheduledRideStatus = "pending"    // Ride is scheduled
	ScheduledStatusConfirmed ScheduledRideStatus = "confirmed"  // Driver pre-assigned
	ScheduledStatusActive    ScheduledRideStatus = "active"     // Ride has started
	ScheduledStatusCancelled ScheduledRideStatus = "cancelled"  // Cancelled before start
	ScheduledStatusCompleted ScheduledRideStatus = "completed"  // Ride completed
	ScheduledStatusExpired   ScheduledRideStatus = "expired"    // Scheduled time passed without execution
)

// ScheduledRide represents a ride booked for future time
type ScheduledRide struct {
	ID                  int
	RiderID             int
	DriverID            *int // Can be pre-assigned or matched later
	VehicleType         common.VehicleType
	PickupLatitude      float64
	PickupLongitude     float64
	PickupAddress       string
	DropoffLatitude     float64
	DropoffLongitude    float64
	DropoffAddress      string
	ScheduledTime       time.Time // When ride should start
	EstimatedDistance   float64
	EstimatedDuration   int
	EstimatedFare       float64
	Status              ScheduledRideStatus
	ActualRideID        *int    // ID of actual ride when executed
	Notes               string
	PromoCode           string
	CancellationReason  string
	CancelledAt         *time.Time
	ConfirmedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Validate validates the scheduled ride
func (sr *ScheduledRide) Validate() error {
	if sr.RiderID == 0 {
		return errors.New("rider ID is required")
	}

	if sr.PickupAddress == "" {
		return errors.New("pickup address is required")
	}

	if sr.DropoffAddress == "" {
		return errors.New("dropoff address is required")
	}

	if sr.ScheduledTime.IsZero() {
		return errors.New("scheduled time is required")
	}

	// Must be scheduled at least 30 minutes in advance
	minAdvanceTime := time.Now().Add(30 * time.Minute)
	if sr.ScheduledTime.Before(minAdvanceTime) {
		return errors.New("ride must be scheduled at least 30 minutes in advance")
	}

	// Cannot schedule more than 7 days in advance
	maxAdvanceTime := time.Now().Add(7 * 24 * time.Hour)
	if sr.ScheduledTime.After(maxAdvanceTime) {
		return errors.New("ride cannot be scheduled more than 7 days in advance")
	}

	return nil
}

// CanBeCancelled checks if ride can be cancelled
func (sr *ScheduledRide) CanBeCancelled() bool {
	// Can cancel if not already cancelled, completed, or active
	return sr.Status == ScheduledStatusPending || sr.Status == ScheduledStatusConfirmed
}

// Cancel cancels the scheduled ride
func (sr *ScheduledRide) Cancel(reason string) error {
	if !sr.CanBeCancelled() {
		return errors.New("ride cannot be cancelled at this stage")
	}

	sr.Status = ScheduledStatusCancelled
	sr.CancellationReason = reason
	now := time.Now()
	sr.CancelledAt = &now

	return nil
}

// Confirm confirms the ride (driver pre-assigned)
func (sr *ScheduledRide) Confirm(driverID int) error {
	if sr.Status != ScheduledStatusPending {
		return errors.New("only pending rides can be confirmed")
	}

	sr.Status = ScheduledStatusConfirmed
	sr.DriverID = &driverID
	now := time.Now()
	sr.ConfirmedAt = &now

	return nil
}

// Activate activates the ride (starts execution)
func (sr *ScheduledRide) Activate(actualRideID int) error {
	if sr.Status != ScheduledStatusPending && sr.Status != ScheduledStatusConfirmed {
		return errors.New("ride is not in valid state to activate")
	}

	sr.Status = ScheduledStatusActive
	sr.ActualRideID = &actualRideID

	return nil
}

// IsReadyToExecute checks if scheduled ride is ready to be executed
func (sr *ScheduledRide) IsReadyToExecute() bool {
	// Execute 10 minutes before scheduled time
	executeTime := sr.ScheduledTime.Add(-10 * time.Minute)
	return time.Now().After(executeTime) && 
		(sr.Status == ScheduledStatusPending || sr.Status == ScheduledStatusConfirmed)
}

// IsExpired checks if scheduled time has passed without execution
func (sr *ScheduledRide) IsExpired() bool {
	// Consider expired if 30 minutes past scheduled time and not active
	expireTime := sr.ScheduledTime.Add(30 * time.Minute)
	return time.Now().After(expireTime) && 
		(sr.Status == ScheduledStatusPending || sr.Status == ScheduledStatusConfirmed)
}

// SearchResultScheduledRide represents paginated scheduled rides
type SearchResultScheduledRide struct {
	Data       *[]ScheduledRide
	Total      int64
	Page       int64
	PageSize   int64
	TotalPages int64
}

// IScheduledRideService defines the interface for scheduled ride operations
type IScheduledRideService interface {
	CreateScheduledRide(ride *ScheduledRide) (*ScheduledRide, error)
	GetByID(id int) (*ScheduledRide, error)
	GetUserScheduledRides(userID int) (*[]ScheduledRide, error)
	GetUpcomingScheduledRides(limit int) (*[]ScheduledRide, error)
	CancelScheduledRide(id int, reason string) error
	ConfirmScheduledRide(id int, driverID int) error
	ActivateScheduledRide(id int, actualRideID int) error
	MarkAsExpired(id int) error
	ProcessScheduledRides() error // Background job to execute scheduled rides
	Update(id int, updates map[string]interface{}) (*ScheduledRide, error)
}
