package payment

import (
	"fmt"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	paymentDomain "github.com/gbrayhan/microservices-go/src/domain/payment"
	rideDomain "github.com/gbrayhan/microservices-go/src/domain/ride"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

// PaymentRepositoryInterface defines the interface for payment repository
type PaymentRepositoryInterface interface {
	GetAll() (*[]paymentDomain.Payment, error)
	GetByID(id int) (*paymentDomain.Payment, error)
	GetByRideID(rideID int) (*paymentDomain.Payment, error)
	GetByRiderID(riderID int) (*[]paymentDomain.Payment, error)
	GetByDriverID(driverID int) (*[]paymentDomain.Payment, error)
	Create(payment *paymentDomain.Payment) (*paymentDomain.Payment, error)
	Update(id int, paymentMap map[string]interface{}) (*paymentDomain.Payment, error)
	Delete(id int) error
	SearchPaginated(filters domain.DataFilters) (*paymentDomain.SearchResultPayment, error)
	GetPendingPayments() (*[]paymentDomain.Payment, error)
	GetTotalEarnings(driverID int, from time.Time, to time.Time) (float64, error)
}

// RideRepositoryInterface defines the interface for ride repository
type RideRepositoryInterface interface {
	GetByID(id int) (*rideDomain.Ride, error)
}

// IPaymentUseCase defines the interface for payment use cases
type IPaymentUseCase interface {
	GetAll() (*[]paymentDomain.Payment, error)
	GetByID(id int) (*paymentDomain.Payment, error)
	GetByRideID(rideID int) (*paymentDomain.Payment, error)
	GetByRiderID(riderID int) (*[]paymentDomain.Payment, error)
	GetByDriverID(driverID int) (*[]paymentDomain.Payment, error)
	Create(newPayment *paymentDomain.Payment) (*paymentDomain.Payment, error)
	Update(id int, paymentMap map[string]interface{}) (*paymentDomain.Payment, error)
	Delete(id int) error
	MarkAsCompleted(id int, transactionID string) (*paymentDomain.Payment, error)
	MarkAsFailed(id int, reason string) (*paymentDomain.Payment, error)
	ProcessRefund(id int, amount float64, reason string) (*paymentDomain.Payment, error)
	SearchPaginated(filters domain.DataFilters) (*paymentDomain.SearchResultPayment, error)
	GetPendingPayments() (*[]paymentDomain.Payment, error)
	GetTotalEarnings(driverID int, from time.Time, to time.Time) (float64, error)
	CreatePaymentForRide(rideID int, paymentMethod common.PaymentMethod) (*paymentDomain.Payment, error)
}

// PaymentUseCase implements IPaymentUseCase
type PaymentUseCase struct {
	paymentRepository PaymentRepositoryInterface
	rideRepository    RideRepositoryInterface
	Logger            *logger.Logger
}

// NewPaymentUseCase creates a new payment use case
func NewPaymentUseCase(
	paymentRepository PaymentRepositoryInterface,
	rideRepository RideRepositoryInterface,
	logger *logger.Logger,
) IPaymentUseCase {
	return &PaymentUseCase{
		paymentRepository: paymentRepository,
		rideRepository:    rideRepository,
		Logger:            logger,
	}
}

// GetAll retrieves all payments
func (s *PaymentUseCase) GetAll() (*[]paymentDomain.Payment, error) {
	s.Logger.Info("Getting all payments")
	return s.paymentRepository.GetAll()
}

// GetByID retrieves a payment by ID
func (s *PaymentUseCase) GetByID(id int) (*paymentDomain.Payment, error) {
	s.Logger.Info("Getting payment by ID", zap.Int("id", id))
	payment, err := s.paymentRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting payment", zap.Error(err), zap.Int("id", id))
		return nil, err
	}
	if payment.ID == 0 {
		s.Logger.Warn("Payment not found", zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return payment, nil
}

// GetByRideID retrieves payment for a specific ride
func (s *PaymentUseCase) GetByRideID(rideID int) (*paymentDomain.Payment, error) {
	s.Logger.Info("Getting payment by ride ID", zap.Int("rideID", rideID))
	return s.paymentRepository.GetByRideID(rideID)
}

// GetByRiderID retrieves all payments for a rider
func (s *PaymentUseCase) GetByRiderID(riderID int) (*[]paymentDomain.Payment, error) {
	s.Logger.Info("Getting payments by rider ID", zap.Int("riderID", riderID))
	return s.paymentRepository.GetByRiderID(riderID)
}

// GetByDriverID retrieves all payments for a driver
func (s *PaymentUseCase) GetByDriverID(driverID int) (*[]paymentDomain.Payment, error) {
	s.Logger.Info("Getting payments by driver ID", zap.Int("driverID", driverID))
	return s.paymentRepository.GetByDriverID(driverID)
}

// Create creates a new payment
func (s *PaymentUseCase) Create(newPayment *paymentDomain.Payment) (*paymentDomain.Payment, error) {
	s.Logger.Info("Creating new payment",
		zap.Int("rideID", newPayment.RideID),
		zap.Float64("amount", newPayment.Amount),
		zap.String("method", string(newPayment.PaymentMethod)))

	// Validate payment method
	if !newPayment.PaymentMethod.IsValid() {
		s.Logger.Warn("Invalid payment method", zap.String("method", string(newPayment.PaymentMethod)))
		return nil, domainErrors.NewAppError(
			fmt.Errorf("invalid payment method: %s", newPayment.PaymentMethod),
			domainErrors.ValidationError)
	}

	// Validate ride exists
	ride, err := s.rideRepository.GetByID(newPayment.RideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppError(fmt.Errorf("ride not found"), domainErrors.NotFound)
	}

	// Check if ride is completed
	if !ride.IsCompleted() {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("cannot create payment for non-completed ride"),
			domainErrors.ValidationError)
	}

	// Check if payment already exists for this ride
	existingPayment, err := s.paymentRepository.GetByRideID(newPayment.RideID)
	if err == nil && existingPayment != nil && existingPayment.ID != 0 {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("payment already exists for this ride"),
			domainErrors.ResourceAlreadyExists)
	}

	// Set initial status
	if newPayment.Status == "" {
		newPayment.Status = common.PaymentStatusPending
	}

	// Set currency if not provided
	if newPayment.Currency == "" {
		newPayment.Currency = "USD"
	}

	payment, err := s.paymentRepository.Create(newPayment)
	if err != nil {
		s.Logger.Error("Error creating payment", zap.Error(err))
		return nil, err
	}

	s.Logger.Info("Payment created successfully", zap.Int("id", payment.ID))
	return payment, nil
}

