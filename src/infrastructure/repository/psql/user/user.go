package user

import (
	"encoding/json"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainUser "github.com/gbrayhan/microservices-go/src/domain/user"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type User struct {
	ID            int        `gorm:"primaryKey"`
	UserName      string     `gorm:"column:user_name;unique"`
	Email         string     `gorm:"unique"`
	FirstName     string     `gorm:"column:first_name"`
	LastName      string     `gorm:"column:last_name"`
	Status        bool       `gorm:"column:status"`
	HashPassword  string     `gorm:"column:hash_password"`
	Role          string     `gorm:"column:role;default:'rider'"`
	PhoneNumber   string     `gorm:"column:phone_number"`
	ProfileImage  string     `gorm:"column:profile_image"`
	Rating        float64    `gorm:"column:rating;default:5.0"`
	TotalRides    int        `gorm:"column:total_rides;default:0"`
	Latitude      float64    `gorm:"column:latitude"`
	Longitude     float64    `gorm:"column:longitude"`
	LastLocation  *time.Time `gorm:"column:last_location"`
	IsAvailable   bool       `gorm:"column:is_available;default:false"`
	LicenseNumber string     `gorm:"column:license_number"`
	CreatedAt     time.Time  `gorm:"autoCreateTime:mili"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime:mili"`
}

func (User) TableName() string {
	return "users"
}

var ColumnsUserMapping = map[string]string{
	"id":            "id",
	"userName":      "user_name",
	"email":         "email",
	"firstName":     "first_name",
	"lastName":      "last_name",
	"status":        "status",
	"hashPassword":  "hash_password",
	"role":          "role",
	"phoneNumber":   "phone_number",
	"profileImage":  "profile_image",
	"rating":        "rating",
	"totalRides":    "total_rides",
	"latitude":      "latitude",
	"longitude":     "longitude",
	"lastLocation":  "last_location",
	"isAvailable":   "is_available",
	"licenseNumber": "license_number",
	"createdAt":     "created_at",
	"updatedAt":     "updated_at",
}

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	GetAll() (*[]domainUser.User, error)
	Create(userDomain *domainUser.User) (*domainUser.User, error)
	GetByID(id int) (*domainUser.User, error)
	GetByEmail(email string) (*domainUser.User, error)
	Update(id int, userMap map[string]interface{}) (*domainUser.User, error)
	Delete(id int) error
	SearchPaginated(filters domain.DataFilters) (*domainUser.SearchResultUser, error)
	SearchByProperty(property string, searchText string) (*[]string, error)
	// Ride-hailing specific methods
	GetByRole(role common.UserRole) (*[]domainUser.User, error)
	GetAvailableDrivers() (*[]domainUser.User, error)
	GetAvailableDriversNearby(latitude, longitude float64, radiusKm float64) (*[]domainUser.User, error)
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewUserRepository(db *gorm.DB, loggerInstance *logger.Logger) UserRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]domainUser.User, error) {
	var users []User
	if err := r.DB.Find(&users).Error; err != nil {
		r.Logger.Error("Error getting all users", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved all users", zap.Int("count", len(users)))
	return arrayToDomainMapper(&users), nil
}

func (r *Repository) Create(userDomain *domainUser.User) (*domainUser.User, error) {
	r.Logger.Info("Creating new user", zap.String("email", userDomain.Email))
	userRepository := fromDomainMapper(userDomain)
	txDb := r.DB.Create(userRepository)
	err := txDb.Error
	if err != nil {
		r.Logger.Error("Error creating user", zap.Error(err), zap.String("email", userDomain.Email))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		errUnmarshal := json.Unmarshal(byteErr, &newError)
		if errUnmarshal != nil {
			return &domainUser.User{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			err = domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			return &domainUser.User{}, err
		default:
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	r.Logger.Info("Successfully created user", zap.String("email", userDomain.Email), zap.Int("id", userRepository.ID))
	return userRepository.toDomainMapper(), err
}

func (r *Repository) GetByID(id int) (*domainUser.User, error) {
	var user User
	err := r.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("User not found", zap.Int("id", id))
			err = domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		} else {
			r.Logger.Error("Error getting user by ID", zap.Error(err), zap.Int("id", id))
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
		return &domainUser.User{}, err
	}
	r.Logger.Info("Successfully retrieved user by ID", zap.Int("id", id))
	return user.toDomainMapper(), nil
}

func (r *Repository) GetByEmail(email string) (*domainUser.User, error) {
	var user User
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("User not found", zap.String("email", email))
			err = domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		} else {
			r.Logger.Error("Error getting user by email", zap.Error(err), zap.String("email", email))
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
		return &domainUser.User{}, err
	}
	r.Logger.Info("Successfully retrieved user by email", zap.String("email", email))
	return user.toDomainMapper(), nil
}

func (r *Repository) Update(id int, userMap map[string]interface{}) (*domainUser.User, error) {
	var userObj User
	userObj.ID = id

	// Map JSON field names to DB column names
	updateData := make(map[string]interface{})
	for k, v := range userMap {
		if column, ok := ColumnsUserMapping[k]; ok {
			updateData[column] = v
		} else {
			updateData[k] = v
		}
	}

	err := r.DB.Model(&userObj).
		Select("user_name", "email", "first_name", "last_name", "status", "role").
		Updates(updateData).Error
	if err != nil {
		r.Logger.Error("Error updating user", zap.Error(err), zap.Int("id", id))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		errUnmarshal := json.Unmarshal(byteErr, &newError)
		if errUnmarshal != nil {
			return &domainUser.User{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			return &domainUser.User{}, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		default:
			return &domainUser.User{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	if err := r.DB.Where("id = ?", id).First(&userObj).Error; err != nil {
		r.Logger.Error("Error retrieving updated user", zap.Error(err), zap.Int("id", id))
		return &domainUser.User{}, err
	}
	r.Logger.Info("Successfully updated user", zap.Int("id", id))
	return userObj.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	tx := r.DB.Delete(&User{}, id)
	if tx.Error != nil {
		r.Logger.Error("Error deleting user", zap.Error(tx.Error), zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if tx.RowsAffected == 0 {
		r.Logger.Warn("User not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	r.Logger.Info("Successfully deleted user", zap.Int("id", id))
	return nil
}

func (r *Repository) SearchPaginated(filters domain.DataFilters) (*domainUser.SearchResultUser, error) {
	query := r.DB.Model(&User{})

	// Apply like filters
	for field, values := range filters.LikeFilters {
		if len(values) > 0 {
			for _, value := range values {
				if value != "" {
					column := ColumnsUserMapping[field]
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
			column := ColumnsUserMapping[field]
			if column != "" {
				query = query.Where(column+" IN ?", values)
			}
		}
	}

	// Apply date range filters
	for _, dateFilter := range filters.DateRangeFilters {
		column := ColumnsUserMapping[dateFilter.Field]
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
			column := ColumnsUserMapping[sortField]
			if column != "" {
				query = query.Order(column + " " + string(filters.SortDirection))
			}
		}
	}

	// Count total records
	var total int64
	clonedQuery := query
	clonedQuery.Count(&total)

	// Apply pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 10
	}
	offset := (filters.Page - 1) * filters.PageSize

	var users []User
	if err := query.Offset(offset).Limit(filters.PageSize).Find(&users).Error; err != nil {
		r.Logger.Error("Error searching users", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))

	result := &domainUser.SearchResultUser{
		Data:       arrayToDomainMapper(&users),
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}

	r.Logger.Info("Successfully searched users",
		zap.Int64("total", total),
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))

	return result, nil
}

func (r *Repository) SearchByProperty(property string, searchText string) (*[]string, error) {
	column := ColumnsUserMapping[property]
	if column == "" {
		r.Logger.Warn("Invalid property for search", zap.String("property", property))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	var coincidences []string
	if err := r.DB.Model(&User{}).
		Distinct(column).
		Where(column+" ILIKE ?", "%"+searchText+"%").
		Limit(20).
		Pluck(column, &coincidences).Error; err != nil {
		r.Logger.Error("Error searching by property", zap.Error(err), zap.String("property", property))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	r.Logger.Info("Successfully searched by property",
		zap.String("property", property),
		zap.Int("results", len(coincidences)))

	return &coincidences, nil
}

// GetByRole retrieves users by role
func (r *Repository) GetByRole(role common.UserRole) (*[]domainUser.User, error) {
	var users []User
	if err := r.DB.Where("role = ?", string(role)).Find(&users).Error; err != nil {
		r.Logger.Error("Error getting users by role", zap.Error(err), zap.String("role", string(role)))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved users by role",
		zap.String("role", string(role)),
		zap.Int("count", len(users)))
	return arrayToDomainMapper(&users), nil
}

// GetAvailableDrivers retrieves all available drivers
func (r *Repository) GetAvailableDrivers() (*[]domainUser.User, error) {
	var users []User
	if err := r.DB.Where("role = ? AND is_available = ? AND status = ?",
		string(common.RoleDriver), true, true).Find(&users).Error; err != nil {
		r.Logger.Error("Error getting available drivers", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved available drivers", zap.Int("count", len(users)))
	return arrayToDomainMapper(&users), nil
}

// GetAvailableDriversNearby retrieves available drivers within radius
func (r *Repository) GetAvailableDriversNearby(latitude, longitude float64, radiusKm float64) (*[]domainUser.User, error) {
	var users []User
	// Using Haversine formula for distance calculation
	// Note: This is a simplified version. For production, consider using PostGIS
	query := `
		SELECT * FROM users
		WHERE role = ?
		AND is_available = ?
		AND status = ?
		AND (
			6371 * acos(
				cos(radians(?)) * cos(radians(latitude)) *
				cos(radians(longitude) - radians(?)) +
				sin(radians(?)) * sin(radians(latitude))
			)
		) <= ?
		ORDER BY (
			6371 * acos(
				cos(radians(?)) * cos(radians(latitude)) *
				cos(radians(longitude) - radians(?)) +
				sin(radians(?)) * sin(radians(latitude))
			)
		) ASC
		LIMIT 10
	`
	if err := r.DB.Raw(query,
		string(common.RoleDriver), true, true,
		latitude, longitude, latitude, radiusKm,
		latitude, longitude, latitude,
	).Scan(&users).Error; err != nil {
		r.Logger.Error("Error getting nearby drivers", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved nearby drivers",
		zap.Int("count", len(users)),
		zap.Float64("latitude", latitude),
		zap.Float64("longitude", longitude),
		zap.Float64("radiusKm", radiusKm))
	return arrayToDomainMapper(&users), nil
}

// Mappers
func (u *User) toDomainMapper() *domainUser.User {
	return &domainUser.User{
		ID:            u.ID,
		UserName:      u.UserName,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Status:        u.Status,
		HashPassword:  u.HashPassword,
		Role:          common.UserRole(u.Role),
		PhoneNumber:   u.PhoneNumber,
		ProfileImage:  u.ProfileImage,
		Rating:        u.Rating,
		TotalRides:    u.TotalRides,
		Latitude:      u.Latitude,
		Longitude:     u.Longitude,
		LastLocation:  u.LastLocation,
		IsAvailable:   u.IsAvailable,
		LicenseNumber: u.LicenseNumber,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func fromDomainMapper(u *domainUser.User) *User {
	return &User{
		ID:            u.ID,
		UserName:      u.UserName,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Status:        u.Status,
		HashPassword:  u.HashPassword,
		Role:          string(u.Role),
		PhoneNumber:   u.PhoneNumber,
		ProfileImage:  u.ProfileImage,
		Rating:        u.Rating,
		TotalRides:    u.TotalRides,
		Latitude:      u.Latitude,
		Longitude:     u.Longitude,
		LastLocation:  u.LastLocation,
		IsAvailable:   u.IsAvailable,
		LicenseNumber: u.LicenseNumber,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func arrayToDomainMapper(users *[]User) *[]domainUser.User {
	usersDomain := make([]domainUser.User, len(*users))
	for i, user := range *users {
		usersDomain[i] = *user.toDomainMapper()
	}
	return &usersDomain
}
