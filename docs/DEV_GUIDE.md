# Developer Guide - Ride Hailing Backend

## 🎯 Overview

This guide provides comprehensive information for developers working on the Ride Hailing Backend. It covers architecture patterns, development workflows, testing strategies, and best practices.

## 📐 Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────────┐
│           EXTERNAL (Mobile App)             │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│     INFRASTRUCTURE LAYER                    │
│  - REST Controllers (Gin)                   │
│  - Repositories (GORM/PostgreSQL)           │
│  - JWT Security                             │
│  - Logging (Zap)                            │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│     APPLICATION LAYER                       │
│  - Use Cases (Business Logic)               │
│  - Auth, User, Vehicle, Ride, Payment       │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│     DOMAIN LAYER (Core Business)            │
│  - Entities (User, Vehicle, Ride, Payment)  │
│  - Business Rules & Validation              │
│  - NO external dependencies                 │
└─────────────────────────────────────────────┘
```

### Dependency Rule

**Critical**: Dependencies ONLY point inward
- ✅ Infrastructure → Application → Domain
- ❌ Domain NEVER imports from Infrastructure
- ❌ Application NEVER imports from Infrastructure (except interfaces)

## 🚀 Getting Started

### Prerequisites

```bash
# Required
- Go 1.24.2+
- Docker & Docker Compose
- Make (optional but recommended)

# Optional
- golangci-lint (for linting)
- Air (for hot reload in development)
```

### Initial Setup

```bash
# 1. Clone and enter directory
git clone <repository-url>
cd microservices-go-ride

# 2. Copy environment variables
cp .env.example .env
# Edit .env with your configurations

# 3. Start services
make run

# 4. Verify health
curl http://localhost:8080/v1/health
```

## 🏗️ Project Structure

```
/workspace/
├── src/
│   ├── domain/                    # 🎯 DOMAIN LAYER
│   │   ├── common/               # Shared enums & types
│   │   │   ├── enums.go          # UserRole, RideStatus, etc.
│   │   │   └── enums_test.go
│   │   ├── user/                 # User entity & logic
│   │   │   ├── user.go           # Entity + business methods
│   │   │   └── user_test.go
│   │   ├── vehicle/              # Vehicle entity
│   │   ├── ride/                 # Ride entity & state machine
│   │   ├── payment/              # Payment entity
│   │   └── errors/               # Domain errors
│   │
│   ├── application/               # 📋 APPLICATION LAYER
│   │   └── usecases/
│   │       ├── auth/             # Authentication logic
│   │       ├── user/             # User management
│   │       ├── vehicle/          # Vehicle management
│   │       ├── ride/             # Ride orchestration
│   │       └── payment/          # Payment processing
│   │
│   └── infrastructure/            # 🔧 INFRASTRUCTURE LAYER
│       ├── di/                   # Dependency Injection
│       │   └── application_context.go
│       ├── repository/psql/      # Database repositories
│       │   ├── user/
│       │   ├── vehicle/
│       │   ├── ride/
│       │   └── payment/
│       ├── rest/
│       │   ├── controllers/      # HTTP handlers
│       │   ├── middlewares/      # JWT, CORS, logging
│       │   └── routes/           # Route configuration
│       ├── security/             # JWT service
│       └── logger/               # Structured logging
│
├── Test/integration/             # Integration tests
├── .github/workflows/            # CI/CD pipelines
├── docs/                         # Documentation
├── docker-compose.yml            # Container orchestration
├── Dockerfile                    # Container definition
├── Makefile                      # Development commands
└── main.go                       # Application entry point
```

## 🔨 Development Workflow

### Adding a New Feature

#### 1. Start with Domain Layer

```go
// Example: Adding a new entity
// File: src/domain/notification/notification.go

package notification

import "time"

type Notification struct {
    ID        int
    UserID    int
    Message   string
    Type      NotificationType
    Read      bool
    CreatedAt time.Time
}

func (n *Notification) MarkAsRead() {
    n.Read = true
}
```

#### 2. Create Use Case

```go
// File: src/application/usecases/notification/notification.go

package notification

import (
    domain "github.com/gbrayhan/microservices-go/src/domain/notification"
    logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
)

