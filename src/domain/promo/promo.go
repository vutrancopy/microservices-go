package promo

import (
	"errors"
	"strings"
	"time"
)

// PromoType defines the type of promotion
type PromoType string

const (
	PromoTypePercentage PromoType = "percentage" // Discount by percentage
	PromoTypeFixed      PromoType = "fixed"      // Fixed amount discount
	PromoTypeFreeRide   PromoType = "free_ride"  // Free ride up to certain amount
)

// PromoCode represents a promotional discount code
type PromoCode struct {
	ID               int
	Code             string    // Unique promo code (e.g., "WELCOME20")
	Description      string    // User-friendly description
	Type             PromoType // Type of discount
	Value            float64   // Discount value (percentage or fixed amount)
	MaxDiscount      float64   // Maximum discount amount (for percentage type)
	MinRideAmount    float64   // Minimum ride amount to apply promo
	MaxUsagePerUser  int       // Maximum times a user can use this code
	MaxTotalUsage    int       // Maximum total usage across all users
	CurrentUsage     int       // Current total usage count
	ValidFrom        time.Time // Start date
	ValidUntil       time.Time // End date
	Active           bool      // Is promo active
	FirstRideOnly    bool      // Only for first ride
	SpecificUserID   *int      // If set, only this user can use
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PromoUsage tracks individual promo code usage
type PromoUsage struct {
	ID           int
	PromoCodeID  int
	UserID       int
	RideID       int
	DiscountAmount float64
	UsedAt       time.Time
}

// IsValid checks if the promo code is currently valid
func (p *PromoCode) IsValid() error {
	if !p.Active {
		return errors.New("promo code is not active")
	}

	now := time.Now()
	if now.Before(p.ValidFrom) {
		return errors.New("promo code is not yet valid")
	}

	if now.After(p.ValidUntil) {
		return errors.New("promo code has expired")
	}

	if p.MaxTotalUsage > 0 && p.CurrentUsage >= p.MaxTotalUsage {
		return errors.New("promo code usage limit reached")
	}

	return nil
}

// CanUserUse checks if a specific user can use this promo code
func (p *PromoCode) CanUserUse(userID int, userUsageCount int, isFirstRide bool) error {
	if err := p.IsValid(); err != nil {
		return err
	}

	// Check if promo is for specific user only
	if p.SpecificUserID != nil && *p.SpecificUserID != userID {
		return errors.New("promo code is not valid for this user")
	}

	// Check first ride restriction
	if p.FirstRideOnly && !isFirstRide {
		return errors.New("promo code is only valid for first ride")
	}

	// Check per-user usage limit
	if p.MaxUsagePerUser > 0 && userUsageCount >= p.MaxUsagePerUser {
		return errors.New("you have reached the usage limit for this promo code")
	}

	return nil
}

// CalculateDiscount calculates the discount amount for a given ride fare
func (p *PromoCode) CalculateDiscount(rideFare float64) (float64, error) {
	if rideFare < p.MinRideAmount {
		return 0, errors.New("ride amount does not meet minimum requirement")
	}

	var discount float64

	switch p.Type {
	case PromoTypePercentage:
		discount = rideFare * (p.Value / 100.0)
		if p.MaxDiscount > 0 && discount > p.MaxDiscount {
			discount = p.MaxDiscount
		}

	case PromoTypeFixed:
		discount = p.Value
		if discount > rideFare {
			discount = rideFare // Cannot discount more than ride fare
		}

	case PromoTypeFreeRide:
		discount = rideFare
		if discount > p.Value {
			discount = p.Value // Max free ride amount
		}

	default:
		return 0, errors.New("invalid promo type")
	}

	return discount, nil
}

// NormalizeCode normalizes promo code to uppercase and removes spaces
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// IPromoService defines the interface for promo code operations
type IPromoService interface {
	GetByCode(code string) (*PromoCode, error)
	ValidatePromoCode(code string, userID int, rideFare float64) (*PromoCode, float64, error)
	ApplyPromoCode(promoCodeID int, userID int, rideID int, discountAmount float64) error
	GetUserPromoUsage(userID int, promoCodeID int) (int, error)
	IsUserFirstRide(userID int) (bool, error)
	CreatePromoCode(promo *PromoCode) (*PromoCode, error)
	UpdatePromoCode(id int, updates map[string]interface{}) (*PromoCode, error)
	DeactivatePromoCode(id int) error
	GetAllActivePromos() (*[]PromoCode, error)
	GetPromoUsageHistory(userID int) (*[]PromoUsage, error)
}

// Common promo code templates
var CommonPromoCodes = []PromoCode{
	{
		Code:            "WELCOME20",
		Description:     "20% off your first ride",
		Type:            PromoTypePercentage,
		Value:           20.0,
		MaxDiscount:     10.0,
		MinRideAmount:   5.0,
		MaxUsagePerUser: 1,
		MaxTotalUsage:   1000,
		FirstRideOnly:   true,
		Active:          true,
	},
	{
		Code:            "SAVE5",
		Description:     "$5 off any ride",
		Type:            PromoTypeFixed,
		Value:           5.0,
		MinRideAmount:   10.0,
		MaxUsagePerUser: 3,
		MaxTotalUsage:   500,
		FirstRideOnly:   false,
		Active:          true,
	},
	{
		Code:            "FREERIDE10",
		Description:     "Free ride up to $10",
		Type:            PromoTypeFreeRide,
		Value:           10.0,
		MinRideAmount:   0.0,
		MaxUsagePerUser: 1,
		MaxTotalUsage:   100,
		FirstRideOnly:   true,
		Active:          true,
	},
}
