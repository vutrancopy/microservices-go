package di

import (
	"sync"

	authUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/auth"
	medicineUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/medicine"
	paymentUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/payment"
	rideUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/ride"
	userUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/user"
	vehicleUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/vehicle"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/medicine"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/payment"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/ride"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/user"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/vehicle"
	authController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/auth"
	medicineController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/medicine"
	paymentController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/payment"
	rideController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/ride"
	userController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/user"
	vehicleController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/vehicle"
	"github.com/gbrayhan/microservices-go/src/infrastructure/security"
	"gorm.io/gorm"
)

// ApplicationContext holds all application dependencies and services
type ApplicationContext struct {
	// Core infrastructure
	DB         *gorm.DB
	Logger     *logger.Logger
	JWTService security.IJWTService

	// Controllers
	AuthController     authController.IAuthController
	UserController     userController.IUserController
	MedicineController medicineController.IMedicineController
	VehicleController  vehicleController.IVehicleController
	RideController     rideController.IRideController
	PaymentController  paymentController.IPaymentController

	// Repositories
	UserRepository     user.UserRepositoryInterface
	MedicineRepository medicine.MedicineRepositoryInterface
	VehicleRepository  vehicle.VehicleRepositoryInterface
	RideRepository     ride.RideRepositoryInterface
	PaymentRepository  payment.PaymentRepositoryInterface

	// Use Cases
	AuthUseCase     authUseCase.IAuthUseCase
	UserUseCase     userUseCase.IUserUseCase
	MedicineUseCase medicineUseCase.IMedicineUseCase
	VehicleUseCase  vehicleUseCase.IVehicleUseCase
	RideUseCase     rideUseCase.IRideUseCase
	PaymentUseCase  paymentUseCase.IPaymentUseCase
}

var (
	loggerInstance *logger.Logger
	loggerOnce     sync.Once
)

func GetLogger() *logger.Logger {
	loggerOnce.Do(func() {
		loggerInstance, _ = logger.NewLogger()
	})
	return loggerInstance
}

// SetupDependencies creates a new application context with all dependencies
func SetupDependencies(loggerInstance *logger.Logger) (*ApplicationContext, error) {
	// Initialize database with logger
	db, err := psql.InitPSQLDB(loggerInstance)
	if err != nil {
		return nil, err
	}

	// Initialize JWT service (manages its own configuration)
	jwtService := security.NewJWTService()

	// Initialize repositories with logger
	userRepo := user.NewUserRepository(db, loggerInstance)
	medicineRepo := medicine.NewMedicineRepository(db, loggerInstance)
	vehicleRepo := vehicle.NewVehicleRepository(db, loggerInstance)
	rideRepo := ride.NewRideRepository(db, loggerInstance)
	paymentRepo := payment.NewPaymentRepository(db, loggerInstance)

	// Initialize use cases with logger
	authUC := authUseCase.NewAuthUseCase(userRepo, jwtService, loggerInstance)
	userUC := userUseCase.NewUserUseCase(userRepo, loggerInstance)
	medicineUC := medicineUseCase.NewMedicineUseCase(medicineRepo, loggerInstance)
	vehicleUC := vehicleUseCase.NewVehicleUseCase(vehicleRepo, loggerInstance)
	rideUC := rideUseCase.NewRideUseCase(rideRepo, userRepo, vehicleRepo, loggerInstance)
	paymentUC := paymentUseCase.NewPaymentUseCase(paymentRepo, rideRepo, loggerInstance)

	// Initialize controllers with logger
	authCtrl := authController.NewAuthController(authUC, loggerInstance)
	userCtrl := userController.NewUserController(userUC, loggerInstance)
	medicineCtrl := medicineController.NewMedicineController(medicineUC, loggerInstance)
	vehicleCtrl := vehicleController.NewVehicleController(vehicleUC, loggerInstance)
	rideCtrl := rideController.NewRideController(rideUC, loggerInstance)
	paymentCtrl := paymentController.NewPaymentController(paymentUC, loggerInstance)

	return &ApplicationContext{
		// Core infrastructure
		DB:         db,
		Logger:     loggerInstance,
		JWTService: jwtService,

		// Controllers
		AuthController:     authCtrl,
		UserController:     userCtrl,
		MedicineController: medicineCtrl,
		VehicleController:  vehicleCtrl,
		RideController:     rideCtrl,
		PaymentController:  paymentCtrl,

		// Repositories
		UserRepository:     userRepo,
		MedicineRepository: medicineRepo,
		VehicleRepository:  vehicleRepo,
		RideRepository:     rideRepo,
		PaymentRepository:  paymentRepo,

		// Use Cases
		AuthUseCase:     authUC,
		UserUseCase:     userUC,
		MedicineUseCase: medicineUC,
		VehicleUseCase:  vehicleUC,
		RideUseCase:     rideUC,
		PaymentUseCase:  paymentUC,
	}, nil
}

// NewTestApplicationContext creates an application context for testing with mocked dependencies
func NewTestApplicationContext(
	mockUserRepo user.UserRepositoryInterface,
	mockMedicineRepo medicine.MedicineRepositoryInterface,
	mockJWTService security.IJWTService,
	loggerInstance *logger.Logger,
) *ApplicationContext {
	// Initialize use cases with mocked repositories and logger
	authUC := authUseCase.NewAuthUseCase(mockUserRepo, mockJWTService, loggerInstance)
	userUC := userUseCase.NewUserUseCase(mockUserRepo, loggerInstance)
	medicineUC := medicineUseCase.NewMedicineUseCase(mockMedicineRepo, loggerInstance)

	// Initialize controllers with logger
	authController := authController.NewAuthController(authUC, loggerInstance)
	userController := userController.NewUserController(userUC, loggerInstance)
	medicineController := medicineController.NewMedicineController(medicineUC, loggerInstance)

	return &ApplicationContext{
		Logger:             loggerInstance,
		AuthController:     authController,
		UserController:     userController,
		MedicineController: medicineController,
		JWTService:         mockJWTService,
		UserRepository:     mockUserRepo,
		MedicineRepository: mockMedicineRepo,
		AuthUseCase:        authUC,
		UserUseCase:        userUC,
		MedicineUseCase:    medicineUC,
	}
}
