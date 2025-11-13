package vehicle

import (
	"encoding/json"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	vehicleDomain "github.com/gbrayhan/microservices-go/src/domain/vehicle"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Vehicle represents the database model for vehicles
type Vehicle struct {
	ID              int        `gorm:"primaryKey"`
	DriverID        int        `gorm:"column:driver_id;index"`
	Make            string     `gorm:"column:make"`
	Model           string     `gorm:"column:model"`
	Year            int        `gorm:"column:year"`
	Color           string     `gorm:"column:color"`
	LicensePlate    string     `gorm:"column:license_plate;unique"`
	VehicleType     string     `gorm:"column:vehicle_type"`
	Capacity        int        `gorm:"column:capacity"`
	Status          string     `gorm:"column:status;default:'active'"`
	InsuranceNo     string     `gorm:"column:insurance_no"`
	InsuranceExp    *time.Time `gorm:"column:insurance_exp"`
	RegistrationNo  string     `gorm:"column:registration_no"`
	RegistrationExp *time.Time `gorm:"column:registration_exp"`
	CreatedAt       time.Time  `gorm:"autoCreateTime:mili"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime:mili"`
}

func (Vehicle) TableName() string {
	return "vehicles"
}

var ColumnsVehicleMapping = map[string]string{
	"id":              "id",
	"driverId":        "driver_id",
	"make":            "make",
	"model":           "model",
	"year":            "year",
	"color":           "color",
	"licensePlate":    "license_plate",
	"vehicleType":     "vehicle_type",
	"capacity":        "capacity",
	"status":          "status",
	"insuranceNo":     "insurance_no",
	"insuranceExp":    "insurance_exp",
	"registrationNo":  "registration_no",
	"registrationExp": "registration_exp",
	"createdAt":       "created_at",
	"updatedAt":       "updated_at",
}

// VehicleRepositoryInterface defines the interface for vehicle repository operations
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

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewVehicleRepository(db *gorm.DB, loggerInstance *logger.Logger) VehicleRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]vehicleDomain.Vehicle, error) {
	var vehicles []Vehicle
	if err := r.DB.Find(&vehicles).Error; err != nil {
		r.Logger.Error("Error getting all vehicles", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved all vehicles", zap.Int("count", len(vehicles)))
	return arrayToDomainMapper(&vehicles), nil
}

func (r *Repository) GetByID(id int) (*vehicleDomain.Vehicle, error) {
	var vehicle Vehicle
	if err := r.DB.First(&vehicle, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("Vehicle not found", zap.Int("id", id))
			return &vehicleDomain.Vehicle{}, nil
		}
		r.Logger.Error("Error getting vehicle by ID", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved vehicle", zap.Int("id", id))
	return vehicle.toDomainMapper(), nil
}

func (r *Repository) GetByDriverID(driverID int) (*[]vehicleDomain.Vehicle, error) {
	var vehicles []Vehicle
	if err := r.DB.Where("driver_id = ?", driverID).Find(&vehicles).Error; err != nil {
		r.Logger.Error("Error getting vehicles by driver ID", zap.Error(err), zap.Int("driverID", driverID))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved vehicles by driver",
		zap.Int("driverID", driverID),
		zap.Int("count", len(vehicles)))
	return arrayToDomainMapper(&vehicles), nil
}

func (r *Repository) Create(vehicleDomain *vehicleDomain.Vehicle) (*vehicleDomain.Vehicle, error) {
	r.Logger.Info("Creating new vehicle",
		zap.Int("driverID", vehicleDomain.DriverID),
		zap.String("licensePlate", vehicleDomain.LicensePlate))

	vehicleRepository := fromDomainMapper(vehicleDomain)
	if err := r.DB.Create(vehicleRepository).Error; err != nil {
		r.Logger.Error("Error creating vehicle", zap.Error(err))
		byteErr, _ := json.Marshal(err)
		var gormErr domainErrors.GormErr
		if unmarshalErr := json.Unmarshal(byteErr, &gormErr); unmarshalErr == nil {
			switch gormErr.Number {
			case 1062:
				return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			}
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	r.Logger.Info("Vehicle created successfully", zap.Int("id", vehicleRepository.ID))
	return vehicleRepository.toDomainMapper(), nil
}

func (r *Repository) Update(id int, vehicleMap map[string]interface{}) (*vehicleDomain.Vehicle, error) {
	r.Logger.Info("Updating vehicle", zap.Int("id", id))

	var vehicle Vehicle
	if err := r.DB.First(&vehicle, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("Vehicle not found for update", zap.Int("id", id))
			return &vehicleDomain.Vehicle{}, nil
		}
		r.Logger.Error("Error finding vehicle for update", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	if err := r.DB.Model(&vehicle).Updates(vehicleMap).Error; err != nil {
		r.Logger.Error("Error updating vehicle", zap.Error(err), zap.Int("id", id))
		byteErr, _ := json.Marshal(err)
		var gormErr domainErrors.GormErr
		if unmarshalErr := json.Unmarshal(byteErr, &gormErr); unmarshalErr == nil {
			switch gormErr.Number {
			case 1062:
				return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			}
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	// Reload to get updated data
	if err := r.DB.First(&vehicle, id).Error; err != nil {
		r.Logger.Error("Error reloading updated vehicle", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	r.Logger.Info("Vehicle updated successfully", zap.Int("id", id))
	return vehicle.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	r.Logger.Info("Deleting vehicle", zap.Int("id", id))

	result := r.DB.Delete(&Vehicle{}, id)
	if result.Error != nil {
		r.Logger.Error("Error deleting vehicle", zap.Error(result.Error), zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	if result.RowsAffected == 0 {
		r.Logger.Warn("Vehicle not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	r.Logger.Info("Vehicle deleted successfully", zap.Int("id", id))
	return nil
}

func (r *Repository) SearchPaginated(filters domain.DataFilters) (*vehicleDomain.SearchResultVehicle, error) {
	r.Logger.Info("Searching vehicles with pagination",
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))

	var vehicles []Vehicle
	var total int64

	query := r.DB.Model(&Vehicle{})

	// Apply like filters
	for field, values := range filters.LikeFilters {
		if len(values) > 0 {
			for _, value := range values {
				if value != "" {
					column := ColumnsVehicleMapping[field]
					if column != "" {
						query = query.Where(column+" ILIKE ?", "%"+value+"%")
					}
				}
			}
		}
	}

	// Apply exact matches
	for field, values := range filters.Matches {
		if len(values) > 0 {
			column := ColumnsVehicleMapping[field]
			if column != "" {
				query = query.Where(column+" IN ?", values)
			}
		}
	}

	// Apply date range filters
	for _, dateFilter := range filters.DateRangeFilters {
		column := ColumnsVehicleMapping[dateFilter.Field]
		if column != "" {
			if dateFilter.Start != nil {
				query = query.Where(column+" >= ?", dateFilter.Start)
			}
			if dateFilter.End != nil {
				query = query.Where(column+" <= ?", dateFilter.End)
			}
		}
	}

	// Apply sorting
	if len(filters.SortBy) > 0 && filters.SortDirection.IsValid() {
		for _, sortField := range filters.SortBy {
			column := ColumnsVehicleMapping[sortField]
			if column != "" {
				query = query.Order(column + " " + string(filters.SortDirection))
			}
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		r.Logger.Error("Error counting vehicles", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Offset(offset).Limit(filters.PageSize).Find(&vehicles).Error; err != nil {
		r.Logger.Error("Error fetching paginated vehicles", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize > 0 {
		totalPages++
	}

	result := &vehicleDomain.SearchResultVehicle{
		Data:       arrayToDomainMapper(&vehicles),
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}

	r.Logger.Info("Successfully retrieved paginated vehicles",
		zap.Int64("total", total),
		zap.Int("page", filters.Page),
		zap.Int("totalPages", totalPages))

	return result, nil
}

func (r *Repository) GetActiveVehiclesByType(vehicleType common.VehicleType) (*[]vehicleDomain.Vehicle, error) {
	var vehicles []Vehicle
	if err := r.DB.Where("vehicle_type = ? AND status = ?",
		string(vehicleType), string(common.VehicleStatusActive)).Find(&vehicles).Error; err != nil {
		r.Logger.Error("Error getting active vehicles by type",
			zap.Error(err),
			zap.String("type", string(vehicleType)))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved active vehicles by type",
		zap.String("type", string(vehicleType)),
		zap.Int("count", len(vehicles)))
	return arrayToDomainMapper(&vehicles), nil
}

// Mappers
func (v *Vehicle) toDomainMapper() *vehicleDomain.Vehicle {
	return &vehicleDomain.Vehicle{
		ID:              v.ID,
		DriverID:        v.DriverID,
		Make:            v.Make,
		Model:           v.Model,
		Year:            v.Year,
		Color:           v.Color,
		LicensePlate:    v.LicensePlate,
		VehicleType:     common.VehicleType(v.VehicleType),
		Capacity:        v.Capacity,
		Status:          common.VehicleStatus(v.Status),
		InsuranceNo:     v.InsuranceNo,
		InsuranceExp:    v.InsuranceExp,
		RegistrationNo:  v.RegistrationNo,
		RegistrationExp: v.RegistrationExp,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

func fromDomainMapper(v *vehicleDomain.Vehicle) *Vehicle {
	return &Vehicle{
		ID:              v.ID,
		DriverID:        v.DriverID,
		Make:            v.Make,
		Model:           v.Model,
		Year:            v.Year,
		Color:           v.Color,
		LicensePlate:    v.LicensePlate,
		VehicleType:     string(v.VehicleType),
		Capacity:        v.Capacity,
		Status:          string(v.Status),
		InsuranceNo:     v.InsuranceNo,
		InsuranceExp:    v.InsuranceExp,
		RegistrationNo:  v.RegistrationNo,
		RegistrationExp: v.RegistrationExp,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

func arrayToDomainMapper(vehicles *[]Vehicle) *[]vehicleDomain.Vehicle {
	vehiclesDomain := make([]vehicleDomain.Vehicle, len(*vehicles))
	for i, vehicle := range *vehicles {
		vehiclesDomain[i] = *vehicle.toDomainMapper()
	}
	return &vehiclesDomain
}
