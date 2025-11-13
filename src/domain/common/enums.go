package common

// UserRole defines the role of a user in the system
type UserRole string

const (
	RoleRider  UserRole = "rider"
	RoleDriver UserRole = "driver"
	RoleAdmin  UserRole = "admin"
)

// IsValid checks if the UserRole is valid
func (r UserRole) IsValid() bool {
	return r == RoleRider || r == RoleDriver || r == RoleAdmin
}

// String returns the string representation of UserRole
func (r UserRole) String() string {
	return string(r)
}

// RideStatus defines the status of a ride
type RideStatus string

const (
	RideStatusPending    RideStatus = "pending"
	RideStatusMatched    RideStatus = "matched"
	RideStatusInProgress RideStatus = "in_progress"
	RideStatusCompleted  RideStatus = "completed"
	RideStatusCancelled  RideStatus = "cancelled"
)

// IsValid checks if the RideStatus is valid
func (s RideStatus) IsValid() bool {
	switch s {
	case RideStatusPending, RideStatusMatched, RideStatusInProgress,
		RideStatusCompleted, RideStatusCancelled:
		return true
	}
	return false
}

// String returns the string representation of RideStatus
func (s RideStatus) String() string {
	return string(s)
}

// CanTransitionTo checks if a status transition is valid
func (s RideStatus) CanTransitionTo(newStatus RideStatus) bool {
	validTransitions := map[RideStatus][]RideStatus{
		RideStatusPending: {
			RideStatusMatched,
			RideStatusCancelled,
		},
		RideStatusMatched: {
			RideStatusInProgress,
			RideStatusCancelled,
		},
		RideStatusInProgress: {
			RideStatusCompleted,
			RideStatusCancelled,
		},
		RideStatusCompleted: {
			// No transitions from completed
		},
		RideStatusCancelled: {
			// No transitions from cancelled
		},
	}

	allowedTransitions, exists := validTransitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}
	return false
}

// PaymentStatus defines the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// IsValid checks if the PaymentStatus is valid
func (s PaymentStatus) IsValid() bool {
	return s == PaymentStatusPending ||
		s == PaymentStatusCompleted ||
		s == PaymentStatusFailed ||
		s == PaymentStatusRefunded
}

// String returns the string representation of PaymentStatus
func (s PaymentStatus) String() string {
	return string(s)
}

// PaymentMethod defines the payment method
type PaymentMethod string

const (
	PaymentMethodCash          PaymentMethod = "cash"
	PaymentMethodCreditCard    PaymentMethod = "credit_card"
	PaymentMethodDebitCard     PaymentMethod = "debit_card"
	PaymentMethodDigitalWallet PaymentMethod = "digital_wallet"
)

// IsValid checks if the PaymentMethod is valid
func (m PaymentMethod) IsValid() bool {
	return m == PaymentMethodCash ||
		m == PaymentMethodCreditCard ||
		m == PaymentMethodDebitCard ||
		m == PaymentMethodDigitalWallet
}

// String returns the string representation of PaymentMethod
func (m PaymentMethod) String() string {
	return string(m)
}

// VehicleType defines the type of vehicle
type VehicleType string

const (
	VehicleTypeSedan VehicleType = "sedan"
	VehicleTypeSUV   VehicleType = "suv"
	VehicleTypeVan   VehicleType = "van"
	VehicleTypeBike  VehicleType = "bike"
)

// IsValid checks if the VehicleType is valid
func (t VehicleType) IsValid() bool {
	return t == VehicleTypeSedan ||
		t == VehicleTypeSUV ||
		t == VehicleTypeVan ||
		t == VehicleTypeBike
}

// String returns the string representation of VehicleType
func (t VehicleType) String() string {
	return string(t)
}

// VehicleStatus defines the status of a vehicle
type VehicleStatus string

const (
	VehicleStatusActive   VehicleStatus = "active"
	VehicleStatusInactive VehicleStatus = "inactive"
	VehicleStatusInspection VehicleStatus = "inspection"
)

// IsValid checks if the VehicleStatus is valid
func (s VehicleStatus) IsValid() bool {
	return s == VehicleStatusActive ||
		s == VehicleStatusInactive ||
		s == VehicleStatusInspection
}

// String returns the string representation of VehicleStatus
func (s VehicleStatus) String() string {
	return string(s)
}
