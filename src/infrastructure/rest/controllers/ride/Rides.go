package ride

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gbrayhan/microservices-go/src/application/usecases/ride"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	rideDomain "github.com/gbrayhan/microservices-go/src/domain/ride"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Request/Response Structures
type CreateRideRequest struct {
	RiderID          int     `json:"riderId" binding:"required"`
	PickupLatitude   float64 `json:"pickupLatitude" binding:"required"`
	PickupLongitude  float64 `json:"pickupLongitude" binding:"required"`
	PickupAddress    string  `json:"pickupAddress" binding:"required"`
	DropoffLatitude  float64 `json:"dropoffLatitude" binding:"required"`
	DropoffLongitude float64 `json:"dropoffLongitude" binding:"required"`
	DropoffAddress   string  `json:"dropoffAddress" binding:"required"`
	Notes            string  `json:"notes"`
}

type AssignDriverRequest struct {
	DriverID  int `json:"driverId" binding:"required"`
	VehicleID int `json:"vehicleId" binding:"required"`
}

type CompleteRideRequest struct {
	ActualDistance float64 `json:"actualDistance" binding:"required,min=0"`
	ActualDuration int     `json:"actualDuration" binding:"required,min=0"`
	ActualFare     float64 `json:"actualFare" binding:"required,min=0"`
}

type CancelRideRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RateRequest struct {
	Rating int `json:"rating" binding:"required,min=1,max=5"`
}

type ResponseRide struct {
	ID                 int        `json:"id"`
	RiderID            int        `json:"riderId"`
	DriverID           *int       `json:"driverId,omitempty"`
	VehicleID          *int       `json:"vehicleId,omitempty"`
	Status             string     `json:"status"`
	PickupLatitude     float64    `json:"pickupLatitude"`
	PickupLongitude    float64    `json:"pickupLongitude"`
	PickupAddress      string     `json:"pickupAddress"`
	DropoffLatitude    float64    `json:"dropoffLatitude"`
	DropoffLongitude   float64    `json:"dropoffLongitude"`
	DropoffAddress     string     `json:"dropoffAddress"`
	RequestedAt        time.Time  `json:"requestedAt"`
	MatchedAt          *time.Time `json:"matchedAt,omitempty"`
	StartedAt          *time.Time `json:"startedAt,omitempty"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	CancelledAt        *time.Time `json:"cancelledAt,omitempty"`
	CancellationReason string     `json:"cancellationReason,omitempty"`
	EstimatedDistance  float64    `json:"estimatedDistance"`
	ActualDistance     float64    `json:"actualDistance,omitempty"`
	EstimatedDuration  int        `json:"estimatedDuration"`
	ActualDuration     int        `json:"actualDuration,omitempty"`
	EstimatedFare      float64    `json:"estimatedFare"`
	ActualFare         float64    `json:"actualFare,omitempty"`
	RiderRating        *int       `json:"riderRating,omitempty"`
	DriverRating       *int       `json:"driverRating,omitempty"`
	Notes              string     `json:"notes,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type IRideController interface {
	CreateRide(ctx *gin.Context)
	GetAllRides(ctx *gin.Context)
	GetRideByID(ctx *gin.Context)
	GetRidesByRider(ctx *gin.Context)
	GetRidesByDriver(ctx *gin.Context)
	GetPendingRides(ctx *gin.Context)
	UpdateRide(ctx *gin.Context)
	DeleteRide(ctx *gin.Context)
	AssignDriver(ctx *gin.Context)
	MatchDriverAutomatically(ctx *gin.Context)
	StartRide(ctx *gin.Context)
	CompleteRide(ctx *gin.Context)
	CancelRide(ctx *gin.Context)
	RateDriver(ctx *gin.Context)
	RateRider(ctx *gin.Context)
	SearchRides(ctx *gin.Context)
}

type RideController struct {
	rideUseCase ride.IRideUseCase
	Logger      *logger.Logger
}

func NewRideController(rideUseCase ride.IRideUseCase, loggerInstance *logger.Logger) IRideController {
	return &RideController{
		rideUseCase: rideUseCase,
		Logger:      loggerInstance,
	}
}

