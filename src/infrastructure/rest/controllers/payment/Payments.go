package payment

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gbrayhan/microservices-go/src/application/usecases/payment"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	paymentDomain "github.com/gbrayhan/microservices-go/src/domain/payment"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Request/Response Structures
type CreatePaymentRequest struct {
	RideID        int     `json:"rideId" binding:"required"`
	PaymentMethod string  `json:"paymentMethod" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,min=0"`
}

type ProcessPaymentRequest struct {
	TransactionID string `json:"transactionId" binding:"required"`
}

type FailPaymentRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RefundPaymentRequest struct {
	Amount float64 `json:"amount" binding:"required,min=0"`
	Reason string  `json:"reason" binding:"required"`
}

type ResponsePayment struct {
	ID              int        `json:"id"`
	RideID          int        `json:"rideId"`
	RiderID         int        `json:"riderId"`
	DriverID        int        `json:"driverId"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	PaymentMethod   string     `json:"paymentMethod"`
	Status          string     `json:"status"`
	TransactionID   string     `json:"transactionId,omitempty"`
	PaymentIntentID string     `json:"paymentIntentId,omitempty"`
	ProcessedAt     *time.Time `json:"processedAt,omitempty"`
	RefundedAt      *time.Time `json:"refundedAt,omitempty"`
	RefundAmount    float64    `json:"refundAmount,omitempty"`
	RefundReason    string     `json:"refundReason,omitempty"`
	FailureReason   string     `json:"failureReason,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type IPaymentController interface {
	CreatePayment(ctx *gin.Context)
	GetAllPayments(ctx *gin.Context)
	GetPaymentByID(ctx *gin.Context)
	GetPaymentByRide(ctx *gin.Context)
	GetPaymentsByRider(ctx *gin.Context)
	GetPaymentsByDriver(ctx *gin.Context)
	UpdatePayment(ctx *gin.Context)
	DeletePayment(ctx *gin.Context)
	ProcessPayment(ctx *gin.Context)
	FailPayment(ctx *gin.Context)
	RefundPayment(ctx *gin.Context)
	GetPendingPayments(ctx *gin.Context)
	GetDriverEarnings(ctx *gin.Context)
	SearchPayments(ctx *gin.Context)
}

type PaymentController struct {
	paymentUseCase payment.IPaymentUseCase
	Logger         *logger.Logger
}

func NewPaymentController(paymentUseCase payment.IPaymentUseCase, loggerInstance *logger.Logger) IPaymentController {
	return &PaymentController{
		paymentUseCase: paymentUseCase,
		Logger:         loggerInstance,
	}
}

