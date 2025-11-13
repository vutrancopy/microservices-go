package ride

import (
	"testing"

	"github.com/gbrayhan/microservices-go/src/domain/common"
)

func TestRide_IsPending(t *testing.T) {
	ride := &Ride{Status: common.RideStatusPending}
	if !ride.IsPending() {
		t.Error("Expected ride to be pending")
	}
}

func TestRide_IsMatched(t *testing.T) {
	ride := &Ride{Status: common.RideStatusMatched}
	if !ride.IsMatched() {
		t.Error("Expected ride to be matched")
	}
}

func TestRide_IsInProgress(t *testing.T) {
	ride := &Ride{Status: common.RideStatusInProgress}
	if !ride.IsInProgress() {
		t.Error("Expected ride to be in progress")
	}
}

func TestRide_IsCompleted(t *testing.T) {
	ride := &Ride{Status: common.RideStatusCompleted}
	if !ride.IsCompleted() {
		t.Error("Expected ride to be completed")
	}
}

func TestRide_IsCancelled(t *testing.T) {
	ride := &Ride{Status: common.RideStatusCancelled}
	if !ride.IsCancelled() {
		t.Error("Expected ride to be cancelled")
	}
}

func TestRide_CanBeCancelled(t *testing.T) {
	tests := []struct {
		name   string
		status common.RideStatus
		want   bool
	}{
		{"Pending can be cancelled", common.RideStatusPending, true},
		{"Matched can be cancelled", common.RideStatusMatched, true},
		{"InProgress cannot be cancelled", common.RideStatusInProgress, false},
		{"Completed cannot be cancelled", common.RideStatusCompleted, false},
		{"Already cancelled", common.RideStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			if got := ride.CanBeCancelled(); got != tt.want {
				t.Errorf("CanBeCancelled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRide_CanStart(t *testing.T) {
	tests := []struct {
		name   string
		status common.RideStatus
		want   bool
	}{
		{"Pending cannot start", common.RideStatusPending, false},
		{"Matched can start", common.RideStatusMatched, true},
		{"InProgress cannot start again", common.RideStatusInProgress, false},
		{"Completed cannot start", common.RideStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			if got := ride.CanStart(); got != tt.want {
				t.Errorf("CanStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRide_CanComplete(t *testing.T) {
	tests := []struct {
		name   string
		status common.RideStatus
		want   bool
	}{
		{"Pending cannot complete", common.RideStatusPending, false},
		{"Matched cannot complete", common.RideStatusMatched, false},
		{"InProgress can complete", common.RideStatusInProgress, true},
		{"Completed cannot complete again", common.RideStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			if got := ride.CanComplete(); got != tt.want {
				t.Errorf("CanComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRide_AssignDriver(t *testing.T) {
	driverID := 123
	vehicleID := 456

	tests := []struct {
		name      string
		status    common.RideStatus
		wantError bool
	}{
		{"Pending ride can be assigned", common.RideStatusPending, false},
		{"Matched ride cannot be reassigned", common.RideStatusMatched, true},
		{"InProgress ride cannot be assigned", common.RideStatusInProgress, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			err := ride.AssignDriver(driverID, vehicleID)
			if (err != nil) != tt.wantError {
				t.Errorf("AssignDriver() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil {
				if ride.DriverID == nil || *ride.DriverID != driverID {
					t.Error("DriverID not set correctly")
				}
				if ride.VehicleID == nil || *ride.VehicleID != vehicleID {
					t.Error("VehicleID not set correctly")
				}
				if ride.MatchedAt == nil {
					t.Error("MatchedAt not set")
				}
			}
		})
	}
}

func TestRide_Start(t *testing.T) {
	tests := []struct {
		name      string
		status    common.RideStatus
		wantError bool
	}{
		{"Pending ride cannot start", common.RideStatusPending, true},
		{"Matched ride can start", common.RideStatusMatched, false},
		{"InProgress ride cannot start again", common.RideStatusInProgress, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			err := ride.Start()
			if (err != nil) != tt.wantError {
				t.Errorf("Start() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil && ride.StartedAt == nil {
				t.Error("StartedAt not set")
			}
		})
	}
}

func TestRide_Complete(t *testing.T) {
	tests := []struct {
		name      string
		status    common.RideStatus
		wantError bool
	}{
		{"Pending ride cannot complete", common.RideStatusPending, true},
		{"Matched ride cannot complete", common.RideStatusMatched, true},
		{"InProgress ride can complete", common.RideStatusInProgress, false},
		{"Completed ride cannot complete again", common.RideStatusCompleted, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			err := ride.Complete(10.5, 25, 15.75)
			if (err != nil) != tt.wantError {
				t.Errorf("Complete() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil {
				if ride.CompletedAt == nil {
					t.Error("CompletedAt not set")
				}
				if ride.ActualDistance != 10.5 {
					t.Error("ActualDistance not set correctly")
				}
				if ride.ActualDuration != 25 {
					t.Error("ActualDuration not set correctly")
				}
				if ride.ActualFare != 15.75 {
					t.Error("ActualFare not set correctly")
				}
			}
		})
	}
}

func TestRide_Cancel(t *testing.T) {
	tests := []struct {
		name      string
		status    common.RideStatus
		wantError bool
	}{
		{"Pending ride can be cancelled", common.RideStatusPending, false},
		{"Matched ride can be cancelled", common.RideStatusMatched, false},
		{"InProgress ride cannot be cancelled", common.RideStatusInProgress, true},
		{"Completed ride cannot be cancelled", common.RideStatusCompleted, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			reason := "Test cancellation"
			err := ride.Cancel(reason)
			if (err != nil) != tt.wantError {
				t.Errorf("Cancel() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil {
				if ride.CancelledAt == nil {
					t.Error("CancelledAt not set")
				}
				if ride.CancellationReason != reason {
					t.Error("CancellationReason not set correctly")
				}
			}
		})
	}
}

func TestRide_SetRiderRating(t *testing.T) {
	tests := []struct {
		name      string
		status    common.RideStatus
		rating    int
		wantError bool
	}{
		{"Valid rating on completed ride", common.RideStatusCompleted, 5, false},
		{"Invalid rating too low", common.RideStatusCompleted, 0, true},
		{"Invalid rating too high", common.RideStatusCompleted, 6, true},
		{"Cannot rate non-completed ride", common.RideStatusInProgress, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			err := ride.SetRiderRating(tt.rating)
			if (err != nil) != tt.wantError {
				t.Errorf("SetRiderRating() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil && (ride.RiderRating == nil || *ride.RiderRating != tt.rating) {
				t.Error("RiderRating not set correctly")
			}
		})
	}
}

func TestRide_SetDriverRating(t *testing.T) {
	tests := []struct {
		name      string
		status    common.RideStatus
		rating    int
		wantError bool
	}{
		{"Valid rating on completed ride", common.RideStatusCompleted, 4, false},
		{"Invalid rating too low", common.RideStatusCompleted, 0, true},
		{"Invalid rating too high", common.RideStatusCompleted, 6, true},
		{"Cannot rate non-completed ride", common.RideStatusMatched, 4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &Ride{Status: tt.status}
			err := ride.SetDriverRating(tt.rating)
			if (err != nil) != tt.wantError {
				t.Errorf("SetDriverRating() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil && (ride.DriverRating == nil || *ride.DriverRating != tt.rating) {
				t.Error("DriverRating not set correctly")
			}
		})
	}
}

func TestRide_GetPickupLocation(t *testing.T) {
	ride := &Ride{
		PickupLatitude:  40.7128,
		PickupLongitude: -74.0060,
		PickupAddress:   "New York, NY",
	}

	location := ride.GetPickupLocation()
	if location.Latitude != 40.7128 {
		t.Errorf("Expected latitude 40.7128, got %f", location.Latitude)
	}
	if location.Longitude != -74.0060 {
		t.Errorf("Expected longitude -74.0060, got %f", location.Longitude)
	}
	if location.Address != "New York, NY" {
		t.Errorf("Expected address 'New York, NY', got %s", location.Address)
	}
}

func TestRide_GetDropoffLocation(t *testing.T) {
	ride := &Ride{
		DropoffLatitude:  34.0522,
		DropoffLongitude: -118.2437,
		DropoffAddress:   "Los Angeles, CA",
	}

	location := ride.GetDropoffLocation()
	if location.Latitude != 34.0522 {
		t.Errorf("Expected latitude 34.0522, got %f", location.Latitude)
	}
	if location.Longitude != -118.2437 {
		t.Errorf("Expected longitude -118.2437, got %f", location.Longitude)
	}
	if location.Address != "Los Angeles, CA" {
		t.Errorf("Expected address 'Los Angeles, CA', got %s", location.Address)
	}
}
