package user

import (
	"fmt"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gbrayhan/microservices-go/src/domain/common"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	userDomain "github.com/gbrayhan/microservices-go/src/domain/user"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/user"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type IUserUseCase interface {
	GetAll() (*[]userDomain.User, error)
	GetByID(id int) (*userDomain.User, error)
	GetByEmail(email string) (*userDomain.User, error)
	Create(newUser *userDomain.User) (*userDomain.User, error)
	Delete(id int) error
	Update(id int, userMap map[string]interface{}) (*userDomain.User, error)
	SearchPaginated(filters domain.DataFilters) (*userDomain.SearchResultUser, error)
	SearchByProperty(property string, searchText string) (*[]string, error)
	// Ride-hailing specific methods
	GetDrivers() (*[]userDomain.User, error)
	GetRiders() (*[]userDomain.User, error)
	GetAvailableDrivers() (*[]userDomain.User, error)
	UpdateLocation(userID int, latitude, longitude float64) (*userDomain.User, error)
	SetDriverAvailability(driverID int, isAvailable bool) (*userDomain.User, error)
	UpdateDriverRating(driverID int, rating float64, totalRides int) (*userDomain.User, error)
}

type UserUseCase struct {
	userRepository user.UserRepositoryInterface
	Logger         *logger.Logger
}

func NewUserUseCase(userRepository user.UserRepositoryInterface, logger *logger.Logger) IUserUseCase {
	return &UserUseCase{
		userRepository: userRepository,
		Logger:         logger,
	}
}

func (s *UserUseCase) GetAll() (*[]userDomain.User, error) {
	s.Logger.Info("Getting all users")
	return s.userRepository.GetAll()
}

func (s *UserUseCase) GetByID(id int) (*userDomain.User, error) {
	s.Logger.Info("Getting user by ID", zap.Int("id", id))
	return s.userRepository.GetByID(id)
}

func (s *UserUseCase) GetByEmail(email string) (*userDomain.User, error) {
	s.Logger.Info("Getting user by email", zap.String("email", email))
	return s.userRepository.GetByEmail(email)
}

func (s *UserUseCase) Create(newUser *userDomain.User) (*userDomain.User, error) {
	s.Logger.Info("Creating new user", zap.String("email", newUser.Email))
	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Logger.Error("Error hashing password", zap.Error(err))
		return &userDomain.User{}, err
	}
	newUser.HashPassword = string(hash)
	newUser.Status = true

	return s.userRepository.Create(newUser)
}

func (s *UserUseCase) Delete(id int) error {
	s.Logger.Info("Deleting user", zap.Int("id", id))
	return s.userRepository.Delete(id)
}

func (s *UserUseCase) Update(id int, userMap map[string]interface{}) (*userDomain.User, error) {
	s.Logger.Info("Updating user", zap.Int("id", id))
	return s.userRepository.Update(id, userMap)
}

func (s *UserUseCase) SearchPaginated(filters domain.DataFilters) (*userDomain.SearchResultUser, error) {
	s.Logger.Info("Searching users with pagination",
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))
	return s.userRepository.SearchPaginated(filters)
}

func (s *UserUseCase) SearchByProperty(property string, searchText string) (*[]string, error) {
	s.Logger.Info("Searching users by property",
		zap.String("property", property),
		zap.String("searchText", searchText))
	return s.userRepository.SearchByProperty(property, searchText)
}

// GetDrivers retrieves all drivers
func (s *UserUseCase) GetDrivers() (*[]userDomain.User, error) {
	s.Logger.Info("Getting all drivers")
	return s.userRepository.GetByRole(common.RoleDriver)
}

// GetRiders retrieves all riders
func (s *UserUseCase) GetRiders() (*[]userDomain.User, error) {
	s.Logger.Info("Getting all riders")
	return s.userRepository.GetByRole(common.RoleRider)
}

// GetAvailableDrivers retrieves all available drivers
func (s *UserUseCase) GetAvailableDrivers() (*[]userDomain.User, error) {
	s.Logger.Info("Getting available drivers")
	return s.userRepository.GetAvailableDrivers()
}

// UpdateLocation updates a user's location
func (s *UserUseCase) UpdateLocation(userID int, latitude, longitude float64) (*userDomain.User, error) {
	s.Logger.Info("Updating user location",
		zap.Int("userID", userID),
		zap.Float64("latitude", latitude),
		zap.Float64("longitude", longitude))

	// Get user to validate existence
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	// Update location using domain logic
	user.UpdateLocation(latitude, longitude)

	// Save to repository
	userMap := map[string]interface{}{
		"latitude":      latitude,
		"longitude":     longitude,
		"last_location": user.LastLocation,
	}

	return s.userRepository.Update(userID, userMap)
}

// SetDriverAvailability sets a driver's availability status
func (s *UserUseCase) SetDriverAvailability(driverID int, isAvailable bool) (*userDomain.User, error) {
	s.Logger.Info("Setting driver availability",
		zap.Int("driverID", driverID),
		zap.Bool("isAvailable", isAvailable))

	// Get user to validate it's a driver
	user, err := s.userRepository.GetByID(driverID)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	if !user.IsDriver() {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("user is not a driver"),
			domainErrors.ValidationError)
	}

	userMap := map[string]interface{}{
		"is_available": isAvailable,
	}

	return s.userRepository.Update(driverID, userMap)
}

// UpdateDriverRating updates a driver's rating
func (s *UserUseCase) UpdateDriverRating(driverID int, rating float64, totalRides int) (*userDomain.User, error) {
	s.Logger.Info("Updating driver rating",
		zap.Int("driverID", driverID),
		zap.Float64("rating", rating),
		zap.Int("totalRides", totalRides))

	// Get user to validate it's a driver
	user, err := s.userRepository.GetByID(driverID)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	if !user.IsDriver() {
		return nil, domainErrors.NewAppError(
			fmt.Errorf("user is not a driver"),
			domainErrors.ValidationError)
	}

	// Update rating using domain logic
	user.UpdateRating(rating, totalRides)

	userMap := map[string]interface{}{
		"rating":      user.Rating,
		"total_rides": user.TotalRides,
	}

	return s.userRepository.Update(driverID, userMap)
}
