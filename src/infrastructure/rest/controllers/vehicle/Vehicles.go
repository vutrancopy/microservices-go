package vehicle

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gbrayhan/microservices-go/src/application/usecases/vehicle"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	vehicleDomain "github.com/gbrayhan/microservices-go/src/domain/vehicle"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Request/Response Structures
type NewVehicleRequest struct {
	DriverID        int     `json:"driverId" binding:"required"`
	Make            string  `json:"make" binding:"required"`
	Model           string  `json:"model" binding:"required"`
	Year            int     `json:"year" binding:"required,min=1900,max=2100"`
	Color           string  `json:"color" binding:"required"`
	LicensePlate    string  `json:"licensePlate" binding:"required"`
	VehicleType     string  `json:"vehicleType" binding:"required"`
	Capacity        int     `json:"capacity" binding:"required,min=1,max=50"`
	InsuranceNo     string  `json:"insuranceNo" binding:"required"`
	InsuranceExp    *string `json:"insuranceExp"`
	RegistrationNo  string  `json:"registrationNo" binding:"required"`
	RegistrationExp *string `json:"registrationExp"`
}

type UpdateVehicleRequest struct {
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	Year            int     `json:"year" binding:"omitempty,min=1900,max=2100"`
	Color           string  `json:"color"`
	VehicleType     string  `json:"vehicleType"`
	Capacity        int     `json:"capacity" binding:"omitempty,min=1,max=50"`
	Status          string  `json:"status"`
	InsuranceNo     string  `json:"insuranceNo"`
	InsuranceExp    *string `json:"insuranceExp"`
	RegistrationNo  string  `json:"registrationNo"`
	RegistrationExp *string `json:"registrationExp"`
}

