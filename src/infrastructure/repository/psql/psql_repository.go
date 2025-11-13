package psql

import (
	"fmt"
	"os"
	"strings"

	"github.com/gbrayhan/microservices-go/src/domain/common"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/medicine"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/payment"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/ride"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/user"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/vehicle"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// loadDatabaseConfig loads database configuration from environment variables
// Returns error if any required environment variable is missing
func loadDatabaseConfig() (DatabaseConfig, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")

	// Check for missing required environment variables
	var missingVars []string
	if host == "" {
		missingVars = append(missingVars, "DB_HOST")
	}
	if port == "" {
		missingVars = append(missingVars, "DB_PORT")
	}
	if user == "" {
		missingVars = append(missingVars, "DB_USER")
	}
	if password == "" {
		missingVars = append(missingVars, "DB_PASSWORD")
	}
	if dbName == "" {
		missingVars = append(missingVars, "DB_NAME")
	}
	if sslMode == "" {
		missingVars = append(missingVars, "DB_SSLMODE")
	}

	if len(missingVars) > 0 {
		return DatabaseConfig{}, fmt.Errorf("missing required database environment variables: %s", strings.Join(missingVars, ", "))
	}

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}, nil
}

type PSQLRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
	Auth   AuthService
}

type AuthService interface {
	HashPassword(password string) (string, error)
}

func NewRepository(db *gorm.DB, loggerInstance *logger.Logger) *PSQLRepository {
	return &PSQLRepository{
		DB:     db,
		Logger: loggerInstance,
	}
}

func (r *PSQLRepository) SetLogger(loggerInstance *logger.Logger) {
	r.Logger = loggerInstance
}

func (r *PSQLRepository) SetAuthService(auth AuthService) {
	r.Auth = auth
}

func (r *PSQLRepository) LoadDBConfig() (DatabaseConfig, error) {
	return loadDatabaseConfig()
}

func (c DatabaseConfig) GetDSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode +
		" TimeZone=America/Mexico_City"
}

func (r *PSQLRepository) InitDatabase() error {
	cfg, err := loadDatabaseConfig()
	if err != nil {
		r.Logger.Error("Failed to load database configuration", zap.Error(err))
		return fmt.Errorf("failed to load database configuration: %w", err)
	}

	// Create GORM logger with zap
	gormZap := logger.NewGormLogger(r.Logger.Log).
		LogMode(gormlogger.Warn) // Silent / Error / Warn / Info

	r.DB, err = gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormZap,
	})
	if err != nil {
		r.Logger.Error("Error connecting to the database", zap.Error(err))
		return err
	}

	err = r.MigrateEntitiesGORM()
	if err != nil {
		r.Logger.Error("Error migrating the database", zap.Error(err))
		return err
	}

	err = r.SeedInitialUser()
	if err != nil {
		r.Logger.Error("Error seeding initial user", zap.Error(err))
		return err
	}

	r.Logger.Info("Database connection and migrations successful")
	return nil
}

func (r *PSQLRepository) MigrateEntitiesGORM() error {
	r.Logger.Info("Starting database migration...")

	// Import the models to register them with GORM
	userModel := &user.User{}
	medicineModel := &medicine.Medicine{}
	vehicleModel := &vehicle.Vehicle{}
	rideModel := &ride.Ride{}
	paymentModel := &payment.Payment{}

	// Auto migrate the models to create/update tables
	// Order matters: users first (referenced by other tables)
	err := r.DB.AutoMigrate(
		userModel,
		medicineModel,
		vehicleModel,
		rideModel,
		paymentModel,
	)
	if err != nil {
		r.Logger.Error("Error migrating database entities", zap.Error(err))
		return err
	}

	// Add indexes for performance (if not exists)
	r.Logger.Info("Adding database indexes...")

	// User indexes
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_is_available ON users(is_available)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_location ON users(latitude, longitude)").Error

	// Vehicle indexes
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_vehicles_driver_id ON vehicles(driver_id)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(status)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_vehicles_type ON vehicles(vehicle_type)").Error

	// Ride indexes
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_rides_rider_id ON rides(rider_id)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_rides_driver_id ON rides(driver_id)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_rides_status ON rides(status)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_rides_created_at ON rides(created_at)").Error

	// Payment indexes
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_payments_ride_id ON payments(ride_id)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_payments_rider_id ON payments(rider_id)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_payments_driver_id ON payments(driver_id)").Error
	_ = r.DB.Exec("CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)").Error

	r.Logger.Info("Database entities migration and indexing completed successfully")
	return nil
}