type NotificationRepositoryInterface interface {
    Create(notification *domain.Notification) error
    GetByUserID(userID int) (*[]domain.Notification, error)
    MarkAsRead(id int) error
}

type INotificationUseCase interface {
    SendNotification(userID int, message string) error
    GetUserNotifications(userID int) (*[]domain.Notification, error)
}

type NotificationUseCase struct {
    repo   NotificationRepositoryInterface
    Logger *logger.Logger
}

func NewNotificationUseCase(repo NotificationRepositoryInterface, logger *logger.Logger) INotificationUseCase {
    return &NotificationUseCase{repo: repo, Logger: logger}
}

func (uc *NotificationUseCase) SendNotification(userID int, message string) error {
    notification := &domain.Notification{
        UserID:  userID,
        Message: message,
        Read:    false,
    }
    return uc.repo.Create(notification)
}
```

#### 3. Create Repository

```go
// File: src/infrastructure/repository/psql/notification/notification.go

package notification

import (
    domain "github.com/gbrayhan/microservices-go/src/domain/notification"
    logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
    "gorm.io/gorm"
)

type Notification struct {
    ID        int       `gorm:"primaryKey"`
    UserID    int       `gorm:"column:user_id;index"`
    Message   string    `gorm:"column:message"`
    Type      string    `gorm:"column:type"`
    Read      bool      `gorm:"column:read;default:false"`
    CreatedAt time.Time `gorm:"autoCreateTime:mili"`
}

func (Notification) TableName() string {
    return "notifications"
}

type Repository struct {
    DB     *gorm.DB
    Logger *logger.Logger
}

func NewNotificationRepository(db *gorm.DB, logger *logger.Logger) *Repository {
    return &Repository{DB: db, Logger: logger}
}

// Implement interface methods...
```

#### 4. Create Controller

```go
// File: src/infrastructure/rest/controllers/notification/notifications.go

package notification

import (
    "github.com/gbrayhan/microservices-go/src/application/usecases/notification"
    logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
    "github.com/gin-gonic/gin"
)

type NotificationController struct {
    useCase notification.INotificationUseCase
    Logger  *logger.Logger
}

func (c *NotificationController) GetNotifications(ctx *gin.Context) {
    // Implementation
}
```

#### 5. Wire in DI Container

```go
// File: src/infrastructure/di/application_context.go

// Add to ApplicationContext struct:
NotificationController notification.INotificationController
NotificationRepository notification.NotificationRepositoryInterface
NotificationUseCase    notification.INotificationUseCase

// Add to SetupDependencies function:
notificationRepo := notification.NewNotificationRepository(db, loggerInstance)
notificationUC := notification.NewNotificationUseCase(notificationRepo, loggerInstance)
notificationCtrl := notification.NewNotificationController(notificationUC, loggerInstance)
```

#### 6. Add Routes

```go
// File: src/infrastructure/rest/routes/notification.go

package routes

import (
    "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/notification"
    "github.com/gin-gonic/gin"
)

func NotificationRoutes(router *gin.RouterGroup, controller notification.INotificationController) {
    notificationGroup := router.Group("/notification")
    {
        notificationGroup.GET("", controller.GetNotifications)
        notificationGroup.POST("", controller.SendNotification)
    }
}

// Update routes.go:
NotificationRoutes(v1, appContext.NotificationController)
```

#### 7. Update Migrations

```go
// File: src/infrastructure/repository/psql/psql_repository.go

// Add to MigrateEntitiesGORM:
notificationModel := &notification.Notification{}
err := r.DB.AutoMigrate(
    userModel,
    vehicleModel,
    rideModel,
    paymentModel,
    notificationModel, // Add this
)
```

## 🧪 Testing Strategy

### Unit Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific package tests
go test ./src/domain/ride/...
go test ./src/application/usecases/ride/...
```

### Test Structure

```go
// Example: Testing domain logic
func TestRide_CanTransitionTo(t *testing.T) {
    tests := []struct {
        name      string
        current   RideStatus
        next      RideStatus
        wantValid bool
    }{
        {"Pending to Matched", RideStatusPending, RideStatusMatched, true},
        {"Matched to Completed", RideStatusMatched, RideStatusCompleted, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.current.CanTransitionTo(tt.next); got != tt.wantValid {
                t.Errorf("got %v, want %v", got, tt.wantValid)
            }
        })
    }
}
```