// CreateRide godoc
// @Summary Create a new ride request
// @Description Create a new ride request from a rider
// @Tags rides
// @Accept json
// @Produce json
// @Param ride body CreateRideRequest true "Ride data"
// @Success 201 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride [post]
func (c *RideController) CreateRide(ctx *gin.Context) {
	c.Logger.Info("Creating new ride request")

	var request CreateRideRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	newRide := &rideDomain.Ride{
		RiderID:          request.RiderID,
		PickupLatitude:   request.PickupLatitude,
		PickupLongitude:  request.PickupLongitude,
		PickupAddress:    request.PickupAddress,
		DropoffLatitude:  request.DropoffLatitude,
		DropoffLongitude: request.DropoffLongitude,
		DropoffAddress:   request.DropoffAddress,
		Notes:            request.Notes,
		Status:           common.RideStatusPending,
	}

	createdRide, err := c.rideUseCase.Create(newRide)
	if err != nil {
		c.Logger.Error("Error creating ride", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	c.Logger.Info("Ride created successfully", zap.Int("id", createdRide.ID))
	ctx.JSON(http.StatusCreated, mapRideToResponse(createdRide))
}

// GetAllRides godoc
// @Summary Get all rides
// @Description Get all rides in the system
// @Tags rides
// @Produce json
// @Success 200 {array} ResponseRide
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride [get]
func (c *RideController) GetAllRides(ctx *gin.Context) {
	c.Logger.Info("Getting all rides")

	rides, err := c.rideUseCase.GetAll()
	if err != nil {
		c.Logger.Error("Error getting rides", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponseRide, 0, len(*rides))
	for _, r := range *rides {
		response = append(response, *mapRideToResponse(&r))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetRideByID godoc
// @Summary Get ride by ID
// @Description Get a specific ride by ID
// @Tags rides
// @Produce json
// @Param id path int true "Ride ID"
// @Success 200 {object} ResponseRide
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id} [get]
func (c *RideController) GetRideByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Getting ride by ID", zap.Int("id", id))

	ride, err := c.rideUseCase.GetByID(id)
	if err != nil {
		c.Logger.Error("Error getting ride", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	if ride.ID == 0 {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotFound))
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(ride))
}

// GetRidesByRider godoc
// @Summary Get rides by rider ID
// @Description Get all rides for a specific rider
// @Tags rides
// @Produce json
// @Param riderId path int true "Rider ID"
// @Success 200 {array} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/rider/{riderId} [get]
func (c *RideController) GetRidesByRider(ctx *gin.Context) {
	riderID, err := strconv.Atoi(ctx.Param("riderId"))
	if err != nil {
		c.Logger.Warn("Invalid rider ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Getting rides by rider", zap.Int("riderId", riderID))

	rides, err := c.rideUseCase.GetByRiderID(riderID)
	if err != nil {
		c.Logger.Error("Error getting rides by rider", zap.Error(err), zap.Int("riderId", riderID))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponseRide, 0, len(*rides))
	for _, r := range *rides {
		response = append(response, *mapRideToResponse(&r))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetRidesByDriver godoc
// @Summary Get rides by driver ID
// @Description Get all rides for a specific driver
// @Tags rides
// @Produce json
// @Param driverId path int true "Driver ID"
// @Success 200 {array} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/driver/{driverId} [get]
func (c *RideController) GetRidesByDriver(ctx *gin.Context) {
	driverID, err := strconv.Atoi(ctx.Param("driverId"))
	if err != nil {
		c.Logger.Warn("Invalid driver ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Getting rides by driver", zap.Int("driverId", driverID))

	rides, err := c.rideUseCase.GetByDriverID(driverID)
	if err != nil {
		c.Logger.Error("Error getting rides by driver", zap.Error(err), zap.Int("driverId", driverID))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponseRide, 0, len(*rides))
	for _, r := range *rides {
		response = append(response, *mapRideToResponse(&r))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetPendingRides godoc
// @Summary Get pending rides
// @Description Get all rides that are waiting for driver assignment
// @Tags rides
// @Produce json
// @Success 200 {array} ResponseRide
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/pending [get]
func (c *RideController) GetPendingRides(ctx *gin.Context) {
	c.Logger.Info("Getting pending rides")

	rides, err := c.rideUseCase.GetPendingRides()
	if err != nil {
		c.Logger.Error("Error getting pending rides", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponseRide, 0, len(*rides))
	for _, r := range *rides {
		response = append(response, *mapRideToResponse(&r))
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateRide godoc
// @Summary Update ride
// @Description Update a ride's information
// @Tags rides
// @Accept json
// @Produce json
// @Param id path int true "Ride ID"
// @Param ride body map[string]interface{} true "Ride data"
// @Success 200 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id} [put]
func (c *RideController) UpdateRide(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var updateMap map[string]interface{}
	if err := controllers.BindJSON(ctx, &updateMap); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Updating ride", zap.Int("id", id))

	updatedRide, err := c.rideUseCase.Update(id, updateMap)
	if err != nil {
		c.Logger.Error("Error updating ride", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	if updatedRide.ID == 0 {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotFound))
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(updatedRide))
}

// DeleteRide godoc
// @Summary Delete ride
// @Description Delete a ride
// @Tags rides
// @Param id path int true "Ride ID"
// @Success 200 {object} controllers.MessageResponse
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id} [delete]
func (c *RideController) DeleteRide(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Deleting ride", zap.Int("id", id))

	err = c.rideUseCase.Delete(id)
	if err != nil {
		if appErr, ok := err.(*domainErrors.AppError); ok && appErr.Type == domainErrors.NotFound {
			_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotFound))
			return
		}
		c.Logger.Error("Error deleting ride", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Ride deleted successfully"})
}

// AssignDriver godoc
// @Summary Manually assign driver to ride
// @Description Manually assign a driver and vehicle to a ride
// @Tags rides
// @Accept json
// @Produce json
// @Param id path int true "Ride ID"
// @Param assignment body AssignDriverRequest true "Driver assignment data"
// @Success 200 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id}/assign-driver [post]
func (c *RideController) AssignDriver(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request AssignDriverRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Assigning driver to ride",
		zap.Int("rideId", id),
		zap.Int("driverId", request.DriverID),
		zap.Int("vehicleId", request.VehicleID))

	ride, err := c.rideUseCase.AssignDriver(id, request.DriverID, request.VehicleID)
	if err != nil {
		c.Logger.Error("Error assigning driver", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(ride))
}

// MatchDriverAutomatically godoc
// @Summary Automatically match driver to ride
// @Description Automatically find and assign the best available driver
// @Tags rides
// @Produce json
// @Param id path int true "Ride ID"
// @Success 200 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id}/match-driver [post]
func (c *RideController) MatchDriverAutomatically(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Automatically matching driver for ride", zap.Int("rideId", id))

	ride, err := c.rideUseCase.MatchDriverAutomatically(id)
	if err != nil {
		c.Logger.Error("Error matching driver", zap.Error(err), zap.Int("rideId", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(ride))
}

// StartRide godoc
// @Summary Start a ride
// @Description Mark a matched ride as started
// @Tags rides
// @Produce json
// @Param id path int true "Ride ID"
// @Success 200 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id}/start [post]
func (c *RideController) StartRide(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Starting ride", zap.Int("rideId", id))

	ride, err := c.rideUseCase.StartRide(id)
	if err != nil {
		c.Logger.Error("Error starting ride", zap.Error(err), zap.Int("rideId", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(ride))
}

// CompleteRide godoc
// @Summary Complete a ride
// @Description Mark an in-progress ride as completed
// @Tags rides
// @Accept json
// @Produce json
// @Param id path int true "Ride ID"
// @Param completion body CompleteRideRequest true "Ride completion data"
// @Success 200 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id}/complete [post]
func (c *RideController) CompleteRide(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request CompleteRideRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Completing ride",
		zap.Int("rideId", id),
		zap.Float64("actualDistance", request.ActualDistance),
		zap.Float64("actualFare", request.ActualFare))

	ride, err := c.rideUseCase.CompleteRide(id, request.ActualDistance, request.ActualDuration, request.ActualFare)
	if err != nil {
		c.Logger.Error("Error completing ride", zap.Error(err), zap.Int("rideId", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(ride))
}

// CancelRide godoc
// @Summary Cancel a ride
// @Description Cancel a pending or matched ride
// @Tags rides
// @Accept json
// @Produce json
// @Param id path int true "Ride ID"
// @Param cancellation body CancelRideRequest true "Cancellation data"
// @Success 200 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id}/cancel [post]
func (c *RideController) CancelRide(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request CancelRideRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Cancelling ride", zap.Int("rideId", id), zap.String("reason", request.Reason))

	ride, err := c.rideUseCase.CancelRide(id, request.Reason)
	if err != nil {
		c.Logger.Error("Error cancelling ride", zap.Error(err), zap.Int("rideId", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(ride))
}

// RateDriver godoc
// @Summary Rate driver
// @Description Rider rates the driver after completed ride
// @Tags rides
// @Accept json
// @Produce json
// @Param id path int true "Ride ID"
// @Param rating body RateRequest true "Rating data"
// @Success 200 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id}/rate-driver [post]
func (c *RideController) RateDriver(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request RateRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Rating driver", zap.Int("rideId", id), zap.Int("rating", request.Rating))

	ride, err := c.rideUseCase.RateDriver(id, request.Rating)
	if err != nil {
		c.Logger.Error("Error rating driver", zap.Error(err), zap.Int("rideId", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(ride))
}

// RateRider godoc
// @Summary Rate rider
// @Description Driver rates the rider after completed ride
// @Tags rides
// @Accept json
// @Produce json
// @Param id path int true "Ride ID"
// @Param rating body RateRequest true "Rating data"
// @Success 200 {object} ResponseRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/{id}/rate-rider [post]
func (c *RideController) RateRider(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request RateRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Rating rider", zap.Int("rideId", id), zap.Int("rating", request.Rating))

	ride, err := c.rideUseCase.RateRider(id, request.Rating)
	if err != nil {
		c.Logger.Error("Error rating rider", zap.Error(err), zap.Int("rideId", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapRideToResponse(ride))
}

// SearchRides godoc
// @Summary Search rides
// @Description Search rides with pagination
// @Tags rides
// @Produce json
// @Param filters body domain.DataFilters true "Search filters"
// @Success 200 {object} ride.SearchResultRide
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/ride/search [post]
func (c *RideController) SearchRides(ctx *gin.Context) {
	filters, err := controllers.GetSearchFilters(ctx)
	if err != nil {
		c.Logger.Error("Invalid search filters", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Searching rides", zap.Int("page", filters.Page), zap.Int("pageSize", filters.PageSize))

	result, err := c.rideUseCase.SearchPaginated(filters)
	if err != nil {
		c.Logger.Error("Error searching rides", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// Mapper helper
func mapRideToResponse(r *rideDomain.Ride) *ResponseRide {
	return &ResponseRide{
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
