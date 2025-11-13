package payment

import (
	"errors"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
)

// Payment represents a payment transaction in the system
type Payment struct {
	ID              int
	RideID          int // Foreign key to Ride
	RiderID         int // Foreign key to User (rider)
	DriverID        int // Foreign key to User (driver)
	Amount          float64
	Currency        string // e.g., USD, EUR
	PaymentMethod   common.PaymentMethod
	Status          common.PaymentStatus
	TransactionID   string // External payment gateway transaction ID
	PaymentIntentID string // Payment intent ID (for multi-step payments)
	ProcessedAt     *time.Time
	RefundedAt      *time.Time
	RefundAmount    float64
	RefundReason    string
	FailureReason   string
	Metadata        string // JSON string for additional data
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// IsPending checks if the payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == common.PaymentStatusPending
}

// IsCompleted checks if the payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == common.PaymentStatusCompleted
}

// IsFailed checks if the payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == common.PaymentStatusFailed
}

// IsRefunded checks if the payment was refunded
func (p *Payment) IsRefunded() bool {
	return p.Status == common.PaymentStatusRefunded
}

// CanBeRefunded checks if the payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.IsCompleted() && p.RefundedAt == nil
}

// MarkAsCompleted marks the payment as completed
func (p *Payment) MarkAsCompleted(transactionID string) error {
	if !p.IsPending() {
		return errors.New("can only complete pending payments")
	}
	p.Status = common.PaymentStatusCompleted
	p.TransactionID = transactionID
	now := time.Now()
	p.ProcessedAt = &now
	return nil
}

// MarkAsFailed marks the payment as failed
func (p *Payment) MarkAsFailed(reason string) error {
	if !p.IsPending() {
		return errors.New("can only fail pending payments")
	}
	p.Status = common.PaymentStatusFailed
	p.FailureReason = reason
	now := time.Now()
	p.ProcessedAt = &now
	return nil
}

// ProcessRefund processes a refund for the payment
func (p *Payment) ProcessRefund(amount float64, reason string) error {
	if !p.CanBeRefunded() {
		return errors.New("payment cannot be refunded")
	}
	if amount <= 0 || amount > p.Amount {
		return errors.New("invalid refund amount")
	}
	p.Status = common.PaymentStatusRefunded
	p.RefundAmount = amount
	p.RefundReason = reason
	now := time.Now()
	p.RefundedAt = &now
	return nil
}

// GetNetAmount returns the net amount after refunds
func (p *Payment) GetNetAmount() float64 {
	if p.IsRefunded() {
		return p.Amount - p.RefundAmount
	}
	return p.Amount
}

// SearchResultPayment represents paginated payment search results
type SearchResultPayment struct {
	Data       *[]Payment
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// IPaymentService defines the interface for payment operations
type IPaymentService interface {
	GetAll() (*[]Payment, error)
	GetByID(id int) (*Payment, error)
	GetByRideID(rideID int) (*Payment, error)
	GetByRiderID(riderID int) (*[]Payment, error)
	GetByDriverID(driverID int) (*[]Payment, error)
	Create(newPayment *Payment) (*Payment, error)
	Delete(id int) error
	Update(id int, paymentMap map[string]interface{}) (*Payment, error)
	MarkAsCompleted(id int, transactionID string) (*Payment, error)
	MarkAsFailed(id int, reason string) (*Payment, error)
	ProcessRefund(id int, amount float64, reason string) (*Payment, error)
	SearchPaginated(filters domain.DataFilters) (*SearchResultPayment, error)
	GetPendingPayments() (*[]Payment, error)
	GetTotalEarnings(driverID int, from time.Time, to time.Time) (float64, error)
}
