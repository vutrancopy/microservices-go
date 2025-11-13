package pricing

import (
	"math"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain/common"
)

// PricingConfig holds pricing configuration for different vehicle types
type PricingConfig struct {
	ID              int
	VehicleType     common.VehicleType
	BaseFare        float64 // Base fare per ride
	CostPerKm       float64 // Cost per kilometer
	CostPerMinute   float64 // Cost per minute
	MinimumFare     float64 // Minimum fare for any ride
	ServiceFee      float64 // Platform service fee (percentage)
	BookingFee      float64 // Fixed booking fee
	CancellationFee float64 // Cancellation fee
	WaitingTimeRate float64 // Cost per minute of waiting
	SurgeMultiplier float64 // Surge pricing multiplier (1.0 = no surge)
	Currency        string
	Active          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// FareBreakdown represents detailed fare calculation
type FareBreakdown struct {
	BaseFare        float64 `json:"baseFare"`
	DistanceFare    float64 `json:"distanceFare"`    // distance * costPerKm
	TimeFare        float64 `json:"timeFare"`        // duration * costPerMinute
	WaitingTimeFare float64 `json:"waitingTimeFare"` // waitingTime * waitingTimeRate
	BookingFee      float64 `json:"bookingFee"`
	ServiceFee      float64 `json:"serviceFee"` // percentage of subtotal
	SurgeCharge     float64 `json:"surgeCharge"`
	Discount        float64 `json:"discount"` // from promo code
	Subtotal        float64 `json:"subtotal"`
	Total           float64 `json:"total"`
	Currency        string  `json:"currency"`
}

// FareEstimateRequest represents a fare estimation request
type FareEstimateRequest struct {
	VehicleType     common.VehicleType
	DistanceKm      float64
	EstimatedMinutes int
	WaitingMinutes  int
	SurgeMultiplier float64
	PromoCode       string
}

// CalculateFare calculates the total fare based on pricing config and ride details
func (pc *PricingConfig) CalculateFare(
	distanceKm float64,
	durationMinutes int,
	waitingMinutes int,
	promoDiscount float64,
) *FareBreakdown {
	// Calculate each component
	baseFare := pc.BaseFare
	distanceFare := distanceKm * pc.CostPerKm
	timeFare := float64(durationMinutes) * pc.CostPerMinute
	waitingTimeFare := float64(waitingMinutes) * pc.WaitingTimeRate
	bookingFee := pc.BookingFee

	// Calculate subtotal (before service fee and surge)
	subtotal := baseFare + distanceFare + timeFare + waitingTimeFare + bookingFee

	// Apply surge multiplier if any
	surgeCharge := 0.0
	if pc.SurgeMultiplier > 1.0 {
		surgeCharge = subtotal * (pc.SurgeMultiplier - 1.0)
		subtotal += surgeCharge
	}

	// Calculate service fee (percentage of subtotal)
	serviceFee := subtotal * (pc.ServiceFee / 100.0)
	subtotal += serviceFee

	// Apply discount
	discount := promoDiscount
	if discount > subtotal {
		discount = subtotal // Cannot discount more than total
	}

	// Calculate final total
	total := subtotal - discount

	// Apply minimum fare
	if total < pc.MinimumFare {
		total = pc.MinimumFare
		// Recalculate components proportionally
		subtotal = total
	}

	// Round to 2 decimal places
	total = math.Round(total*100) / 100

	return &FareBreakdown{
		BaseFare:        math.Round(baseFare*100) / 100,
		DistanceFare:    math.Round(distanceFare*100) / 100,
		TimeFare:        math.Round(timeFare*100) / 100,
		WaitingTimeFare: math.Round(waitingTimeFare*100) / 100,
		BookingFee:      math.Round(bookingFee*100) / 100,
		ServiceFee:      math.Round(serviceFee*100) / 100,
		SurgeCharge:     math.Round(surgeCharge*100) / 100,
		Discount:        math.Round(discount*100) / 100,
		Subtotal:        math.Round(subtotal*100) / 100,
		Total:           total,
		Currency:        pc.Currency,
	}
}

// GetCancellationFee calculates cancellation fee based on time elapsed
func (pc *PricingConfig) GetCancellationFee(minutesSinceBooking int) float64 {
	// Free cancellation within first 2 minutes
	if minutesSinceBooking <= 2 {
		return 0.0
	}

	// After driver is matched (assuming > 2 minutes)
	if minutesSinceBooking <= 5 {
		return pc.CancellationFee * 0.5 // 50% of cancellation fee
	}

	// Full cancellation fee after 5 minutes
	return pc.CancellationFee
}

// ApplySurge applies surge pricing multiplier
func (pc *PricingConfig) ApplySurge(multiplier float64) {
	if multiplier >= 1.0 {
		pc.SurgeMultiplier = multiplier
	}
}

// IPricingService defines the interface for pricing operations
type IPricingService interface {
	GetPricingByVehicleType(vehicleType common.VehicleType) (*PricingConfig, error)
	CalculateFareEstimate(request *FareEstimateRequest) (*FareBreakdown, error)
	GetCancellationFee(vehicleType common.VehicleType, minutesSinceBooking int) (float64, error)
	UpdatePricingConfig(config *PricingConfig) error
	GetAllPricingConfigs() (*[]PricingConfig, error)
}

// Default pricing configurations
var DefaultPricingConfigs = map[common.VehicleType]PricingConfig{
	common.VehicleTypeSedan: {
		VehicleType:     common.VehicleTypeSedan,
		BaseFare:        3.00,
		CostPerKm:       1.50,
		CostPerMinute:   0.30,
		MinimumFare:     5.00,
		ServiceFee:      15.0, // 15%
		BookingFee:      2.00,
		CancellationFee: 5.00,
		WaitingTimeRate: 0.50,
		SurgeMultiplier: 1.0,
		Currency:        "USD",
		Active:          true,
	},
	common.VehicleTypeSUV: {
		VehicleType:     common.VehicleTypeSUV,
		BaseFare:        5.00,
		CostPerKm:       2.00,
		CostPerMinute:   0.40,
		MinimumFare:     8.00,
		ServiceFee:      15.0,
		BookingFee:      2.50,
		CancellationFee: 7.00,
		WaitingTimeRate: 0.60,
		SurgeMultiplier: 1.0,
		Currency:        "USD",
		Active:          true,
	},
	common.VehicleTypeVan: {
		VehicleType:     common.VehicleTypeVan,
		BaseFare:        6.00,
		CostPerKm:       2.50,
		CostPerMinute:   0.50,
		MinimumFare:     10.00,
		ServiceFee:      15.0,
		BookingFee:      3.00,
		CancellationFee: 8.00,
		WaitingTimeRate: 0.70,
		SurgeMultiplier: 1.0,
		Currency:        "USD",
		Active:          true,
	},
	common.VehicleTypeBike: {
		VehicleType:     common.VehicleTypeBike,
		BaseFare:        2.00,
		CostPerKm:       1.00,
		CostPerMinute:   0.20,
		MinimumFare:     3.00,
		ServiceFee:      15.0,
		BookingFee:      1.50,
		CancellationFee: 3.00,
		WaitingTimeRate: 0.30,
		SurgeMultiplier: 1.0,
		Currency:        "USD",
		Active:          true,
	},
}
