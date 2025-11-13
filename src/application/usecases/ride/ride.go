package ride

import (
	"fmt"
	"math"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	rideDomain "github.com/gbrayhan/microservices-go/src/domain/ride"
	userDomain "github.com/gbrayhan/microservices-go/src/domain/user"
	vehicleDomain "github.com/gbrayhan/microservices-go/src/domain/vehicle"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

// RideRepositoryInterface defines the interface for ride repository
type RideRepositoryInterface interface {
	GetAll() (*[]rideDomain.Ride, error)
	GetByID(id int) (*rideDomain.Ride, error)
	GetByRiderID(riderID int) (*[]rideDomain.Ride, error)
	GetByDriverID(driverID int) (*[]rideDomain.Ride, error)
	GetPendingRides() (*[]rideDomain.Ride, error)
	Create(ride *rideDomain.Ride) (*rideDomain.Ride, error)
	Update(id int, rideMap map[string]interface{}) (*rideDomain.Ride, error)
	Delete(id int) error
	SearchPaginated(filters domain.DataFilters) (*rideDomain.SearchResultRide, error)
}

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	GetByID(id int) (*userDomain.User, error)
	GetAvailableDriversNearby(latitude, longitude float64, radius float64) (*[]userDomain.User, error)
	Update(id int, userMap map[string]interface{}) (*userDomain.User, error)
}

// VehicleRepositoryInterface defines the interface for vehicle repository
type VehicleRepositoryInterface interface {
	GetByDriverID(driverID int) (*[]vehicleDomain.Vehicle, error)
	GetByID(id int) (*vehicleDomain.Vehicle, error)
}

// IRideUseCase defines the interface for ride use cases
type IRideUseCase interface {
	GetAll() (*[]rideDomain.Ride, error)
	GetByID(id int) (*rideDomain.Ride, error)
	GetByRiderID(riderID int) (*[]rideDomain.Ride, error)
	GetByDriverID(driverID int) (*[]rideDomain.Ride, error)
	GetPendingRides() (*[]rideDomain.Ride, error)
	Create(newRide *rideDomain.Ride) (*rideDomain.Ride, error)
	Update(id int, rideMap map[string]interface{}) (*rideDomain.Ride, error)
	Delete(id int) error
	UpdateStatus(id int, status common.RideStatus) (*rideDomain.Ride, error)
	AssignDriver(rideID int, driverID int, vehicleID int) (*rideDomain.Ride, error)
	StartRide(rideID int) (*rideDomain.Ride, error)
	CompleteRide(rideID int, actualDistance float64, actualDuration int, actualFare float64) (*rideDomain.Ride, error)
	CancelRide(rideID int, reason string) (*rideDomain.Ride, error)
	SearchPaginated(filters domain.DataFilters) (*rideDomain.SearchResultRide, error)
	MatchDriverAutomatically(rideID int) (*rideDomain.Ride, error)
	RateDriver(rideID int, rating int) (*rideDomain.Ride, error)
	RateRider(rideID int, rating int) (*rideDomain.Ride, error)
}

// RideUseCase implements IRideUseCase
type RideUseCase struct {
	rideRepository    RideRepositoryInterface
	userRepository    UserRepositoryInterface
	vehicleRepository VehicleRepositoryInterface
	Logger            *logger.Logger
}

// NewRideUseCase creates a new ride use case
func NewRideUseCase(
	rideRepository RideRepositoryInterface,
	userRepository UserRepositoryInterface,
	vehicleRepository VehicleRepositoryInterface,
	logger *logger.Logger,
) IRideUseCase {
	return &RideUseCase{
		rideRepository:    rideRepository,
		userRepository:    userRepository,
		vehicleRepository: vehicleRepository,
		Logger:            logger,
	}
}

// GetAll retrieves all rides
func (s *RideUseCase) GetAll() (*[]rideDomain.Ride, error) {
	s.Logger.Info("Getting all rides")
	return s.rideRepository.GetAll()
}