### Integration Tests

```bash
# Run integration tests (requires Docker)
make test-integration

# Or manually
./scripts/run-integration-test.bash
```

## 🔐 Authentication & Authorization

### JWT Flow

1. **Login**: POST `/v1/auth/login`
2. **Get Access Token**: Returns `accessToken` + `refreshToken`
3. **Use Access Token**: Add header `Authorization: Bearer <accessToken>`
4. **Refresh**: POST `/v1/auth/access-token` with `refreshToken`

### Role-Based Access

```go
// Middleware example
func RequireRole(role common.UserRole) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        user := ctx.MustGet("user").(domainUser.User)
        if user.Role != role {
            ctx.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
            return
        }
        ctx.Next()
    }
}

// Usage in routes
rideGroup.POST("/pending", RequireRole(common.RoleDriver), controller.GetPendingRides)
```

## 🗄️ Database Operations

### Migrations

```bash
# Migrations run automatically on startup via GORM AutoMigrate
# Tables created: users, vehicles, rides, payments

# Manual migration check
make db-shell
\dt  # List all tables
```

### Seed Data

Default users created on first run:
- **Admin**: `admin@ridehailing.com` / `Admin@123`
- **Sample Driver**: `driver.sample@ridehailing.com` / `Driver@123`
- **Sample Rider**: `rider.sample@ridehailing.com` / `Rider@123`

### Manual Queries

```bash
# Connect to database
make db-shell

# Example queries
SELECT * FROM users WHERE role = 'driver';
SELECT * FROM rides WHERE status = 'pending';
SELECT SUM(amount) FROM payments WHERE driver_id = 2 AND status = 'completed';
```

## 🔍 Debugging

### Logs

```bash
# View API logs
make logs

# View specific service
docker-compose logs -f api

# Filter logs
docker-compose logs api | grep ERROR
```

### Common Issues

#### 1. Port Already in Use

```bash
# Solution
docker-compose down
lsof -ti:8080 | xargs kill -9  # Kill process on port 8080
make run
```

#### 2. Database Connection Failed

```bash
# Check .env configuration
cat .env | grep DB_

# Verify PostgreSQL is running
docker-compose ps db

# Recreate database
make db-reset
```

#### 3. Migration Errors

```bash
# Check logs
docker-compose logs api | grep migration

# Manually trigger migration
docker-compose exec api ./microservice
```

## 📊 Performance Optimization

### Database Indexes

Already created for:
- Users: `role`, `is_available`, `(latitude, longitude)`
- Vehicles: `driver_id`, `status`, `vehicle_type`
- Rides: `rider_id`, `driver_id`, `status`, `created_at`
- Payments: `ride_id`, `rider_id`, `driver_id`, `status`

### Query Optimization Tips

```go
// ❌ Bad: N+1 query problem
for _, ride := range rides {
    driver, _ := userRepo.GetByID(ride.DriverID)
    // Process driver
}

// ✅ Good: Preload relationships
var rides []Ride
db.Preload("Driver").Preload("Vehicle").Find(&rides)
```

### Caching Strategy (Future Enhancement)

```go
// Redis integration example
type CachedUserRepository struct {
    repo  user.UserRepositoryInterface
    cache *redis.Client
}

func (r *CachedUserRepository) GetByID(id int) (*User, error) {
    // Check cache first
    if cached, err := r.cache.Get(ctx, "user:"+strconv.Itoa(id)).Result(); err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }
    
    // Fallback to database
    user, err := r.repo.GetByID(id)
    if err == nil {
        // Cache for 5 minutes
        r.cache.Set(ctx, "user:"+strconv.Itoa(id), user, 5*time.Minute)
    }
    return user, err
}
```

## 🎨 Code Style Guidelines

### Naming Conventions

```go
// ✅ Good
type VehicleUseCase struct { }
func (uc *VehicleUseCase) GetByDriverID(driverID int) { }

// ❌ Bad
type vehicleUC struct { }
func (uc *vehicleUC) get_by_driver_id(driver_id int) { }
```

### Error Handling

```go
// ✅ Good: Structured error with logging
user, err := r.userRepository.GetByID(id)
if err != nil {
    r.Logger.Error("Error getting user", zap.Error(err), zap.Int("id", id))
    return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
}

// ❌ Bad: Silent failure
user, _ := r.userRepository.GetByID(id)
```

