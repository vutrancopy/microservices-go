package ride

import (
	"errors"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
)

// Location represents a geographical location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

// Ride represents a ride request/booking in the system
type Ride struct {
	ID                int
	RiderID           int    // Foreign key to User (rider)
	DriverID          *int   // Foreign key to User (driver), nullable until matched
	VehicleID         *int   // Foreign key to Vehicle, nullable until matched
	Status            common.RideStatus
	PickupLatitude    float64
	PickupLongitude   float64
	PickupAddress     string
	DropoffLatitude   float64
	DropoffLongitude  float64
	DropoffAddress    string
	RequestedAt       time.Time
	MatchedAt         *time.Time
	StartedAt         *time.Time
	CompletedAt       *time.Time
	CancelledAt       *time.Time
	CancellationReason string
	EstimatedDistance float64 // in kilometers
	ActualDistance    float64 // in kilometers
	EstimatedDuration int     // in minutes
	ActualDuration    int     // in minutes
	EstimatedFare     float64
	ActualFare        float64
	RiderRating       *int    // Rating given by rider (1-5)
	DriverRating      *int    // Rating given by driver (1-5)
	Notes             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// GetPickupLocation returns the pickup location
func (r *Ride) GetPickupLocation() Location {
	return Location{
		Latitude:  r.PickupLatitude,
		Longitude: r.PickupLongitude,
		Address:   r.PickupAddress,
	}
}

// GetDropoffLocation returns the dropoff location
func (r *Ride) GetDropoffLocation() Location {
	return Location{
		Latitude:  r.DropoffLatitude,
		Longitude: r.DropoffLongitude,
		Address:   r.DropoffAddress,
	}
}

// IsCompleted checks if the ride is completed
func (r *Ride) IsCompleted() bool {
	return r.Status == common.RideStatusCompleted
}

// IsCancelled checks if the ride is cancelled
func (r *Ride) IsCancelled() bool {
	return r.Status == common.RideStatusCancelled
}

// IsInProgress checks if the ride is in progress
func (r *Ride) IsInProgress() bool {
	return r.Status == common.RideStatusInProgress
}

// IsPending checks if the ride is pending
func (r *Ride) IsPending() bool {
	return r.Status == common.RideStatusPending
}

// IsMatched checks if the ride has been matched with a driver
func (r *Ride) IsMatched() bool {
	return r.Status == common.RideStatusMatched
}

// CanBeCancelled checks if the ride can be cancelled
func (r *Ride) CanBeCancelled() bool {
	return r.Status == common.RideStatusPending ||
		r.Status == common.RideStatusMatched
}

// CanStart checks if the ride can be started
func (r *Ride) CanStart() bool {
	return r.Status == common.RideStatusMatched
}

// CanComplete checks if the ride can be completed
func (r *Ride) CanComplete() bool {
	return r.Status == common.RideStatusInProgress
}

// AssignDriver assigns a driver to the ride
func (r *Ride) AssignDriver(driverID int, vehicleID int) error {
	if r.Status != common.RideStatusPending {
		return errors.New("can only assign driver to pending rides")
	}
	r.DriverID = &driverID
	r.VehicleID = &vehicleID
	now := time.Now()
	r.MatchedAt = &now
	return nil
}

// Start marks the ride as started
func (r *Ride) Start() error {
	if !r.CanStart() {
		return errors.New("ride cannot be started in current status")
	}
	now := time.Now()
	r.StartedAt = &now
	return nil
}

// Complete marks the ride as completed
func (r *Ride) Complete(actualDistance float64, actualDuration int, actualFare float64) error {
	if !r.CanComplete() {
		return errors.New("ride cannot be completed in current status")
	}
	now := time.Now()
	r.CompletedAt = &now
	r.ActualDistance = actualDistance
	r.ActualDuration = actualDuration
	r.ActualFare = actualFare
	return nil
}

// Cancel marks the ride as cancelled
func (r *Ride) Cancel(reason string) error {
	if !r.CanBeCancelled() {
		return errors.New("ride cannot be cancelled in current status")
	}
	now := time.Now()
	r.CancelledAt = &now
	r.CancellationReason = reason
	return nil
}

// SetRiderRating sets the rating given by the rider
func (r *Ride) SetRiderRating(rating int) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	if !r.IsCompleted() {
		return errors.New("can only rate completed rides")
	}
	r.RiderRating = &rating
	return nil
}

// SetDriverRating sets the rating given by the driver
func (r *Ride) SetDriverRating(rating int) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	if !r.IsCompleted() {
		return errors.New("can only rate completed rides")
	}
	r.DriverRating = &rating
	return nil
}

// SearchResultRide represents paginated ride search results
type SearchResultRide struct {
	Data       *[]Ride
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// IRideService defines the interface for ride operations
type IRideService interface {
	GetAll() (*[]Ride, error)
	GetByID(id int) (*Ride, error)
	GetByRiderID(riderID int) (*[]Ride, error)
	GetByDriverID(driverID int) (*[]Ride, error)
	GetPendingRides() (*[]Ride, error)
	Create(newRide *Ride) (*Ride, error)
	Delete(id int) error
	Update(id int, rideMap map[string]interface{}) (*Ride, error)
	UpdateStatus(id int, status common.RideStatus) (*Ride, error)
	AssignDriver(rideID int, driverID int, vehicleID int) (*Ride, error)
	StartRide(rideID int) (*Ride, error)
	CompleteRide(rideID int, actualDistance float64, actualDuration int, actualFare float64) (*Ride, error)
	CancelRide(rideID int, reason string) (*Ride, error)
	SearchPaginated(filters domain.DataFilters) (*SearchResultRide, error)
}
