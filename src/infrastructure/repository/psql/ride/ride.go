package ride

import (
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	rideDomain "github.com/gbrayhan/microservices-go/src/domain/ride"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Ride represents the database model for rides
type Ride struct {
	ID                 int        `gorm:"primaryKey"`
	RiderID            int        `gorm:"column:rider_id;index"`
	DriverID           *int       `gorm:"column:driver_id;index"`
	VehicleID          *int       `gorm:"column:vehicle_id"`
	Status             string     `gorm:"column:status;index"`
	PickupLatitude     float64    `gorm:"column:pickup_latitude"`
	PickupLongitude    float64    `gorm:"column:pickup_longitude"`
	PickupAddress      string     `gorm:"column:pickup_address"`
	DropoffLatitude    float64    `gorm:"column:dropoff_latitude"`
	DropoffLongitude   float64    `gorm:"column:dropoff_longitude"`
	DropoffAddress     string     `gorm:"column:dropoff_address"`
	RequestedAt        time.Time  `gorm:"column:requested_at;autoCreateTime"`
	MatchedAt          *time.Time `gorm:"column:matched_at"`
	StartedAt          *time.Time `gorm:"column:started_at"`
	CompletedAt        *time.Time `gorm:"column:completed_at"`
	CancelledAt        *time.Time `gorm:"column:cancelled_at"`
	CancellationReason string     `gorm:"column:cancellation_reason"`
	EstimatedDistance  float64    `gorm:"column:estimated_distance"`
	ActualDistance     float64    `gorm:"column:actual_distance"`
	EstimatedDuration  int        `gorm:"column:estimated_duration"`
	ActualDuration     int        `gorm:"column:actual_duration"`
	EstimatedFare      float64    `gorm:"column:estimated_fare"`
	ActualFare         float64    `gorm:"column:actual_fare"`
	RiderRating        *int       `gorm:"column:rider_rating"`
	DriverRating       *int       `gorm:"column:driver_rating"`
	Notes              string     `gorm:"column:notes"`
	CreatedAt          time.Time  `gorm:"autoCreateTime:mili"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime:mili"`
}

func (Ride) TableName() string {
	return "rides"
}

var ColumnsRideMapping = map[string]string{
	"id":                 "id",
	"riderId":            "rider_id",
	"driverId":           "driver_id",
	"vehicleId":          "vehicle_id",
	"status":             "status",
	"pickupLatitude":     "pickup_latitude",
	"pickupLongitude":    "pickup_longitude",
	"pickupAddress":      "pickup_address",
	"dropoffLatitude":    "dropoff_latitude",
	"dropoffLongitude":   "dropoff_longitude",
	"dropoffAddress":     "dropoff_address",
	"requestedAt":        "requested_at",
	"matchedAt":          "matched_at",
	"startedAt":          "started_at",
	"completedAt":        "completed_at",
	"cancelledAt":        "cancelled_at",
	"cancellationReason": "cancellation_reason",
	"estimatedDistance":  "estimated_distance",
	"actualDistance":     "actual_distance",
	"estimatedDuration":  "estimated_duration",
	"actualDuration":     "actual_duration",
	"estimatedFare":      "estimated_fare",
	"actualFare":         "actual_fare",
	"riderRating":        "rider_rating",
	"driverRating":       "driver_rating",
	"notes":              "notes",
	"createdAt":          "created_at",
	"updatedAt":          "updated_at",
}

// RideRepositoryInterface defines the interface for ride repository operations
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

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewRideRepository(db *gorm.DB, loggerInstance *logger.Logger) RideRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]rideDomain.Ride, error) {
	var rides []Ride
	if err := r.DB.Order("created_at DESC").Find(&rides).Error; err != nil {
		r.Logger.Error("Error getting all rides", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved all rides", zap.Int("count", len(rides)))
	return arrayToDomainMapper(&rides), nil
}

func (r *Repository) GetByID(id int) (*rideDomain.Ride, error) {
	var ride Ride
	if err := r.DB.First(&ride, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("Ride not found", zap.Int("id", id))
			return &rideDomain.Ride{}, nil
		}
		r.Logger.Error("Error getting ride by ID", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved ride", zap.Int("id", id))
	return ride.toDomainMapper(), nil
}

func (r *Repository) GetByRiderID(riderID int) (*[]rideDomain.Ride, error) {
	var rides []Ride
	if err := r.DB.Where("rider_id = ?", riderID).
		Order("created_at DESC").
		Find(&rides).Error; err != nil {
		r.Logger.Error("Error getting rides by rider ID", zap.Error(err), zap.Int("riderID", riderID))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved rides by rider",
		zap.Int("riderID", riderID),
		zap.Int("count", len(rides)))
	return arrayToDomainMapper(&rides), nil
}

func (r *Repository) GetByDriverID(driverID int) (*[]rideDomain.Ride, error) {
	var rides []Ride
	if err := r.DB.Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Find(&rides).Error; err != nil {
		r.Logger.Error("Error getting rides by driver ID", zap.Error(err), zap.Int("driverID", driverID))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved rides by driver",
		zap.Int("driverID", driverID),
		zap.Int("count", len(rides)))
	return arrayToDomainMapper(&rides), nil
}

func (r *Repository) GetPendingRides() (*[]rideDomain.Ride, error) {
	var rides []Ride
	if err := r.DB.Where("status = ?", string(common.RideStatusPending)).
		Order("created_at ASC").
		Find(&rides).Error; err != nil {
		r.Logger.Error("Error getting pending rides", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved pending rides", zap.Int("count", len(rides)))
	return arrayToDomainMapper(&rides), nil
}

func (r *Repository) Create(rideDomain *rideDomain.Ride) (*rideDomain.Ride, error) {
	r.Logger.Info("Creating new ride",
		zap.Int("riderID", rideDomain.RiderID),
		zap.String("status", string(rideDomain.Status)))

	rideRepository := fromDomainMapper(rideDomain)
	if err := r.DB.Create(rideRepository).Error; err != nil {
		r.Logger.Error("Error creating ride", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	r.Logger.Info("Ride created successfully", zap.Int("id", rideRepository.ID))
	return rideRepository.toDomainMapper(), nil
}

func (r *Repository) Update(id int, rideMap map[string]interface{}) (*rideDomain.Ride, error) {
	r.Logger.Info("Updating ride", zap.Int("id", id))

	var ride Ride
	if err := r.DB.First(&ride, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("Ride not found for update", zap.Int("id", id))
			return &rideDomain.Ride{}, nil
		}
		r.Logger.Error("Error finding ride for update", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	if err := r.DB.Model(&ride).Updates(rideMap).Error; err != nil {
		r.Logger.Error("Error updating ride", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	// Reload to get updated data
	if err := r.DB.First(&ride, id).Error; err != nil {
		r.Logger.Error("Error reloading updated ride", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	r.Logger.Info("Ride updated successfully", zap.Int("id", id))
	return ride.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	r.Logger.Info("Deleting ride", zap.Int("id", id))

	result := r.DB.Delete(&Ride{}, id)
	if result.Error != nil {
		r.Logger.Error("Error deleting ride", zap.Error(result.Error), zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	if result.RowsAffected == 0 {
		r.Logger.Warn("Ride not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	r.Logger.Info("Ride deleted successfully", zap.Int("id", id))
	return nil
}

func (r *Repository) SearchPaginated(filters domain.DataFilters) (*rideDomain.SearchResultRide, error) {
	r.Logger.Info("Searching rides with pagination",
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))

	var rides []Ride
	var total int64

	query := r.DB.Model(&Ride{})

	// Apply like filters
	for field, values := range filters.LikeFilters {
		if len(values) > 0 {
			for _, value := range values {
				if value != "" {
					column := ColumnsRideMapping[field]
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
			column := ColumnsRideMapping[field]
			if column != "" {
				query = query.Where(column+" IN ?", values)
			}
		}
	}

	// Apply date range filters
	for _, dateFilter := range filters.DateRangeFilters {
		column := ColumnsRideMapping[dateFilter.Field]
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
			column := ColumnsRideMapping[sortField]
			if column != "" {
				query = query.Order(column + " " + string(filters.SortDirection))
			}
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		r.Logger.Error("Error counting rides", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Offset(offset).Limit(filters.PageSize).
		Order("created_at DESC").
		Find(&rides).Error; err != nil {
		r.Logger.Error("Error fetching paginated rides", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize > 0 {
		totalPages++
	}

	result := &rideDomain.SearchResultRide{
		Data:       arrayToDomainMapper(&rides),
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}

	r.Logger.Info("Successfully retrieved paginated rides",
		zap.Int64("total", total),
		zap.Int("page", filters.Page),
		zap.Int("totalPages", totalPages))

	return result, nil
}

// Mappers
func (r *Ride) toDomainMapper() *rideDomain.Ride {
	return &rideDomain.Ride{
		ID:                 r.ID,
		RiderID:            r.RiderID,
		DriverID:           r.DriverID,
		VehicleID:          r.VehicleID,
		Status:             common.RideStatus(r.Status),
		PickupLatitude:     r.PickupLatitude,
		PickupLongitude:    r.PickupLongitude,
		PickupAddress:      r.PickupAddress,
		DropoffLatitude:    r.DropoffLatitude,
		DropoffLongitude:   r.DropoffLongitude,
		DropoffAddress:     r.DropoffAddress,
		RequestedAt:        r.RequestedAt,
		MatchedAt:          r.MatchedAt,
		StartedAt:          r.StartedAt,
		CompletedAt:        r.CompletedAt,
		CancelledAt:        r.CancelledAt,
		CancellationReason: r.CancellationReason,
		EstimatedDistance:  r.EstimatedDistance,
		ActualDistance:     r.ActualDistance,
		EstimatedDuration:  r.EstimatedDuration,
		ActualDuration:     r.ActualDuration,
		EstimatedFare:      r.EstimatedFare,
		ActualFare:         r.ActualFare,
		RiderRating:        r.RiderRating,
		DriverRating:       r.DriverRating,
		Notes:              r.Notes,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

func fromDomainMapper(r *rideDomain.Ride) *Ride {
	return &Ride{
		ID:                 r.ID,
		RiderID:            r.RiderID,
		DriverID:           r.DriverID,
		VehicleID:          r.VehicleID,
		Status:             string(r.Status),
		PickupLatitude:     r.PickupLatitude,
		PickupLongitude:    r.PickupLongitude,
		PickupAddress:      r.PickupAddress,
		DropoffLatitude:    r.DropoffLatitude,
		DropoffLongitude:   r.DropoffLongitude,
		DropoffAddress:     r.DropoffAddress,
		RequestedAt:        r.RequestedAt,
		MatchedAt:          r.MatchedAt,
		StartedAt:          r.StartedAt,
		CompletedAt:        r.CompletedAt,
		CancelledAt:        r.CancelledAt,
		CancellationReason: r.CancellationReason,
		EstimatedDistance:  r.EstimatedDistance,
		ActualDistance:     r.ActualDistance,
		EstimatedDuration:  r.EstimatedDuration,
		ActualDuration:     r.ActualDuration,
		EstimatedFare:      r.EstimatedFare,
		ActualFare:         r.ActualFare,
		RiderRating:        r.RiderRating,
		DriverRating:       r.DriverRating,
		Notes:              r.Notes,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

func arrayToDomainMapper(rides *[]Ride) *[]rideDomain.Ride {
	ridesDomain := make([]rideDomain.Ride, len(*rides))
	for i, ride := range *rides {
		ridesDomain[i] = *ride.toDomainMapper()
	}
	return &ridesDomain
}