### Logging

```go
// ✅ Good: Structured logging with context
r.Logger.Info("Creating ride",
    zap.Int("riderId", ride.RiderID),
    zap.Float64("estimatedFare", ride.EstimatedFare),
    zap.String("status", string(ride.Status)))

// ❌ Bad: String concatenation
fmt.Println("Creating ride for rider " + strconv.Itoa(ride.RiderID))
```

## 🔄 Git Workflow

### Branch Naming

```bash
feature/add-rating-system
bugfix/fix-driver-matching
hotfix/payment-refund-error
refactor/improve-ride-matching
```

### Commit Messages

```bash
# Format: <type>: <description>

feat: Add automatic driver matching algorithm
fix: Correct fare calculation for long distances
refactor: Optimize database queries for ride search
docs: Update API documentation for payment endpoints
test: Add unit tests for payment refund logic
```

### Pre-Commit Checklist

- [ ] Code formatted (`go fmt`)
- [ ] Tests pass (`make test`)
- [ ] No linter errors (`make lint`)
- [ ] Documentation updated
- [ ] CHANGELOG updated (if applicable)

## 📦 Deployment

### Environment Variables

```bash
# Production checklist
✅ Change JWT_ACCESS_SECRET_KEY (strong random)
✅ Change JWT_REFRESH_SECRET_KEY (strong random)
✅ Change DB_PASSWORD (strong password)
✅ Set APP_ENV=production
✅ Set GO_ENV=production
✅ Configure proper SSL certificates
✅ Set up database backups
```

### Docker Production Build

```bash
# Build for production
docker build -t ride-hailing-api:latest .

# Run with production env
docker run -d \
  --name ride-api \
  -p 8080:8080 \
  --env-file .env.production \
  ride-hailing-api:latest
```

### Health Checks

```bash
# API Health
curl http://localhost:8080/v1/health

# Database Health
docker-compose exec db pg_isready

# Service Status
docker-compose ps
```

## 🐛 Troubleshooting

### Common Errors

#### "Driver not available"

```bash
# Check driver status
curl http://localhost:8080/v1/user/2

# Update driver availability
curl -X PUT http://localhost:8080/v1/user/2 \
  -H "Content-Type: application/json" \
  -d '{"is_available": true}'
```

#### "Vehicle cannot accept rides"

```bash
# Check vehicle status
curl http://localhost:8080/v1/vehicle/1

# Activate vehicle
curl -X PUT http://localhost:8080/v1/vehicle/1/activate
```

## 📈 Monitoring & Observability

### Structured Logging

All logs include:
- `timestamp`: ISO 8601 format
- `level`: info, warn, error, debug
- `message`: Human-readable description
- Context fields: `userId`, `rideId`, etc.

### Metrics (Future Enhancement)

```go
// Prometheus integration example
var (
    ridesCreated = promauto.NewCounter(prometheus.CounterOpts{
        Name: "rides_created_total",
        Help: "Total number of rides created",
    })
    
    rideMatchDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "ride_match_duration_seconds",
        Help: "Time taken to match driver to ride",
    })
)
```

## 🎓 Best Practices

### DO's ✅

- **Always validate input** at controller level
- **Use interfaces** for dependencies
- **Log business operations** with context
- **Write tests** for critical logic
- **Keep controllers thin** - business logic in use cases
- **Use transactions** for multi-step operations
- **Handle all error cases**

### DON'Ts ❌

- **Never** import infrastructure in domain
- **Never** hardcode** configuration values
- **Never** expose sensitive data in logs
- **Never** trust user input without validation
- **Never** commit secrets to version control
- **Never** use `panic` in production code
- **Never** ignore errors

## 🔗 Useful Resources

- [Clean Architecture by Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [GORM Documentation](https://gorm.io/)
- [Gin Web Framework](https://gin-gonic.com/)

## 🆘 Getting Help

- **Internal Docs**: `/docs` directory
- **API Docs**: `docs/API_DOCUMENTATION.md`
- **Issues**: GitHub Issues
- **Team Chat**: Slack #ride-hailing-backend

## 📝 Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.

---

**Last Updated**: 2025-11-13
**Version**: 1.0.0