func (r *PSQLRepository) SeedInitialUser() error {
	email := os.Getenv("START_USER_EMAIL")
	pw := os.Getenv("START_USER_PW")
	if email == "" || pw == "" {
		r.Logger.Info("Initial user seed skipped: START_USER_EMAIL or START_USER_PW not set")
		return nil
	}

	// Check if user already exists
	var existingUser user.User
	err := r.DB.Where("email = ?", email).First(&existingUser).Error
	if err == nil {
		r.Logger.Info("Initial admin user already exists, skipping seed", zap.String("email", email))
		return nil
	}

	// Create initial admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		r.Logger.Error("Error hashing password for initial user", zap.Error(err))
		return err
	}

	newUser := user.User{
		UserName:     "admin",
		Email:        email,
		FirstName:    "System",
		LastName:     "Administrator",
		HashPassword: string(hashedPassword),
		Role:         string(common.RoleAdmin),
		Status:       true,
		Rating:       5.0,
		TotalRides:   0,
		IsAvailable:  false,
	}

	err = r.DB.Create(&newUser).Error
	if err != nil {
		r.Logger.Error("Error creating initial admin user", zap.Error(err))
		return err
	}

	r.Logger.Info("Initial admin user created successfully",
		zap.String("email", email),
		zap.String("role", string(common.RoleAdmin)))

	// Seed sample driver
	r.seedSampleDriver()

	// Seed sample rider
	r.seedSampleRider()

	return nil
}

func (r *PSQLRepository) seedSampleDriver() {
	// Check if sample driver exists
	var existingDriver user.User
	driverEmail := "driver.sample@ridehailing.com"
	err := r.DB.Where("email = ?", driverEmail).First(&existingDriver).Error
	if err == nil {
		return // Already exists
	}

	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("Driver@123"), bcrypt.DefaultCost)
	sampleDriver := user.User{
		UserName:      "driver_sample",
		Email:         driverEmail,
		FirstName:     "John",
		LastName:      "Driver",
		HashPassword:  string(hashedPw),
		Role:          string(common.RoleDriver),
		Status:        true,
		PhoneNumber:   "+1234567890",
		Rating:        4.8,
		TotalRides:    150,
		IsAvailable:   true,
		LicenseNumber: "DL123456789",
		Latitude:      37.7749,
		Longitude:     -122.4194,
	}

	if err := r.DB.Create(&sampleDriver).Error; err == nil {
		r.Logger.Info("Sample driver created", zap.String("email", driverEmail))
	}
}

func (r *PSQLRepository) seedSampleRider() {
	// Check if sample rider exists
	var existingRider user.User
	riderEmail := "rider.sample@ridehailing.com"
	err := r.DB.Where("email = ?", riderEmail).First(&existingRider).Error
	if err == nil {
		return // Already exists
	}

	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("Rider@123"), bcrypt.DefaultCost)
	sampleRider := user.User{
		UserName:    "rider_sample",
		Email:       riderEmail,
		FirstName:   "Jane",
		LastName:    "Rider",
		HashPassword: string(hashedPw),
		Role:        string(common.RoleRider),
		Status:      true,
		PhoneNumber: "+1987654321",
		Rating:      4.9,
		TotalRides:  80,
		IsAvailable: false,
		Latitude:    37.7849,
		Longitude:   -122.4094,
	}

	if err := r.DB.Create(&sampleRider).Error; err == nil {
		r.Logger.Info("Sample rider created", zap.String("email", riderEmail))
	}
}

// InitPSQLDB initializes the database connection with logger
func InitPSQLDB(loggerInstance *logger.Logger) (*gorm.DB, error) {
	repo := &PSQLRepository{
		Logger: loggerInstance,
	}

	err := repo.InitDatabase()
	if err != nil {
		return nil, err
	}

	return repo.DB, nil
}
