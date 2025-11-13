package payment

import (
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	paymentDomain "github.com/gbrayhan/microservices-go/src/domain/payment"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Payment represents the database model for payments
type Payment struct {
	ID              int        `gorm:"primaryKey"`
	RideID          int        `gorm:"column:ride_id;uniqueIndex;not null"`
	RiderID         int        `gorm:"column:rider_id;index"`
	DriverID        int        `gorm:"column:driver_id;index"`
	Amount          float64    `gorm:"column:amount"`
	Currency        string     `gorm:"column:currency;default:'USD'"`
	PaymentMethod   string     `gorm:"column:payment_method"`
	Status          string     `gorm:"column:status;index"`
	TransactionID   string     `gorm:"column:transaction_id"`
	PaymentIntentID string     `gorm:"column:payment_intent_id"`
	ProcessedAt     *time.Time `gorm:"column:processed_at"`
	RefundedAt      *time.Time `gorm:"column:refunded_at"`
	RefundAmount    float64    `gorm:"column:refund_amount"`
	RefundReason    string     `gorm:"column:refund_reason"`
	FailureReason   string     `gorm:"column:failure_reason"`
	Metadata        string     `gorm:"column:metadata;type:text"`
	CreatedAt       time.Time  `gorm:"autoCreateTime:mili"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime:mili"`
}

func (Payment) TableName() string {
	return "payments"
}

var ColumnsPaymentMapping = map[string]string{
	"id":              "id",
	"rideId":          "ride_id",
	"riderId":         "rider_id",
	"driverId":        "driver_id",
	"amount":          "amount",
	"currency":        "currency",
	"paymentMethod":   "payment_method",
	"status":          "status",
	"transactionId":   "transaction_id",
	"paymentIntentId": "payment_intent_id",
	"processedAt":     "processed_at",
	"refundedAt":      "refunded_at",
	"refundAmount":    "refund_amount",
	"refundReason":    "refund_reason",
	"failureReason":   "failure_reason",
	"metadata":        "metadata",
	"createdAt":       "created_at",
	"updatedAt":       "updated_at",
}

// PaymentRepositoryInterface defines the interface for payment repository operations
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

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewPaymentRepository(db *gorm.DB, loggerInstance *logger.Logger) PaymentRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]paymentDomain.Payment, error) {
	var payments []Payment
	if err := r.DB.Order("created_at DESC").Find(&payments).Error; err != nil {
		r.Logger.Error("Error getting all payments", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved all payments", zap.Int("count", len(payments)))
	return arrayToDomainMapper(&payments), nil
}

func (r *Repository) GetByID(id int) (*paymentDomain.Payment, error) {
	var payment Payment
	if err := r.DB.First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("Payment not found", zap.Int("id", id))
			return &paymentDomain.Payment{}, nil
		}
		r.Logger.Error("Error getting payment by ID", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved payment", zap.Int("id", id))
	return payment.toDomainMapper(), nil
}

func (r *Repository) GetByRideID(rideID int) (*paymentDomain.Payment, error) {
	var payment Payment
	if err := r.DB.Where("ride_id = ?", rideID).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("Payment not found for ride", zap.Int("rideID", rideID))
			return &paymentDomain.Payment{}, nil
		}
		r.Logger.Error("Error getting payment by ride ID", zap.Error(err), zap.Int("rideID", rideID))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved payment by ride", zap.Int("rideID", rideID))
	return payment.toDomainMapper(), nil
}

func (r *Repository) GetByRiderID(riderID int) (*[]paymentDomain.Payment, error) {
	var payments []Payment
	if err := r.DB.Where("rider_id = ?", riderID).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		r.Logger.Error("Error getting payments by rider ID", zap.Error(err), zap.Int("riderID", riderID))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved payments by rider",
		zap.Int("riderID", riderID),
		zap.Int("count", len(payments)))
	return arrayToDomainMapper(&payments), nil
}

func (r *Repository) GetByDriverID(driverID int) (*[]paymentDomain.Payment, error) {
	var payments []Payment
	if err := r.DB.Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		r.Logger.Error("Error getting payments by driver ID", zap.Error(err), zap.Int("driverID", driverID))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved payments by driver",
		zap.Int("driverID", driverID),
		zap.Int("count", len(payments)))
	return arrayToDomainMapper(&payments), nil
}

