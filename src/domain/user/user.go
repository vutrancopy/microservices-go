package user

import (
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
)

type User struct {
	ID            int
	UserName      string
	Email         string
	FirstName     string
	LastName      string
	Status        bool
	HashPassword  string
	Password      string
	Role          common.UserRole
	PhoneNumber   string
	ProfileImage  string
	Rating        float64 // Average rating (for drivers)
	TotalRides    int     // Total completed rides
	Latitude      float64 // Current latitude (for real-time location)
	Longitude     float64 // Current longitude (for real-time location)
	LastLocation  *time.Time
	IsAvailable   bool      // For drivers: available to accept rides
	LicenseNumber string    // For drivers
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IsDriver checks if the user is a driver
func (u *User) IsDriver() bool {
	return u.Role == common.RoleDriver
}

// IsRider checks if the user is a rider
func (u *User) IsRider() bool {
	return u.Role == common.RoleRider
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == common.RoleAdmin
}

// CanAcceptRides checks if a driver can accept rides
func (u *User) CanAcceptRides() bool {
	return u.IsDriver() && u.IsAvailable && u.Status
}

// UpdateLocation updates the user's current location
func (u *User) UpdateLocation(lat, lng float64) {
	u.Latitude = lat
	u.Longitude = lng
	now := time.Now()
	u.LastLocation = &now
}

// UpdateRating updates the user's rating (for drivers)
func (u *User) UpdateRating(newRating float64, totalRatings int) {
	if totalRatings == 0 {
		u.Rating = newRating
		return
	}
	// Calculate running average
	u.Rating = ((u.Rating * float64(totalRatings)) + newRating) / float64(totalRatings+1)
}

type SearchResultUser struct {
	Data       *[]User
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

type IUserService interface {
	GetAll() (*[]User, error)
	GetByID(id int) (*User, error)
	Create(newUser *User) (*User, error)
	Delete(id int) error
	Update(id int, userMap map[string]interface{}) (*User, error)
	SearchPaginated(filters domain.DataFilters) (*SearchResultUser, error)
	SearchByProperty(property string, searchText string) (*[]string, error)
}
