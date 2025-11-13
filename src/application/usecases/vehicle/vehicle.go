package vehicle

import (
	"fmt"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	vehicleDomain "github.com/gbrayhan/microservices-go/src/domain/vehicle"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

// VehicleRepositoryInterface defines the interface for vehicle repository
type VehicleRepositoryInterface interface {
	GetAll() (*[]vehicleDomain.Vehicle, error)
	GetByID(id int) (*vehicleDomain.Vehicle, error)
	GetByDriverID(driverID int) (*[]vehicleDomain.Vehicle, error)
	Create(vehicle *vehicleDomain.Vehicle) (*vehicleDomain.Vehicle, error)
	Update(id int, vehicleMap map[string]interface{}) (*vehicleDomain.Vehicle, error)
	Delete(id int) error
	SearchPaginated(filters domain.DataFilters) (*vehicleDomain.SearchResultVehicle, error)
	GetActiveVehiclesByType(vehicleType common.VehicleType) (*[]vehicleDomain.Vehicle, error)
}

// IVehicleUseCase defines the interface for vehicle use cases
type IVehicleUseCase interface {
	GetAll() (*[]vehicleDomain.Vehicle, error)
	GetByID(id int) (*vehicleDomain.Vehicle, error)
	GetByDriverID(driverID int) (*[]vehicleDomain.Vehicle, error)
	Create(newVehicle *vehicleDomain.Vehicle) (*vehicleDomain.Vehicle, error)
	Update(id int, vehicleMap map[string]interface{}) (*vehicleDomain.Vehicle, error)
	Delete(id int) error
	SearchPaginated(filters domain.DataFilters) (*vehicleDomain.SearchResultVehicle, error)
	GetActiveVehiclesByType(vehicleType common.VehicleType) (*[]vehicleDomain.Vehicle, error)
	ActivateVehicle(id int) (*vehicleDomain.Vehicle, error)
	DeactivateVehicle(id int) (*vehicleDomain.Vehicle, error)
	ValidateVehicleForRides(id int) error
}

// VehicleUseCase implements IVehicleUseCase
type VehicleUseCase struct {
	vehicleRepository VehicleRepositoryInterface
	Logger            *logger.Logger
}

// NewVehicleUseCase creates a new vehicle use case
func NewVehicleUseCase(vehicleRepository VehicleRepositoryInterface, logger *logger.Logger) IVehicleUseCase {
	return &VehicleUseCase{
		vehicleRepository: vehicleRepository,
		Logger:            logger,
	}
}

// GetAll retrieves all vehicles
func (s *VehicleUseCase) GetAll() (*[]vehicleDomain.Vehicle, error) {
	s.Logger.Info("Getting all vehicles")
	return s.vehicleRepository.GetAll()
}

// GetByID retrieves a vehicle by ID
func (s *VehicleUseCase) GetByID(id int) (*vehicleDomain.Vehicle, error) {
	s.Logger.Info("Getting vehicle by ID", zap.Int("id", id))
	vehicle, err := s.vehicleRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting vehicle", zap.Error(err), zap.Int("id", id))
		return nil, err
	}
	if vehicle.ID == 0 {
		s.Logger.Warn("Vehicle not found", zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return vehicle, nil
}

// GetByDriverID retrieves all vehicles for a specific driver
func (s *VehicleUseCase) GetByDriverID(driverID int) (*[]vehicleDomain.Vehicle, error) {
	s.Logger.Info("Getting vehicles by driver ID", zap.Int("driverID", driverID))
	return s.vehicleRepository.GetByDriverID(driverID)
}

// Create creates a new vehicle
func (s *VehicleUseCase) Create(newVehicle *vehicleDomain.Vehicle) (*vehicleDomain.Vehicle, error) {
	s.Logger.Info("Creating new vehicle",
		zap.Int("driverID", newVehicle.DriverID),
		zap.String("licensePlate", newVehicle.LicensePlate))

	// Validate vehicle type
	if !newVehicle.VehicleType.IsValid() {
		s.Logger.Warn("Invalid vehicle type", zap.String("type", string(newVehicle.VehicleType)))
		return nil, domainErrors.NewAppError(fmt.Errorf("invalid vehicle type: %s", newVehicle.VehicleType), domainErrors.ValidationError)
	}

	// Set default status if not provided
	if newVehicle.Status == "" {
		newVehicle.Status = common.VehicleStatusActive
	}

	// Validate status
	if !newVehicle.Status.IsValid() {
		s.Logger.Warn("Invalid vehicle status", zap.String("status", string(newVehicle.Status)))
		return nil, domainErrors.NewAppError(fmt.Errorf("invalid vehicle status: %s", newVehicle.Status), domainErrors.ValidationError)
	}

	vehicle, err := s.vehicleRepository.Create(newVehicle)
	if err != nil {
		s.Logger.Error("Error creating vehicle", zap.Error(err))
		return nil, err
	}

	s.Logger.Info("Vehicle created successfully", zap.Int("id", vehicle.ID))
	return vehicle, nil
}

// Update updates a vehicle
func (s *VehicleUseCase) Update(id int, vehicleMap map[string]interface{}) (*vehicleDomain.Vehicle, error) {
	s.Logger.Info("Updating vehicle", zap.Int("id", id))

	// Check if vehicle exists
	existingVehicle, err := s.vehicleRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting vehicle for update", zap.Error(err), zap.Int("id", id))
		return nil, err
	}
	if existingVehicle.ID == 0 {
		s.Logger.Warn("Vehicle not found for update", zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	// Validate vehicle type if being updated
	if vehicleType, ok := vehicleMap["vehicle_type"].(string); ok {
		vType := common.VehicleType(vehicleType)
		if !vType.IsValid() {
			s.Logger.Warn("Invalid vehicle type in update", zap.String("type", vehicleType))
			return nil, domainErrors.NewAppError(fmt.Errorf("invalid vehicle type: %s", vehicleType), domainErrors.ValidationError)
		}
	}

	// Validate status if being updated
	if status, ok := vehicleMap["status"].(string); ok {
		vStatus := common.VehicleStatus(status)
		if !vStatus.IsValid() {
			s.Logger.Warn("Invalid vehicle status in update", zap.String("status", status))
			return nil, domainErrors.NewAppError(fmt.Errorf("invalid vehicle status: %s", status), domainErrors.ValidationError)
		}
	}

	vehicle, err := s.vehicleRepository.Update(id, vehicleMap)
	if err != nil {
		s.Logger.Error("Error updating vehicle", zap.Error(err), zap.Int("id", id))
		return nil, err
	}

	s.Logger.Info("Vehicle updated successfully", zap.Int("id", id))
	return vehicle, nil
}

// Delete deletes a vehicle
func (s *VehicleUseCase) Delete(id int) error {
	s.Logger.Info("Deleting vehicle", zap.Int("id", id))

	// Check if vehicle exists
	existingVehicle, err := s.vehicleRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting vehicle for deletion", zap.Error(err), zap.Int("id", id))
		return err
	}
	if existingVehicle.ID == 0 {
		s.Logger.Warn("Vehicle not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = s.vehicleRepository.Delete(id)
	if err != nil {
		s.Logger.Error("Error deleting vehicle", zap.Error(err), zap.Int("id", id))
		return err
	}

	s.Logger.Info("Vehicle deleted successfully", zap.Int("id", id))
	return nil
}

// SearchPaginated performs paginated search
func (s *VehicleUseCase) SearchPaginated(filters domain.DataFilters) (*vehicleDomain.SearchResultVehicle, error) {
	s.Logger.Info("Searching vehicles with pagination",
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))
	return s.vehicleRepository.SearchPaginated(filters)
}

// GetActiveVehiclesByType retrieves active vehicles by type
func (s *VehicleUseCase) GetActiveVehiclesByType(vehicleType common.VehicleType) (*[]vehicleDomain.Vehicle, error) {
	s.Logger.Info("Getting active vehicles by type", zap.String("type", string(vehicleType)))

	if !vehicleType.IsValid() {
		s.Logger.Warn("Invalid vehicle type", zap.String("type", string(vehicleType)))
		return nil, domainErrors.NewAppError(fmt.Errorf("invalid vehicle type: %s", vehicleType), domainErrors.ValidationError)
	}

	return s.vehicleRepository.GetActiveVehiclesByType(vehicleType)
}

// ActivateVehicle activates a vehicle
func (s *VehicleUseCase) ActivateVehicle(id int) (*vehicleDomain.Vehicle, error) {
	s.Logger.Info("Activating vehicle", zap.Int("id", id))

	vehicle, err := s.vehicleRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if vehicle.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	// Check if vehicle can be activated (insurance and registration valid)
	if !vehicle.IsInsuranceValid() {
		s.Logger.Warn("Cannot activate vehicle: insurance expired", zap.Int("id", id))
		return nil, domainErrors.NewAppError(fmt.Errorf("cannot activate vehicle: insurance expired"), domainErrors.ValidationError)
	}

	if !vehicle.IsRegistrationValid() {
		s.Logger.Warn("Cannot activate vehicle: registration expired", zap.Int("id", id))
		return nil, domainErrors.NewAppError(fmt.Errorf("cannot activate vehicle: registration expired"), domainErrors.ValidationError)
	}

	vehicleMap := map[string]interface{}{
		"status": string(common.VehicleStatusActive),
	}

	return s.vehicleRepository.Update(id, vehicleMap)
}

// DeactivateVehicle deactivates a vehicle
func (s *VehicleUseCase) DeactivateVehicle(id int) (*vehicleDomain.Vehicle, error) {
	s.Logger.Info("Deactivating vehicle", zap.Int("id", id))

	vehicle, err := s.vehicleRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if vehicle.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	vehicleMap := map[string]interface{}{
		"status": string(common.VehicleStatusInactive),
	}

	return s.vehicleRepository.Update(id, vehicleMap)
}

// ValidateVehicleForRides validates if a vehicle can accept rides
func (s *VehicleUseCase) ValidateVehicleForRides(id int) error {
	s.Logger.Info("Validating vehicle for rides", zap.Int("id", id))

	vehicle, err := s.vehicleRepository.GetByID(id)
	if err != nil {
		return err
	}
	if vehicle.ID == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	if !vehicle.CanAcceptRides() {
		s.Logger.Warn("Vehicle cannot accept rides",
			zap.Int("id", id),
			zap.Bool("isActive", vehicle.IsActive()),
			zap.Bool("insuranceValid", vehicle.IsInsuranceValid()),
			zap.Bool("registrationValid", vehicle.IsRegistrationValid()))

		return domainErrors.NewAppError(
			fmt.Errorf("vehicle cannot accept rides: check status, insurance, and registration"),
			domainErrors.ValidationError)
	}

	return nil
}
