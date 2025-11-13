package user

import (
	"testing"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain/common"
)

func TestUser_IsDriver(t *testing.T) {
	tests := []struct {
		name string
		role common.UserRole
		want bool
	}{
		{"User is driver", common.RoleDriver, true},
		{"User is rider", common.RoleRider, false},
		{"User is admin", common.RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.IsDriver(); got != tt.want {
				t.Errorf("IsDriver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsRider(t *testing.T) {
	tests := []struct {
		name string
		role common.UserRole
		want bool
	}{
		{"User is rider", common.RoleRider, true},
		{"User is driver", common.RoleDriver, false},
		{"User is admin", common.RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.IsRider(); got != tt.want {
				t.Errorf("IsRider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role common.UserRole
		want bool
	}{
		{"User is admin", common.RoleAdmin, true},
		{"User is driver", common.RoleDriver, false},
		{"User is rider", common.RoleRider, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanAcceptRides(t *testing.T) {
	tests := []struct {
		name        string
		role        common.UserRole
		isAvailable bool
		status      bool
		want        bool
	}{
		{"Available active driver can accept", common.RoleDriver, true, true, true},
		{"Unavailable driver cannot accept", common.RoleDriver, false, true, false},
		{"Inactive driver cannot accept", common.RoleDriver, true, false, false},
		{"Rider cannot accept rides", common.RoleRider, true, true, false},
		{"Admin cannot accept rides", common.RoleAdmin, true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Role:        tt.role,
				IsAvailable: tt.isAvailable,
				Status:      tt.status,
			}
			if got := user.CanAcceptRides(); got != tt.want {
				t.Errorf("CanAcceptRides() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_UpdateLocation(t *testing.T) {
	user := &User{}
	lat := 37.7749
	lng := -122.4194

	user.UpdateLocation(lat, lng)

	if user.Latitude != lat {
		t.Errorf("Expected latitude %f, got %f", lat, user.Latitude)
	}
	if user.Longitude != lng {
		t.Errorf("Expected longitude %f, got %f", lng, user.Longitude)
	}
	if user.LastLocation == nil {
		t.Error("LastLocation should be set")
	}
	if user.LastLocation.After(time.Now()) {
		t.Error("LastLocation should not be in the future")
	}
}

func TestUser_UpdateRating(t *testing.T) {
	tests := []struct {
		name         string
		currentRating float64
		totalRatings int
		newRating    float64
		wantRating   float64
	}{
		{
			name:         "First rating",
			currentRating: 0.0,
			totalRatings: 0,
			newRating:    5.0,
			wantRating:   5.0,
		},
		{
			name:         "Second rating",
			currentRating: 5.0,
			totalRatings: 1,
			newRating:    3.0,
			wantRating:   4.0, // (5.0 * 1 + 3.0) / 2 = 4.0
		},
		{
			name:         "Third rating",
			currentRating: 4.0,
			totalRatings: 2,
			newRating:    5.0,
			wantRating:   4.333333333333333, // (4.0 * 2 + 5.0) / 3 = 4.333...
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Rating: tt.currentRating}
			user.UpdateRating(tt.newRating, tt.totalRatings)
			if user.Rating != tt.wantRating {
				t.Errorf("UpdateRating() rating = %v, want %v", user.Rating, tt.wantRating)
			}
		})
	}
}