// GetByID retrieves a ride by ID
func (s *RideUseCase) GetByID(id int) (*rideDomain.Ride, error) {
	s.Logger.Info("Getting ride by ID", zap.Int("id", id))
	ride, err := s.rideRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting ride", zap.Error(err), zap.Int("id", id))
		return nil, err
	}
	if ride.ID == 0 {
		s.Logger.Warn("Ride not found", zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return ride, nil
}

// GetByRiderID retrieves all rides for a rider
func (s *RideUseCase) GetByRiderID(riderID int) (*[]rideDomain.Ride, error) {
	s.Logger.Info("Getting rides by rider ID", zap.Int("riderID", riderID))
	return s.rideRepository.GetByRiderID(riderID)
}

// GetByDriverID retrieves all rides for a driver
func (s *RideUseCase) GetByDriverID(driverID int) (*[]rideDomain.Ride, error) {
	s.Logger.Info("Getting rides by driver ID", zap.Int("driverID", driverID))
	return s.rideRepository.GetByDriverID(driverID)
}

// GetPendingRides retrieves all pending rides
func (s *RideUseCase) GetPendingRides() (*[]rideDomain.Ride, error) {
	s.Logger.Info("Getting pending rides")
	return s.rideRepository.GetPendingRides()
}

// Create creates a new ride request
func (s *RideUseCase) Create(newRide *rideDomain.Ride) (*rideDomain.Ride, error) {
	s.Logger.Info("Creating new ride",
		zap.Int("riderID", newRide.RiderID),
		zap.Float64("pickupLat", newRide.PickupLatitude),
		zap.Float64("pickupLng", newRide.PickupLongitude))

	// Validate rider exists
	rider, err := s.userRepository.GetByID(newRide.RiderID)
	if err != nil {
		return nil, err
	}
	if rider.ID == 0 {
		return nil, domainErrors.NewAppError(fmt.Errorf("rider not found"), domainErrors.NotFound)
	}

	// Set initial status
	newRide.Status = common.RideStatusPending

	// Calculate estimated distance (simple Haversine formula)
	newRide.EstimatedDistance = calculateDistance(
		newRide.PickupLatitude, newRide.PickupLongitude,
		newRide.DropoffLatitude, newRide.DropoffLongitude,
	)

	// Calculate estimated duration (assuming average speed of 40 km/h)
	newRide.EstimatedDuration = int((newRide.EstimatedDistance / 40.0) * 60)

	// Calculate estimated fare (base fare + per km rate)
	baseFare := 2.50
	perKmRate := 1.20
	newRide.EstimatedFare = baseFare + (newRide.EstimatedDistance * perKmRate)

	ride, err := s.rideRepository.Create(newRide)
	if err != nil {
		s.Logger.Error("Error creating ride", zap.Error(err))
		return nil, err
	}

	s.Logger.Info("Ride created successfully", zap.Int("id", ride.ID))
	return ride, nil
}

// Update updates a ride
func (s *RideUseCase) Update(id int, rideMap map[string]interface{}) (*rideDomain.Ride, error) {
	s.Logger.Info("Updating ride", zap.Int("id", id))

	// Check if ride exists
	existingRide, err := s.rideRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting ride for update", zap.Error(err), zap.Int("id", id))
		return nil, err
	}
	if existingRide.ID == 0 {
		s.Logger.Warn("Ride not found for update", zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	ride, err := s.rideRepository.Update(id, rideMap)
	if err != nil {
		s.Logger.Error("Error updating ride", zap.Error(err), zap.Int("id", id))
		return nil, err
	}

	s.Logger.Info("Ride updated successfully", zap.Int("id", id))
	return ride, nil
}

// Delete deletes a ride
func (s *RideUseCase) Delete(id int) error {
	s.Logger.Info("Deleting ride", zap.Int("id", id))

	// Check if ride exists
	existingRide, err := s.rideRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting ride for deletion", zap.Error(err), zap.Int("id", id))
		return err
	}
	if existingRide.ID == 0 {
		s.Logger.Warn("Ride not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = s.rideRepository.Delete(id)
	if err != nil {
		s.Logger.Error("Error deleting ride", zap.Error(err), zap.Int("id", id))
		return err
	}

	s.Logger.Info("Ride deleted successfully", zap.Int("id", id))
	return nil
}

// UpdateStatus updates the status of a ride
func (s *RideUseCase) UpdateStatus(id int, status common.RideStatus) (*rideDomain.Ride, error) {
	s.Logger.Info("Updating ride status", zap.Int("id", id), zap.String("status", string(status)))

	ride, err := s.rideRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	// Validate status transition
	if !ride.Status.CanTransitionTo(status) {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("invalid status transition from %s to %s", ride.Status, status),
			domainErrors.ValidationError)
	}

	rideMap := map[string]interface{}{
		"status": string(status),
	}

	return s.rideRepository.Update(id, rideMap)
}

// AssignDriver assigns a driver to a ride
func (s *RideUseCase) AssignDriver(rideID int, driverID int, vehicleID int) (*rideDomain.Ride, error) {
	s.Logger.Info("Assigning driver to ride",
		zap.Int("rideID", rideID),
		zap.Int("driverID", driverID),
		zap.Int("vehicleID", vehicleID))

	ride, err := s.rideRepository.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	// Validate driver
	driver, err := s.userRepository.GetByID(driverID)
	if err != nil {
		return nil, err
	}
	if driver.ID == 0 || !driver.IsDriver() {
		return nil, domainErrors.NewAppError(fmt.Errorf("invalid driver"), domainErrors.ValidationError)
	}

	if !driver.CanAcceptRides() {
		return nil, domainErrors.NewAppError(fmt.Errorf("driver not available"), domainErrors.ValidationError)
	}

	// Validate vehicle
	vehicle, err := s.vehicleRepository.GetByID(vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle.ID == 0 || vehicle.DriverID != driverID {
		return nil, domainErrors.NewAppError(fmt.Errorf("invalid vehicle for driver"), domainErrors.ValidationError)
	}

	if !vehicle.CanAcceptRides() {
		return nil, domainErrors.NewAppError(fmt.Errorf("vehicle not ready"), domainErrors.ValidationError)
	}

	// Use domain logic to assign driver
	err = ride.AssignDriver(driverID, vehicleID)
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	rideMap := map[string]interface{}{
		"driver_id":  driverID,
		"vehicle_id": vehicleID,
		"status":     string(common.RideStatusMatched),
		"matched_at": ride.MatchedAt,
	}

	// Update driver availability
	_, err = s.userRepository.Update(driverID, map[string]interface{}{
		"is_available": false,
	})
	if err != nil {
		s.Logger.Error("Error updating driver availability", zap.Error(err))
	}

	return s.rideRepository.Update(rideID, rideMap)
}

// StartRide starts a matched ride
func (s *RideUseCase) StartRide(rideID int) (*rideDomain.Ride, error) {
	s.Logger.Info("Starting ride", zap.Int("rideID", rideID))

	ride, err := s.rideRepository.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = ride.Start()
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	rideMap := map[string]interface{}{
		"status":     string(common.RideStatusInProgress),
		"started_at": ride.StartedAt,
	}

	return s.rideRepository.Update(rideID, rideMap)
}

// CompleteRide completes an in-progress ride
func (s *RideUseCase) CompleteRide(rideID int, actualDistance float64, actualDuration int, actualFare float64) (*rideDomain.Ride, error) {
	s.Logger.Info("Completing ride",
		zap.Int("rideID", rideID),
		zap.Float64("distance", actualDistance),
		zap.Float64("fare", actualFare))

	ride, err := s.rideRepository.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = ride.Complete(actualDistance, actualDuration, actualFare)
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	rideMap := map[string]interface{}{
		"status":          string(common.RideStatusCompleted),
		"completed_at":    ride.CompletedAt,
		"actual_distance": actualDistance,
		"actual_duration": actualDuration,
		"actual_fare":     actualFare,
	}

	// Update driver availability
	if ride.DriverID != nil {
		_, err = s.userRepository.Update(*ride.DriverID, map[string]interface{}{
			"is_available": true,
			"total_rides":  fmt.Sprintf("total_rides + 1"), // This will be handled by repository
		})
		if err != nil {
			s.Logger.Error("Error updating driver after ride completion", zap.Error(err))
		}
	}

	return s.rideRepository.Update(rideID, rideMap)
}

// CancelRide cancels a ride
func (s *RideUseCase) CancelRide(rideID int, reason string) (*rideDomain.Ride, error) {
	s.Logger.Info("Cancelling ride", zap.Int("rideID", rideID), zap.String("reason", reason))

	ride, err := s.rideRepository.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = ride.Cancel(reason)
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	rideMap := map[string]interface{}{
		"status":              string(common.RideStatusCancelled),
		"cancelled_at":        ride.CancelledAt,
		"cancellation_reason": reason,
	}

	// If driver was assigned, make them available again
	if ride.DriverID != nil {
		_, err = s.userRepository.Update(*ride.DriverID, map[string]interface{}{
			"is_available": true,
		})
		if err != nil {
			s.Logger.Error("Error updating driver availability after cancellation", zap.Error(err))
		}
	}

	return s.rideRepository.Update(rideID, rideMap)
}

// SearchPaginated performs paginated search
func (s *RideUseCase) SearchPaginated(filters domain.DataFilters) (*rideDomain.SearchResultRide, error) {
	s.Logger.Info("Searching rides with pagination",
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))
	return s.rideRepository.SearchPaginated(filters)
}

// MatchDriverAutomatically finds and assigns the best driver automatically
func (s *RideUseCase) MatchDriverAutomatically(rideID int) (*rideDomain.Ride, error) {
	s.Logger.Info("Matching driver automatically", zap.Int("rideID", rideID))

	ride, err := s.rideRepository.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	if !ride.IsPending() {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("ride is not in pending status"),
			domainErrors.ValidationError)
	}

	// Find available drivers nearby (within 5km radius)
	drivers, err := s.userRepository.GetAvailableDriversNearby(
		ride.PickupLatitude,
		ride.PickupLongitude,
		5.0,
	)
	if err != nil {
		return nil, err
	}

	if drivers == nil || len(*drivers) == 0 {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("no available drivers found"),
			domainErrors.NotFound)
	}

	// Find the closest driver with highest rating
	var bestDriver *userDomain.User
	var bestVehicle *vehicleDomain.Vehicle
	var bestScore float64 = -1

	for _, driver := range *drivers {
		vehicles, err := s.vehicleRepository.GetByDriverID(driver.ID)
		if err != nil || vehicles == nil || len(*vehicles) == 0 {
			continue
		}

		// Find active vehicle
		var activeVehicle *vehicleDomain.Vehicle
		for _, v := range *vehicles {
			if v.CanAcceptRides() {
				activeVehicle = &v
				break
			}
		}

		if activeVehicle == nil {
			continue
		}

		// Calculate score based on distance and rating
		distance := calculateDistance(
			ride.PickupLatitude, ride.PickupLongitude,
			driver.Latitude, driver.Longitude,
		)

		// Score: 70% weight on proximity, 30% on rating
		// Normalize distance (closer = higher score)
		distanceScore := math.Max(0, 1.0-(distance/5.0))
		ratingScore := driver.Rating / 5.0

		score := (distanceScore * 0.7) + (ratingScore * 0.3)

		if score > bestScore {
			bestScore = score
			bestDriver = &driver
			bestVehicle = activeVehicle
		}
	}

	if bestDriver == nil || bestVehicle == nil {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("no suitable driver found"),
			domainErrors.NotFound)
	}

	s.Logger.Info("Found best driver",
		zap.Int("driverID", bestDriver.ID),
		zap.Float64("score", bestScore))

	return s.AssignDriver(rideID, bestDriver.ID, bestVehicle.ID)
}