func (r *Repository) Create(paymentDomain *paymentDomain.Payment) (*paymentDomain.Payment, error) {
	r.Logger.Info("Creating new payment",
		zap.Int("rideID", paymentDomain.RideID),
		zap.Float64("amount", paymentDomain.Amount))

	paymentRepository := fromDomainMapper(paymentDomain)
	if err := r.DB.Create(paymentRepository).Error; err != nil {
		r.Logger.Error("Error creating payment", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	r.Logger.Info("Payment created successfully", zap.Int("id", paymentRepository.ID))
	return paymentRepository.toDomainMapper(), nil
}

func (r *Repository) Update(id int, paymentMap map[string]interface{}) (*paymentDomain.Payment, error) {
	r.Logger.Info("Updating payment", zap.Int("id", id))

	var payment Payment
	if err := r.DB.First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("Payment not found for update", zap.Int("id", id))
			return &paymentDomain.Payment{}, nil
		}
		r.Logger.Error("Error finding payment for update", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	if err := r.DB.Model(&payment).Updates(paymentMap).Error; err != nil {
		r.Logger.Error("Error updating payment", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	// Reload to get updated data
	if err := r.DB.First(&payment, id).Error; err != nil {
		r.Logger.Error("Error reloading updated payment", zap.Error(err), zap.Int("id", id))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	r.Logger.Info("Payment updated successfully", zap.Int("id", id))
	return payment.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	r.Logger.Info("Deleting payment", zap.Int("id", id))

	result := r.DB.Delete(&Payment{}, id)
	if result.Error != nil {
		r.Logger.Error("Error deleting payment", zap.Error(result.Error), zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	if result.RowsAffected == 0 {
		r.Logger.Warn("Payment not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	r.Logger.Info("Payment deleted successfully", zap.Int("id", id))
	return nil
}

func (r *Repository) SearchPaginated(filters domain.DataFilters) (*paymentDomain.SearchResultPayment, error) {
	r.Logger.Info("Searching payments with pagination",
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))

	var payments []Payment
	var total int64

	query := r.DB.Model(&Payment{})

	// Apply like filters
	for field, values := range filters.LikeFilters {
		if len(values) > 0 {
			for _, value := range values {
				if value != "" {
					column := ColumnsPaymentMapping[field]
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
			column := ColumnsPaymentMapping[field]
			if column != "" {
				query = query.Where(column+" IN ?", values)
			}
		}
	}

	// Apply date range filters
	for _, dateFilter := range filters.DateRangeFilters {
		column := ColumnsPaymentMapping[dateFilter.Field]
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
			column := ColumnsPaymentMapping[sortField]
			if column != "" {
				query = query.Order(column + " " + string(filters.SortDirection))
			}
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		r.Logger.Error("Error counting payments", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Offset(offset).Limit(filters.PageSize).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		r.Logger.Error("Error fetching paginated payments", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize > 0 {
		totalPages++
	}

	result := &paymentDomain.SearchResultPayment{
		Data:       arrayToDomainMapper(&payments),
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}

	r.Logger.Info("Successfully retrieved paginated payments",
		zap.Int64("total", total),
		zap.Int("page", filters.Page),
		zap.Int("totalPages", totalPages))

	return result, nil
}

func (r *Repository) GetPendingPayments() (*[]paymentDomain.Payment, error) {
	var payments []Payment
	if err := r.DB.Where("status = ?", string(common.PaymentStatusPending)).
		Order("created_at ASC").
		Find(&payments).Error; err != nil {
		r.Logger.Error("Error getting pending payments", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved pending payments", zap.Int("count", len(payments)))
	return arrayToDomainMapper(&payments), nil
}

func (r *Repository) GetTotalEarnings(driverID int, from time.Time, to time.Time) (float64, error) {
	var total float64
	if err := r.DB.Model(&Payment{}).
		Where("driver_id = ? AND status = ? AND created_at BETWEEN ? AND ?",
			driverID, string(common.PaymentStatusCompleted), from, to).
		Select("COALESCE(SUM(amount - refund_amount), 0)").
		Scan(&total).Error; err != nil {
		r.Logger.Error("Error calculating total earnings",
			zap.Error(err),
			zap.Int("driverID", driverID))
		return 0, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully calculated total earnings",
		zap.Int("driverID", driverID),
		zap.Float64("total", total))
	return total, nil
}

// Mappers
func (p *Payment) toDomainMapper() *paymentDomain.Payment {
	return &paymentDomain.Payment{
		ID:              p.ID,
		RideID:          p.RideID,
		RiderID:         p.RiderID,
		DriverID:        p.DriverID,
		Amount:          p.Amount,
		Currency:        p.Currency,
		PaymentMethod:   common.PaymentMethod(p.PaymentMethod),
		Status:          common.PaymentStatus(p.Status),
		TransactionID:   p.TransactionID,
		PaymentIntentID: p.PaymentIntentID,
		ProcessedAt:     p.ProcessedAt,
		RefundedAt:      p.RefundedAt,
		RefundAmount:    p.RefundAmount,
		RefundReason:    p.RefundReason,
		FailureReason:   p.FailureReason,
		Metadata:        p.Metadata,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func fromDomainMapper(p *paymentDomain.Payment) *Payment {
	return &Payment{
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
		Metadata:        p.Metadata,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func arrayToDomainMapper(payments *[]Payment) *[]paymentDomain.Payment {
	paymentsDomain := make([]paymentDomain.Payment, len(*payments))
	for i, payment := range *payments {
		paymentsDomain[i] = *payment.toDomainMapper()
	}
	return &paymentsDomain
}
