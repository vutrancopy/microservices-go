package vehicle

import (
	"testing"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain/common"
)

func TestVehicle_IsActive(t *testing.T) {
	tests := []struct {
		name    string
		vehicle *Vehicle
		want    bool
	}{
		{
			name: "Active Vehicle",
			vehicle: &Vehicle{
				Status: common.VehicleStatusActive,
			},
			want: true,
		},
		{
			name: "Inactive Vehicle",
			vehicle: &Vehicle{
				Status: common.VehicleStatusInactive,
			},
			want: false,
		},
		{
			name: "Vehicle in Inspection",
			vehicle: &Vehicle{
				Status: common.VehicleStatusInspection,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vehicle.IsActive(); got != tt.want {
				t.Errorf("Vehicle.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVehicle_IsInsuranceValid(t *testing.T) {
	futureDate := time.Now().Add(30 * 24 * time.Hour)
	pastDate := time.Now().Add(-30 * 24 * time.Hour)

	tests := []struct {
		name    string
		vehicle *Vehicle
		want    bool
	}{
		{
			name: "Valid Insurance",
			vehicle: &Vehicle{
				InsuranceExp: &futureDate,
			},
			want: true,
		},
		{
			name: "Expired Insurance",
			vehicle: &Vehicle{
				InsuranceExp: &pastDate,
			},
			want: false,
		},
		{
			name: "No Insurance Date",
			vehicle: &Vehicle{
				InsuranceExp: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vehicle.IsInsuranceValid(); got != tt.want {
				t.Errorf("Vehicle.IsInsuranceValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVehicle_IsRegistrationValid(t *testing.T) {
	futureDate := time.Now().Add(30 * 24 * time.Hour)
	pastDate := time.Now().Add(-30 * 24 * time.Hour)

	tests := []struct {
		name    string
		vehicle *Vehicle
		want    bool
	}{
		{
			name: "Valid Registration",
			vehicle: &Vehicle{
				RegistrationExp: &futureDate,
			},
			want: true,
		},
		{
			name: "Expired Registration",
			vehicle: &Vehicle{
				RegistrationExp: &pastDate,
			},
			want: false,
		},
		{
			name: "No Registration Date",
			vehicle: &Vehicle{
				RegistrationExp: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vehicle.IsRegistrationValid(); got != tt.want {
				t.Errorf("Vehicle.IsRegistrationValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVehicle_CanAcceptRides(t *testing.T) {
	futureDate := time.Now().Add(30 * 24 * time.Hour)
	pastDate := time.Now().Add(-30 * 24 * time.Hour)

	tests := []struct {
		name    string
		vehicle *Vehicle
		want    bool
	}{
		{
			name: "Vehicle Can Accept Rides",
			vehicle: &Vehicle{
				Status:          common.VehicleStatusActive,
				InsuranceExp:    &futureDate,
				RegistrationExp: &futureDate,
			},
			want: true,
		},
		{
			name: "Inactive Vehicle",
			vehicle: &Vehicle{
				Status:          common.VehicleStatusInactive,
				InsuranceExp:    &futureDate,
				RegistrationExp: &futureDate,
			},
			want: false,
		},
		{
			name: "Expired Insurance",
			vehicle: &Vehicle{
				Status:          common.VehicleStatusActive,
				InsuranceExp:    &pastDate,
				RegistrationExp: &futureDate,
			},
			want: false,
		},
		{
			name: "Expired Registration",
			vehicle: &Vehicle{
				Status:          common.VehicleStatusActive,
				InsuranceExp:    &futureDate,
				RegistrationExp: &pastDate,
			},
			want: false,
		},
		{
			name: "All Expired",
			vehicle: &Vehicle{
				Status:          common.VehicleStatusInactive,
				InsuranceExp:    &pastDate,
				RegistrationExp: &pastDate,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vehicle.CanAcceptRides(); got != tt.want {
				t.Errorf("Vehicle.CanAcceptRides() = %v, want %v", got, tt.want)
			}
		})
	}
}
