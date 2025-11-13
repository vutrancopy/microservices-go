package location

import (
	"errors"
	"time"
)

// LocationType defines type of saved location
type LocationType string

const (
	LocationTypeHome  LocationType = "home"
	LocationTypeWork  LocationType = "work"
	LocationTypeOther LocationType = "other"
)

// FavoriteLocation represents a user's saved location
type FavoriteLocation struct {
	ID        int
	UserID    int
	Label     string       // e.g., "Home", "Office", "Mom's House"
	Type      LocationType
	Address   string
	Latitude  float64
	Longitude float64
	IsPrimary bool // Is this the primary location for this type
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate validates the favorite location
func (fl *FavoriteLocation) Validate() error {
	if fl.UserID == 0 {
		return errors.New("user ID is required")
	}

	if fl.Label == "" {
		return errors.New("label is required")
	}

	if fl.Address == "" {
		return errors.New("address is required")
	}

	if fl.Latitude < -90 || fl.Latitude > 90 {
		return errors.New("invalid latitude")
	}

	if fl.Longitude < -180 || fl.Longitude > 180 {
		return errors.New("invalid longitude")
	}

	// Validate type
	switch fl.Type {
	case LocationTypeHome, LocationTypeWork, LocationTypeOther:
		// valid
	default:
		return errors.New("invalid location type")
	}

	return nil
}

// SearchResultLocation represents paginated favorite locations
type SearchResultLocation struct {
	Data       *[]FavoriteLocation
	Total      int64
	Page       int64
	PageSize   int64
	TotalPages int64
}

// ILocationService defines the interface for location operations
type ILocationService interface {
	CreateFavorite(location *FavoriteLocation) (*FavoriteLocation, error)
	GetUserFavorites(userID int) (*[]FavoriteLocation, error)
	GetByID(id int) (*FavoriteLocation, error)
	Update(id int, updates map[string]interface{}) (*FavoriteLocation, error)
	Delete(id int) error
	SetAsPrimary(id int) error
}
