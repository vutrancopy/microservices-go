package vehicle

import (
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
)

// Vehicle represents a driver's vehicle in the system
type Vehicle struct {
	ID            int
	DriverID      int    // Foreign key to User (driver)
	Make          string // e.g., Toyota, Honda
	Model         string // e.g., Camry, Civic
	Year          int
	Color         string
	LicensePlate  string
	VehicleType   common.VehicleType
	Capacity      int // Number of passengers
	Status        common.VehicleStatus
	InsuranceNo   string
	InsuranceExp  *time.Time
	RegistrationNo string
	RegistrationExp *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IsActive checks if the vehicle is active and ready for use
func (v *Vehicle) IsActive() bool {
	return v.Status == common.VehicleStatusActive
}

// IsInsuranceValid checks if the insurance is still valid
func (v *Vehicle) IsInsuranceValid() bool {
	if v.InsuranceExp == nil {
		return false
	}
	return v.InsuranceExp.After(time.Now())
}

// IsRegistrationValid checks if the registration is still valid
func (v *Vehicle) IsRegistrationValid() bool {
	if v.RegistrationExp == nil {
		return false
	}
	return v.RegistrationExp.After(time.Now())
}

// CanAcceptRides checks if the vehicle can accept rides
func (v *Vehicle) CanAcceptRides() bool {
	return v.IsActive() && v.IsInsuranceValid() && v.IsRegistrationValid()
}

// SearchResultVehicle represents paginated vehicle search results
type SearchResultVehicle struct {
	Data       *[]Vehicle
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// IVehicleService defines the interface for vehicle operations
type IVehicleService interface {
	GetAll() (*[]Vehicle, error)
	GetByID(id int) (*Vehicle, error)
	GetByDriverID(driverID int) (*[]Vehicle, error)
	Create(newVehicle *Vehicle) (*Vehicle, error)
	Delete(id int) error
	Update(id int, vehicleMap map[string]interface{}) (*Vehicle, error)
	SearchPaginated(filters domain.DataFilters) (*SearchResultVehicle, error)
	GetActiveVehiclesByType(vehicleType common.VehicleType) (*[]Vehicle, error)
}
