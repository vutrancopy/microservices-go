package common

import (
	"testing"
)

func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{"Valid Rider", RoleRider, true},
		{"Valid Driver", RoleDriver, true},
		{"Valid Admin", RoleAdmin, true},
		{"Invalid Role", UserRole("invalid"), false},
		{"Empty Role", UserRole(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.want {
				t.Errorf("UserRole.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRideStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status RideStatus
		want   bool
	}{
		{"Valid Pending", RideStatusPending, true},
		{"Valid Matched", RideStatusMatched, true},
		{"Valid InProgress", RideStatusInProgress, true},
		{"Valid Completed", RideStatusCompleted, true},
		{"Valid Cancelled", RideStatusCancelled, true},
		{"Invalid Status", RideStatus("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("RideStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRideStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name      string
		current   RideStatus
		next      RideStatus
		wantValid bool
	}{
		// Valid transitions
		{"Pending to Matched", RideStatusPending, RideStatusMatched, true},
		{"Pending to Cancelled", RideStatusPending, RideStatusCancelled, true},
		{"Matched to InProgress", RideStatusMatched, RideStatusInProgress, true},
		{"Matched to Cancelled", RideStatusMatched, RideStatusCancelled, true},
		{"InProgress to Completed", RideStatusInProgress, RideStatusCompleted, true},
		{"InProgress to Cancelled", RideStatusInProgress, RideStatusCancelled, true},

		// Invalid transitions
		{"Pending to InProgress", RideStatusPending, RideStatusInProgress, false},
		{"Pending to Completed", RideStatusPending, RideStatusCompleted, false},
		{"Matched to Completed", RideStatusMatched, RideStatusCompleted, false},
		{"Completed to Pending", RideStatusCompleted, RideStatusPending, false},
		{"Completed to Cancelled", RideStatusCompleted, RideStatusCancelled, false},
		{"Cancelled to Pending", RideStatusCancelled, RideStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.current.CanTransitionTo(tt.next); got != tt.wantValid {
				t.Errorf("RideStatus.CanTransitionTo() = %v, want %v for %s -> %s",
					got, tt.wantValid, tt.current, tt.next)
			}
		})
	}
}

func TestPaymentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status PaymentStatus
		want   bool
	}{
		{"Valid Pending", PaymentStatusPending, true},
		{"Valid Completed", PaymentStatusCompleted, true},
		{"Valid Failed", PaymentStatusFailed, true},
		{"Valid Refunded", PaymentStatusRefunded, true},
		{"Invalid Status", PaymentStatus("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("PaymentStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaymentMethod_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		method PaymentMethod
		want   bool
	}{
		{"Valid Cash", PaymentMethodCash, true},
		{"Valid Credit Card", PaymentMethodCreditCard, true},
		{"Valid Debit Card", PaymentMethodDebitCard, true},
		{"Valid Digital Wallet", PaymentMethodDigitalWallet, true},
		{"Invalid Method", PaymentMethod("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method.IsValid(); got != tt.want {
				t.Errorf("PaymentMethod.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVehicleType_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		vehicleType VehicleType
		want        bool
	}{
		{"Valid Sedan", VehicleTypeSedan, true},
		{"Valid SUV", VehicleTypeSUV, true},
		{"Valid Van", VehicleTypeVan, true},
		{"Valid Bike", VehicleTypeBike, true},
		{"Invalid Type", VehicleType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vehicleType.IsValid(); got != tt.want {
				t.Errorf("VehicleType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVehicleStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status VehicleStatus
		want   bool
	}{
		{"Valid Active", VehicleStatusActive, true},
		{"Valid Inactive", VehicleStatusInactive, true},
		{"Valid Inspection", VehicleStatusInspection, true},
		{"Invalid Status", VehicleStatus("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("VehicleStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