// RateDriver allows rider to rate the driver
func (s *RideUseCase) RateDriver(rideID int, rating int) (*rideDomain.Ride, error) {
	s.Logger.Info("Rating driver", zap.Int("rideID", rideID), zap.Int("rating", rating))

	ride, err := s.rideRepository.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = ride.SetRiderRating(rating)
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	rideMap := map[string]interface{}{
		"rider_rating": rating,
	}

	// Update driver's rating
	if ride.DriverID != nil {
		driver, err := s.userRepository.GetByID(*ride.DriverID)
		if err == nil && driver.ID != 0 {
			driver.UpdateRating(float64(rating), driver.TotalRides)
			_, err = s.userRepository.Update(*ride.DriverID, map[string]interface{}{
				"rating": driver.Rating,
			})
			if err != nil {
				s.Logger.Error("Error updating driver rating", zap.Error(err))
			}
		}
	}

	return s.rideRepository.Update(rideID, rideMap)
}

// RateRider allows driver to rate the rider
func (s *RideUseCase) RateRider(rideID int, rating int) (*rideDomain.Ride, error) {
	s.Logger.Info("Rating rider", zap.Int("rideID", rideID), zap.Int("rating", rating))

	ride, err := s.rideRepository.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = ride.SetDriverRating(rating)
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	rideMap := map[string]interface{}{
		"driver_rating": rating,
	}

	return s.rideRepository.Update(rideID, rideMap)
}

// calculateDistance calculates the distance between two points using Haversine formula
// Returns distance in kilometers
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