type ResponseVehicle struct {
	ID              int        `json:"id"`
	DriverID        int        `json:"driverId"`
	Make            string     `json:"make"`
	Model           string     `json:"model"`
	Year            int        `json:"year"`
	Color           string     `json:"color"`
	LicensePlate    string     `json:"licensePlate"`
	VehicleType     string     `json:"vehicleType"`
	Capacity        int        `json:"capacity"`
	Status          string     `json:"status"`
	InsuranceNo     string     `json:"insuranceNo"`
	InsuranceExp    *time.Time `json:"insuranceExp,omitempty"`
	RegistrationNo  string     `json:"registrationNo"`
	RegistrationExp *time.Time `json:"registrationExp,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type IVehicleController interface {
	CreateVehicle(ctx *gin.Context)
	GetAllVehicles(ctx *gin.Context)
	GetVehicleByID(ctx *gin.Context)
	GetVehiclesByDriver(ctx *gin.Context)
	UpdateVehicle(ctx *gin.Context)
	DeleteVehicle(ctx *gin.Context)
	ActivateVehicle(ctx *gin.Context)
	DeactivateVehicle(ctx *gin.Context)
	SearchVehicles(ctx *gin.Context)
}

type VehicleController struct {
	vehicleUseCase vehicle.IVehicleUseCase
	Logger         *logger.Logger
}

func NewVehicleController(vehicleUseCase vehicle.IVehicleUseCase, loggerInstance *logger.Logger) IVehicleController {
	return &VehicleController{
		vehicleUseCase: vehicleUseCase,
		Logger:         loggerInstance,
	}
}

// CreateVehicle godoc
// @Summary Create a new vehicle
// @Description Create a new vehicle for a driver
// @Tags vehicles
// @Accept json
// @Produce json
// @Param vehicle body NewVehicleRequest true "Vehicle data"
// @Success 201 {object} ResponseVehicle
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle [post]
func (c *VehicleController) CreateVehicle(ctx *gin.Context) {
	c.Logger.Info("Creating new vehicle")

	var request NewVehicleRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	// Validate vehicle type
	vehicleType := common.VehicleType(request.VehicleType)
	if !vehicleType.IsValid() {
		appError := domainErrors.NewAppError(domainErrors.NewAppErrorWithType(domainErrors.ValidationError), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	// Parse dates if provided
	var insuranceExp, registrationExp *time.Time
	if request.InsuranceExp != nil {
		if parsed, err := time.Parse(time.RFC3339, *request.InsuranceExp); err == nil {
			insuranceExp = &parsed
		}
	}
	if request.RegistrationExp != nil {
		if parsed, err := time.Parse(time.RFC3339, *request.RegistrationExp); err == nil {
			registrationExp = &parsed
		}
	}

	newVehicle := &vehicleDomain.Vehicle{
		DriverID:        request.DriverID,
		Make:            request.Make,
		Model:           request.Model,
		Year:            request.Year,
		Color:           request.Color,
		LicensePlate:    request.LicensePlate,
		VehicleType:     vehicleType,
		Capacity:        request.Capacity,
		Status:          common.VehicleStatusActive,
		InsuranceNo:     request.InsuranceNo,
		InsuranceExp:    insuranceExp,
		RegistrationNo:  request.RegistrationNo,
		RegistrationExp: registrationExp,
	}

	createdVehicle, err := c.vehicleUseCase.Create(newVehicle)
	if err != nil {
		c.Logger.Error("Error creating vehicle", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	c.Logger.Info("Vehicle created successfully", zap.Int("id", createdVehicle.ID))
	ctx.JSON(http.StatusCreated, mapVehicleToResponse(createdVehicle))
}

// GetAllVehicles godoc
// @Summary Get all vehicles
// @Description Get all vehicles in the system
// @Tags vehicles
// @Produce json
// @Success 200 {array} ResponseVehicle
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle [get]
func (c *VehicleController) GetAllVehicles(ctx *gin.Context) {
	c.Logger.Info("Getting all vehicles")

	vehicles, err := c.vehicleUseCase.GetAll()
	if err != nil {
		c.Logger.Error("Error getting vehicles", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponseVehicle, 0, len(*vehicles))
	for _, v := range *vehicles {
		response = append(response, *mapVehicleToResponse(&v))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetVehicleByID godoc
// @Summary Get vehicle by ID
// @Description Get a specific vehicle by ID
// @Tags vehicles
// @Produce json
// @Param id path int true "Vehicle ID"
// @Success 200 {object} ResponseVehicle
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle/{id} [get]
func (c *VehicleController) GetVehicleByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid vehicle ID", zap.Error(err))
		appError := domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	c.Logger.Info("Getting vehicle by ID", zap.Int("id", id))

	vehicle, err := c.vehicleUseCase.GetByID(id)
	if err != nil {
		c.Logger.Error("Error getting vehicle", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	if vehicle.ID == 0 {
		appError := domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		_ = ctx.Error(appError)
		return
	}

	ctx.JSON(http.StatusOK, mapVehicleToResponse(vehicle))
}

// GetVehiclesByDriver godoc
// @Summary Get vehicles by driver ID
// @Description Get all vehicles for a specific driver
// @Tags vehicles
// @Produce json
// @Param driverId path int true "Driver ID"
// @Success 200 {array} ResponseVehicle
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle/driver/{driverId} [get]
func (c *VehicleController) GetVehiclesByDriver(ctx *gin.Context) {
	driverID, err := strconv.Atoi(ctx.Param("driverId"))
	if err != nil {
		c.Logger.Warn("Invalid driver ID", zap.Error(err))
		appError := domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}

	c.Logger.Info("Getting vehicles by driver", zap.Int("driverId", driverID))

	vehicles, err := c.vehicleUseCase.GetByDriverID(driverID)
	if err != nil {
		c.Logger.Error("Error getting vehicles by driver", zap.Error(err), zap.Int("driverId", driverID))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponseVehicle, 0, len(*vehicles))
	for _, v := range *vehicles {
		response = append(response, *mapVehicleToResponse(&v))
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateVehicle godoc
// @Summary Update vehicle
// @Description Update a vehicle's information
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path int true "Vehicle ID"
// @Param vehicle body UpdateVehicleRequest true "Vehicle data"
// @Success 200 {object} ResponseVehicle
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle/{id} [put]
func (c *VehicleController) UpdateVehicle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid vehicle ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request UpdateVehicleRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Updating vehicle", zap.Int("id", id))

	updateMap := make(map[string]interface{})
	if request.Make != "" {
		updateMap["make"] = request.Make
	}
	if request.Model != "" {
		updateMap["model"] = request.Model
	}
	if request.Year > 0 {
		updateMap["year"] = request.Year
	}
	if request.Color != "" {
		updateMap["color"] = request.Color
	}
	if request.VehicleType != "" {
		vehicleType := common.VehicleType(request.VehicleType)
		if !vehicleType.IsValid() {
			_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
			return
		}
		updateMap["vehicle_type"] = request.VehicleType
	}
	if request.Capacity > 0 {
		updateMap["capacity"] = request.Capacity
	}
	if request.Status != "" {
		status := common.VehicleStatus(request.Status)
		if !status.IsValid() {
			_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
			return
		}
		updateMap["status"] = request.Status
	}
	if request.InsuranceNo != "" {
		updateMap["insurance_no"] = request.InsuranceNo
	}
	if request.InsuranceExp != nil {
		if parsed, err := time.Parse(time.RFC3339, *request.InsuranceExp); err == nil {
			updateMap["insurance_exp"] = parsed
		}
	}
	if request.RegistrationNo != "" {
		updateMap["registration_no"] = request.RegistrationNo
	}
	if request.RegistrationExp != nil {
		if parsed, err := time.Parse(time.RFC3339, *request.RegistrationExp); err == nil {
			updateMap["registration_exp"] = parsed
		}
	}

	updatedVehicle, err := c.vehicleUseCase.Update(id, updateMap)
	if err != nil {
		c.Logger.Error("Error updating vehicle", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	if updatedVehicle.ID == 0 {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotFound))
		return
	}

	ctx.JSON(http.StatusOK, mapVehicleToResponse(updatedVehicle))
}

// DeleteVehicle godoc
// @Summary Delete vehicle
// @Description Delete a vehicle
// @Tags vehicles
// @Param id path int true "Vehicle ID"
// @Success 200 {object} controllers.MessageResponse
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle/{id} [delete]
func (c *VehicleController) DeleteVehicle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid vehicle ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Deleting vehicle", zap.Int("id", id))

	err = c.vehicleUseCase.Delete(id)
	if err != nil {
		c.Logger.Error("Error deleting vehicle", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Vehicle deleted successfully"})
}

// ActivateVehicle godoc
// @Summary Activate vehicle
// @Description Activate a vehicle for accepting rides
// @Tags vehicles
// @Param id path int true "Vehicle ID"
// @Success 200 {object} ResponseVehicle
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle/{id}/activate [put]
func (c *VehicleController) ActivateVehicle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid vehicle ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Activating vehicle", zap.Int("id", id))

	vehicle, err := c.vehicleUseCase.ActivateVehicle(id)
	if err != nil {
		c.Logger.Error("Error activating vehicle", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapVehicleToResponse(vehicle))
}

// DeactivateVehicle godoc
// @Summary Deactivate vehicle
// @Description Deactivate a vehicle from accepting rides
// @Tags vehicles
// @Param id path int true "Vehicle ID"
// @Success 200 {object} ResponseVehicle
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle/{id}/deactivate [put]
func (c *VehicleController) DeactivateVehicle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid vehicle ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Deactivating vehicle", zap.Int("id", id))

	vehicle, err := c.vehicleUseCase.DeactivateVehicle(id)
	if err != nil {
		c.Logger.Error("Error deactivating vehicle", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapVehicleToResponse(vehicle))
}

// SearchVehicles godoc
// @Summary Search vehicles
// @Description Search vehicles with pagination
// @Tags vehicles
// @Produce json
// @Param filters body domain.DataFilters true "Search filters"
// @Success 200 {object} vehicle.SearchResultVehicle
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/vehicle/search [post]
func (c *VehicleController) SearchVehicles(ctx *gin.Context) {
	filters, err := controllers.GetSearchFilters(ctx)
	if err != nil {
		c.Logger.Error("Invalid search filters", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Searching vehicles", zap.Int("page", filters.Page), zap.Int("pageSize", filters.PageSize))

	result, err := c.vehicleUseCase.SearchPaginated(filters)
	if err != nil {
		c.Logger.Error("Error searching vehicles", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// Mapper helper
func mapVehicleToResponse(v *vehicleDomain.Vehicle) *ResponseVehicle {
	return &ResponseVehicle{
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
