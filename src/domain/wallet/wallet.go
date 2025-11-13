package wallet

import (
	"errors"
	"time"
)

// TransactionType defines the type of wallet transaction
type TransactionType string

const (
	TransactionTypeEarning    TransactionType = "earning"    // Ride earnings
	TransactionTypeWithdrawal TransactionType = "withdrawal" // Money withdrawn
	TransactionTypeBonus      TransactionType = "bonus"      // Platform bonus
	TransactionTypeFee        TransactionType = "fee"        // Platform fee deduction
	TransactionTypeRefund     TransactionType = "refund"     // Refund to driver
	TransactionTypeAdjustment TransactionType = "adjustment" // Manual adjustment
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"   // Awaiting processing
	TransactionStatusCompleted TransactionStatus = "completed" // Completed successfully
	TransactionStatusFailed    TransactionStatus = "failed"    // Failed
	TransactionStatusCancelled TransactionStatus = "cancelled" // Cancelled
)

// WithdrawalMethod defines withdrawal method
type WithdrawalMethod string

const (
	WithdrawalMethodBankTransfer WithdrawalMethod = "bank_transfer"
	WithdrawalMethodPayPal       WithdrawalMethod = "paypal"
	WithdrawalMethodCash         WithdrawalMethod = "cash"
	WithdrawalMethodCheck        WithdrawalMethod = "check"
)

// DriverWallet represents a driver's wallet/earnings account
type DriverWallet struct {
	ID                int
	DriverID          int
	CurrentBalance    float64   // Current available balance
	TotalEarnings     float64   // Lifetime earnings
	TotalWithdrawals  float64   // Total withdrawn
	PendingAmount     float64   // Pending (not yet available)
	LastWithdrawal    *time.Time
	Currency          string
	Active            bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// WalletTransaction represents a single wallet transaction
type WalletTransaction struct {
	ID              int
	WalletID        int
	Type            TransactionType
	Amount          float64 // Positive for credit, negative for debit
	BalanceBefore   float64
	BalanceAfter    float64
	Status          TransactionStatus
	RideID          *int    // Associated ride (for earnings)
	WithdrawalID    *int    // Associated withdrawal
	Description     string
	ReferenceNumber string
	ProcessedAt     *time.Time
	CreatedAt       time.Time
}

// WithdrawalRequest represents a driver's withdrawal request
type WithdrawalRequest struct {
	ID              int
	DriverID        int
	WalletID        int
	Amount          float64
	Method          WithdrawalMethod
	Status          TransactionStatus
	BankAccountInfo string // Encrypted bank details
	PayPalEmail     string
	RequestedAt     time.Time
	ProcessedAt     *time.Time
	ProcessedBy     *int    // Admin ID
	ReferenceNumber string
	FailureReason   string
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CanWithdraw checks if driver can make a withdrawal
func (w *DriverWallet) CanWithdraw(amount float64) error {
	if !w.Active {
		return errors.New("wallet is not active")
	}

	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}

	minWithdrawal := 10.0 // Minimum $10
	if amount < minWithdrawal {
		return errors.New("amount is below minimum withdrawal limit")
	}

	if amount > w.CurrentBalance {
		return errors.New("insufficient balance")
	}

	return nil
}

// AddEarning adds earnings to the wallet
func (w *DriverWallet) AddEarning(amount float64) error {
	if amount <= 0 {
		return errors.New("earning amount must be positive")
	}

	w.CurrentBalance += amount
	w.TotalEarnings += amount
	return nil
}

// ProcessWithdrawal processes a withdrawal
func (w *DriverWallet) ProcessWithdrawal(amount float64) error {
	if err := w.CanWithdraw(amount); err != nil {
		return err
	}

	w.CurrentBalance -= amount
	w.TotalWithdrawals += amount
	now := time.Now()
	w.LastWithdrawal = &now

	return nil
}

// AddPending adds amount to pending (not yet available)
func (w *DriverWallet) AddPending(amount float64) {
	w.PendingAmount += amount
}

// ReleasePending releases pending amount to available balance
func (w *DriverWallet) ReleasePending(amount float64) error {
	if amount > w.PendingAmount {
		return errors.New("cannot release more than pending amount")
	}

	w.PendingAmount -= amount
	w.CurrentBalance += amount
	return nil
}

// Validate validates the withdrawal request
func (wr *WithdrawalRequest) Validate() error {
	if wr.DriverID == 0 {
		return errors.New("driver ID is required")
	}

	if wr.Amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}

	if wr.Method == "" {
		return errors.New("withdrawal method is required")
	}

	// Validate method
	switch wr.Method {
	case WithdrawalMethodBankTransfer, WithdrawalMethodPayPal, 
		 WithdrawalMethodCash, WithdrawalMethodCheck:
		// valid
	default:
		return errors.New("invalid withdrawal method")
	}

	// Validate method-specific requirements
	if wr.Method == WithdrawalMethodBankTransfer && wr.BankAccountInfo == "" {
		return errors.New("bank account information is required")
	}

	if wr.Method == WithdrawalMethodPayPal && wr.PayPalEmail == "" {
		return errors.New("PayPal email is required")
	}

	return nil
}

// Approve approves the withdrawal request
func (wr *WithdrawalRequest) Approve(processedBy int) {
	wr.Status = TransactionStatusCompleted
	now := time.Now()
	wr.ProcessedAt = &now
	wr.ProcessedBy = &processedBy
}

// Reject rejects the withdrawal request
func (wr *WithdrawalRequest) Reject(reason string, processedBy int) error {
	if reason == "" {
		return errors.New("rejection reason is required")
	}

	wr.Status = TransactionStatusFailed
	now := time.Now()
	wr.ProcessedAt = &now
	wr.ProcessedBy = &processedBy
	wr.FailureReason = reason

	return nil
}

// IWalletService defines the interface for wallet operations
type IWalletService interface {
	GetDriverWallet(driverID int) (*DriverWallet, error)
	CreateWallet(driverID int) (*DriverWallet, error)
	AddEarning(driverID int, amount float64, rideID int, description string) error
	RequestWithdrawal(request *WithdrawalRequest) (*WithdrawalRequest, error)
	GetWithdrawalRequests(driverID int) (*[]WithdrawalRequest, error)
	GetPendingWithdrawals() (*[]WithdrawalRequest, error)
	ApproveWithdrawal(id int, processedBy int) error
	RejectWithdrawal(id int, reason string, processedBy int) error
	GetTransactionHistory(walletID int, limit int) (*[]WalletTransaction, error)
	GetWalletBalance(driverID int) (float64, error)
}