// Update updates a payment
func (s *PaymentUseCase) Update(id int, paymentMap map[string]interface{}) (*paymentDomain.Payment, error) {
	s.Logger.Info("Updating payment", zap.Int("id", id))

	// Check if payment exists
	existingPayment, err := s.paymentRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting payment for update", zap.Error(err), zap.Int("id", id))
		return nil, err
	}
	if existingPayment.ID == 0 {
		s.Logger.Warn("Payment not found for update", zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	payment, err := s.paymentRepository.Update(id, paymentMap)
	if err != nil {
		s.Logger.Error("Error updating payment", zap.Error(err), zap.Int("id", id))
		return nil, err
	}

	s.Logger.Info("Payment updated successfully", zap.Int("id", id))
	return payment, nil
}

// Delete deletes a payment
func (s *PaymentUseCase) Delete(id int) error {
	s.Logger.Info("Deleting payment", zap.Int("id", id))

	// Check if payment exists
	existingPayment, err := s.paymentRepository.GetByID(id)
	if err != nil {
		s.Logger.Error("Error getting payment for deletion", zap.Error(err), zap.Int("id", id))
		return err
	}
	if existingPayment.ID == 0 {
		s.Logger.Warn("Payment not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	// Don't allow deletion of completed payments
	if existingPayment.IsCompleted() {
		return domainErrors.NewAppError(
			fmt.Errorf("cannot delete completed payment"),
			domainErrors.ValidationError)
	}

	err = s.paymentRepository.Delete(id)
	if err != nil {
		s.Logger.Error("Error deleting payment", zap.Error(err), zap.Int("id", id))
		return err
	}

	s.Logger.Info("Payment deleted successfully", zap.Int("id", id))
	return nil
}

// MarkAsCompleted marks a payment as completed
func (s *PaymentUseCase) MarkAsCompleted(id int, transactionID string) (*paymentDomain.Payment, error) {
	s.Logger.Info("Marking payment as completed",
		zap.Int("id", id),
		zap.String("transactionID", transactionID))

	payment, err := s.paymentRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if payment.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = payment.MarkAsCompleted(transactionID)
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	paymentMap := map[string]interface{}{
		"status":         string(common.PaymentStatusCompleted),
		"transaction_id": transactionID,
		"processed_at":   payment.ProcessedAt,
	}

	updatedPayment, err := s.paymentRepository.Update(id, paymentMap)
	if err != nil {
		return nil, err
	}

	s.Logger.Info("Payment marked as completed", zap.Int("id", id))
	return updatedPayment, nil
}

// MarkAsFailed marks a payment as failed
func (s *PaymentUseCase) MarkAsFailed(id int, reason string) (*paymentDomain.Payment, error) {
	s.Logger.Info("Marking payment as failed",
		zap.Int("id", id),
		zap.String("reason", reason))

	payment, err := s.paymentRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if payment.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = payment.MarkAsFailed(reason)
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	paymentMap := map[string]interface{}{
		"status":         string(common.PaymentStatusFailed),
		"failure_reason": reason,
		"processed_at":   payment.ProcessedAt,
	}

	updatedPayment, err := s.paymentRepository.Update(id, paymentMap)
	if err != nil {
		return nil, err
	}

	s.Logger.Info("Payment marked as failed", zap.Int("id", id))
	return updatedPayment, nil
}

// ProcessRefund processes a refund for a payment
func (s *PaymentUseCase) ProcessRefund(id int, amount float64, reason string) (*paymentDomain.Payment, error) {
	s.Logger.Info("Processing refund",
		zap.Int("id", id),
		zap.Float64("amount", amount),
		zap.String("reason", reason))

	payment, err := s.paymentRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if payment.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	err = payment.ProcessRefund(amount, reason)
	if err != nil {
		return nil, domainErrors.NewAppError(err, domainErrors.ValidationError)
	}

	paymentMap := map[string]interface{}{
		"status":        string(common.PaymentStatusRefunded),
		"refund_amount": amount,
		"refund_reason": reason,
		"refunded_at":   payment.RefundedAt,
	}

	updatedPayment, err := s.paymentRepository.Update(id, paymentMap)
	if err != nil {
		return nil, err
	}

	s.Logger.Info("Refund processed successfully", zap.Int("id", id))
	return updatedPayment, nil
}

// SearchPaginated performs paginated search
func (s *PaymentUseCase) SearchPaginated(filters domain.DataFilters) (*paymentDomain.SearchResultPayment, error) {
	s.Logger.Info("Searching payments with pagination",
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))
	return s.paymentRepository.SearchPaginated(filters)
}

// GetPendingPayments retrieves all pending payments
func (s *PaymentUseCase) GetPendingPayments() (*[]paymentDomain.Payment, error) {
	s.Logger.Info("Getting pending payments")
	return s.paymentRepository.GetPendingPayments()
}

// GetTotalEarnings calculates total earnings for a driver in a time period
func (s *PaymentUseCase) GetTotalEarnings(driverID int, from time.Time, to time.Time) (float64, error) {
	s.Logger.Info("Calculating total earnings",
		zap.Int("driverID", driverID),
		zap.Time("from", from),
		zap.Time("to", to))
	return s.paymentRepository.GetTotalEarnings(driverID, from, to)
}

// CreatePaymentForRide creates a payment for a completed ride
func (s *PaymentUseCase) CreatePaymentForRide(rideID int, paymentMethod common.PaymentMethod) (*paymentDomain.Payment, error) {
	s.Logger.Info("Creating payment for ride",
		zap.Int("rideID", rideID),
		zap.String("method", string(paymentMethod)))

	// Get ride details
	ride, err := s.rideRepository.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.ID == 0 {
		return nil, domainErrors.NewAppError(fmt.Errorf("ride not found"), domainErrors.NotFound)
	}

	// Check if ride is completed
	if !ride.IsCompleted() {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("cannot create payment for non-completed ride"),
			domainErrors.ValidationError)
	}

	// Check if driver is assigned
	if ride.DriverID == nil {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("ride has no assigned driver"),
			domainErrors.ValidationError)
	}

	// Check if payment already exists
	existingPayment, err := s.paymentRepository.GetByRideID(rideID)
	if err == nil && existingPayment != nil && existingPayment.ID != 0 {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("payment already exists for this ride"),
			domainErrors.ResourceAlreadyExists)
	}

	// Create payment
	newPayment := &paymentDomain.Payment{
		RideID:        rideID,
		RiderID:       ride.RiderID,
		DriverID:      *ride.DriverID,
		Amount:        ride.ActualFare,
		Currency:      "USD",
		PaymentMethod: paymentMethod,
		Status:        common.PaymentStatusPending,
	}

	return s.Create(newPayment)
}