// CreatePayment godoc
// @Summary Create a new payment
// @Description Create a payment for a completed ride
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body CreatePaymentRequest true "Payment data"
// @Success 201 {object} ResponsePayment
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment [post]
func (c *PaymentController) CreatePayment(ctx *gin.Context) {
	c.Logger.Info("Creating new payment")

	var request CreatePaymentRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	// Validate payment method
	paymentMethod := common.PaymentMethod(request.PaymentMethod)
	if !paymentMethod.IsValid() {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	createdPayment, err := c.paymentUseCase.CreatePaymentForRide(request.RideID, paymentMethod)
	if err != nil {
		c.Logger.Error("Error creating payment", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	c.Logger.Info("Payment created successfully", zap.Int("id", createdPayment.ID))
	ctx.JSON(http.StatusCreated, mapPaymentToResponse(createdPayment))
}

// GetAllPayments godoc
// @Summary Get all payments
// @Description Get all payments in the system
// @Tags payments
// @Produce json
// @Success 200 {array} ResponsePayment
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment [get]
func (c *PaymentController) GetAllPayments(ctx *gin.Context) {
	c.Logger.Info("Getting all payments")

	payments, err := c.paymentUseCase.GetAll()
	if err != nil {
		c.Logger.Error("Error getting payments", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponsePayment, 0, len(*payments))
	for _, p := range *payments {
		response = append(response, *mapPaymentToResponse(&p))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetPaymentByID godoc
// @Summary Get payment by ID
// @Description Get a specific payment by ID
// @Tags payments
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} ResponsePayment
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/{id} [get]
func (c *PaymentController) GetPaymentByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid payment ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Getting payment by ID", zap.Int("id", id))

	payment, err := c.paymentUseCase.GetByID(id)
	if err != nil {
		c.Logger.Error("Error getting payment", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	if payment.ID == 0 {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotFound))
		return
	}

	ctx.JSON(http.StatusOK, mapPaymentToResponse(payment))
}

// GetPaymentByRide godoc
// @Summary Get payment by ride ID
// @Description Get payment for a specific ride
// @Tags payments
// @Produce json
// @Param rideId path int true "Ride ID"
// @Success 200 {object} ResponsePayment
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/ride/{rideId} [get]
func (c *PaymentController) GetPaymentByRide(ctx *gin.Context) {
	rideID, err := strconv.Atoi(ctx.Param("rideId"))
	if err != nil {
		c.Logger.Warn("Invalid ride ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Getting payment by ride", zap.Int("rideId", rideID))

	payment, err := c.paymentUseCase.GetByRideID(rideID)
	if err != nil {
		c.Logger.Error("Error getting payment by ride", zap.Error(err), zap.Int("rideId", rideID))
		_ = ctx.Error(err)
		return
	}

	if payment.ID == 0 {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotFound))
			return
	}

	ctx.JSON(http.StatusOK, mapPaymentToResponse(payment))
}

// GetPaymentsByRider godoc
// @Summary Get payments by rider ID
// @Description Get all payments for a specific rider
// @Tags payments
// @Produce json
// @Param riderId path int true "Rider ID"
// @Success 200 {array} ResponsePayment
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/rider/{riderId} [get]
func (c *PaymentController) GetPaymentsByRider(ctx *gin.Context) {
	riderID, err := strconv.Atoi(ctx.Param("riderId"))
	if err != nil {
		c.Logger.Warn("Invalid rider ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Getting payments by rider", zap.Int("riderId", riderID))

	payments, err := c.paymentUseCase.GetByRiderID(riderID)
	if err != nil {
		c.Logger.Error("Error getting payments by rider", zap.Error(err), zap.Int("riderId", riderID))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponsePayment, 0, len(*payments))
	for _, p := range *payments {
		response = append(response, *mapPaymentToResponse(&p))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetPaymentsByDriver godoc
// @Summary Get payments by driver ID
// @Description Get all payments for a specific driver
// @Tags payments
// @Produce json
// @Param driverId path int true "Driver ID"
// @Success 200 {array} ResponsePayment
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/driver/{driverId} [get]
func (c *PaymentController) GetPaymentsByDriver(ctx *gin.Context) {
	driverID, err := strconv.Atoi(ctx.Param("driverId"))
	if err != nil {
		c.Logger.Warn("Invalid driver ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Getting payments by driver", zap.Int("driverId", driverID))

	payments, err := c.paymentUseCase.GetByDriverID(driverID)
	if err != nil {
		c.Logger.Error("Error getting payments by driver", zap.Error(err), zap.Int("driverId", driverID))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponsePayment, 0, len(*payments))
	for _, p := range *payments {
		response = append(response, *mapPaymentToResponse(&p))
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdatePayment godoc
// @Summary Update payment
// @Description Update a payment's information
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param payment body map[string]interface{} true "Payment data"
// @Success 200 {object} ResponsePayment
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/{id} [put]
func (c *PaymentController) UpdatePayment(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid payment ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var updateMap map[string]interface{}
	if err := controllers.BindJSON(ctx, &updateMap); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Updating payment", zap.Int("id", id))

	updatedPayment, err := c.paymentUseCase.Update(id, updateMap)
	if err != nil {
		c.Logger.Error("Error updating payment", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	if updatedPayment.ID == 0 {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotFound))
		return
	}

	ctx.JSON(http.StatusOK, mapPaymentToResponse(updatedPayment))
}

// DeletePayment godoc
// @Summary Delete payment
// @Description Delete a payment
// @Tags payments
// @Param id path int true "Payment ID"
// @Success 200 {object} controllers.MessageResponse
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/{id} [delete]
func (c *PaymentController) DeletePayment(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid payment ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Deleting payment", zap.Int("id", id))

	err = c.paymentUseCase.Delete(id)
	if err != nil {
		if appErr, ok := err.(*domainErrors.AppError); ok && appErr.Type == domainErrors.NotFound {
			_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.NotFound))
			return
		}
		c.Logger.Error("Error deleting payment", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Payment deleted successfully"})
}

// ProcessPayment godoc
// @Summary Process payment
// @Description Mark a payment as completed
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param process body ProcessPaymentRequest true "Process data"
// @Success 200 {object} ResponsePayment
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/{id}/process [post]
func (c *PaymentController) ProcessPayment(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid payment ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request ProcessPaymentRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Processing payment", zap.Int("id", id), zap.String("transactionId", request.TransactionID))

	payment, err := c.paymentUseCase.MarkAsCompleted(id, request.TransactionID)
	if err != nil {
		c.Logger.Error("Error processing payment", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapPaymentToResponse(payment))
}

// FailPayment godoc
// @Summary Fail payment
// @Description Mark a payment as failed
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param fail body FailPaymentRequest true "Failure data"
// @Success 200 {object} ResponsePayment
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/{id}/fail [post]
func (c *PaymentController) FailPayment(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid payment ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request FailPaymentRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Failing payment", zap.Int("id", id), zap.String("reason", request.Reason))

	payment, err := c.paymentUseCase.MarkAsFailed(id, request.Reason)
	if err != nil {
		c.Logger.Error("Error failing payment", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapPaymentToResponse(payment))
}

// RefundPayment godoc
// @Summary Refund payment
// @Description Process a refund for a payment
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param refund body RefundPaymentRequest true "Refund data"
// @Success 200 {object} ResponsePayment
// @Failure 400 {object} controllers.MessageResponse
// @Failure 404 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/{id}/refund [post]
func (c *PaymentController) RefundPayment(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Warn("Invalid payment ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var request RefundPaymentRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Invalid request body", zap.Error(err))
		return
	}

	c.Logger.Info("Refunding payment",
		zap.Int("id", id),
		zap.Float64("amount", request.Amount),
		zap.String("reason", request.Reason))

	payment, err := c.paymentUseCase.ProcessRefund(id, request.Amount, request.Reason)
	if err != nil {
		c.Logger.Error("Error refunding payment", zap.Error(err), zap.Int("id", id))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, mapPaymentToResponse(payment))
}

// GetPendingPayments godoc
// @Summary Get pending payments
// @Description Get all payments that are pending
// @Tags payments
// @Produce json
// @Success 200 {array} ResponsePayment
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/pending [get]
func (c *PaymentController) GetPendingPayments(ctx *gin.Context) {
	c.Logger.Info("Getting pending payments")

	payments, err := c.paymentUseCase.GetPendingPayments()
	if err != nil {
		c.Logger.Error("Error getting pending payments", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	response := make([]ResponsePayment, 0, len(*payments))
	for _, p := range *payments {
		response = append(response, *mapPaymentToResponse(&p))
	}

	ctx.JSON(http.StatusOK, response)
}

// GetDriverEarnings godoc
// @Summary Get driver earnings
// @Description Get total earnings for a driver in a time period
// @Tags payments
// @Produce json
// @Param driverId path int true "Driver ID"
// @Param from query string false "From date (RFC3339)"
// @Param to query string false "To date (RFC3339)"
// @Success 200 {object} map[string]float64
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/driver/{driverId}/earnings [get]
func (c *PaymentController) GetDriverEarnings(ctx *gin.Context) {
	driverID, err := strconv.Atoi(ctx.Param("driverId"))
	if err != nil {
		c.Logger.Warn("Invalid driver ID", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	// Parse date range
	from := time.Now().AddDate(0, -1, 0) // Default: last month
	to := time.Now()

	if fromStr := ctx.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsed
		}
	}
	if toStr := ctx.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsed
		}
	}

	c.Logger.Info("Getting driver earnings",
		zap.Int("driverId", driverID),
		zap.Time("from", from),
		zap.Time("to", to))

	total, err := c.paymentUseCase.GetTotalEarnings(driverID, from, to)
	if err != nil {
		c.Logger.Error("Error getting driver earnings", zap.Error(err), zap.Int("driverId", driverID))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"driverId": driverID,
		"from":     from,
		"to":       to,
		"total":    total,
		"currency": "USD",
	})
}

// SearchPayments godoc
// @Summary Search payments
// @Description Search payments with pagination
// @Tags payments
// @Produce json
// @Param filters body domain.DataFilters true "Search filters"
// @Success 200 {object} payment.SearchResultPayment
// @Failure 400 {object} controllers.MessageResponse
// @Failure 500 {object} controllers.MessageResponse
// @Router /v1/payment/search [post]
func (c *PaymentController) SearchPayments(ctx *gin.Context) {
	filters, err := controllers.GetSearchFilters(ctx)
	if err != nil {
		c.Logger.Error("Invalid search filters", zap.Error(err))
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.Logger.Info("Searching payments", zap.Int("page", filters.Page), zap.Int("pageSize", filters.PageSize))

	result, err := c.paymentUseCase.SearchPaginated(filters)
	if err != nil {
		c.Logger.Error("Error searching payments", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// Mapper helper
func mapPaymentToResponse(p *paymentDomain.Payment) *ResponsePayment {
	return &ResponsePayment{
		ID:              p.ID,
		RideID:          p.RideID,
		RiderID:         p.RiderID,
		DriverID:        p.DriverID,
		Amount:          p.Amount,
		Currency:        p.Currency,
		PaymentMethod:   string(p.PaymentMethod),
		Status:          string(p.Status),
		TransactionID:   p.TransactionID,
		PaymentIntentID: p.PaymentIntentID,
		ProcessedAt:     p.ProcessedAt,
		RefundedAt:      p.RefundedAt,
		RefundAmount:    p.RefundAmount,
		RefundReason:    p.RefundReason,
		FailureReason:   p.FailureReason,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}
