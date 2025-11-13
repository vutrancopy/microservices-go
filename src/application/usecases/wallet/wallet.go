package wallet

import (
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/wallet"
	walletRepo "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/wallet"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

type IWalletUseCase interface {
	GetDriverWallet(driverID int) (*wallet.DriverWallet, error)
	CreateWallet(driverID int) (*wallet.DriverWallet, error)
	AddEarning(driverID int, amount float64, rideID int, description string) error
	RequestWithdrawal(request *wallet.WithdrawalRequest) (*wallet.WithdrawalRequest, error)
	GetWithdrawalRequests(driverID int) (*[]wallet.WithdrawalRequest, error)
	GetPendingWithdrawals() (*[]wallet.WithdrawalRequest, error)
	ApproveWithdrawal(id int, processedBy int) error
	RejectWithdrawal(id int, reason string, processedBy int) error
	GetTransactionHistory(driverID int, limit int) (*[]wallet.WalletTransaction, error)
	GetWalletBalance(driverID int) (float64, error)
}

type WalletUseCase struct {
	walletRepo *walletRepo.WalletRepository
	logger     *logger.Logger
}

func NewWalletUseCase(
	walletRepo *walletRepo.WalletRepository,
	logger *logger.Logger,
) *WalletUseCase {
	return &WalletUseCase{
		walletRepo: walletRepo,
		logger:     logger,
	}
}

func (u *WalletUseCase) GetDriverWallet(driverID int) (*wallet.DriverWallet, error) {
	u.logger.Info("Getting driver wallet", zap.Int("driverID", driverID))

	w, err := u.walletRepo.GetByDriverID(driverID)
	if err != nil {
		if err.Error() == "not_found" {
			// Create wallet if doesn't exist
			return u.CreateWallet(driverID)
		}
		return nil, err
	}

	return w, nil
}

func (u *WalletUseCase) CreateWallet(driverID int) (*wallet.DriverWallet, error) {
	u.logger.Info("Creating wallet for driver", zap.Int("driverID", driverID))

	newWallet := &wallet.DriverWallet{
		DriverID:         driverID,
		CurrentBalance:   0,
		TotalEarnings:    0,
		TotalWithdrawals: 0,
		PendingAmount:    0,
		Currency:         "USD",
		Active:           true,
	}

	createdWallet, err := u.walletRepo.CreateWallet(newWallet)
	if err != nil {
		u.logger.Error("Error creating wallet", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Wallet created successfully", zap.Int("id", createdWallet.ID))
	return createdWallet, nil
}

func (u *WalletUseCase) AddEarning(driverID int, amount float64, rideID int, description string) error {
	u.logger.Info("Adding earning to driver wallet", 
		zap.Int("driverID", driverID),
		zap.Float64("amount", amount),
		zap.Int("rideID", rideID))

	// Get or create wallet
	w, err := u.GetDriverWallet(driverID)
	if err != nil {
		return err
	}

	// Record balance before
	balanceBefore := w.CurrentBalance

	// Update wallet balance
	if err := u.walletRepo.UpdateBalance(w.ID, amount, wallet.TransactionTypeEarning); err != nil {
		u.logger.Error("Error updating wallet balance", zap.Error(err))
		return err
	}

	// Get updated balance
	w, _ = u.walletRepo.GetByID(w.ID)
	balanceAfter := w.CurrentBalance

	// Create transaction record
	now := time.Now()
	transaction := &wallet.WalletTransaction{
		WalletID:      w.ID,
		Type:          wallet.TransactionTypeEarning,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Status:        wallet.TransactionStatusCompleted,
		RideID:        &rideID,
		Description:   description,
		ProcessedAt:   &now,
	}

	if err := u.walletRepo.CreateTransaction(transaction); err != nil {
		u.logger.Error("Error creating transaction", zap.Error(err))
		return err
	}

	u.logger.Info("Earning added successfully", 
		zap.Int("walletID", w.ID),
		zap.Float64("newBalance", balanceAfter))

	return nil
}

func (u *WalletUseCase) RequestWithdrawal(request *wallet.WithdrawalRequest) (*wallet.WithdrawalRequest, error) {
	u.logger.Info("Processing withdrawal request", 
		zap.Int("driverID", request.DriverID),
		zap.Float64("amount", request.Amount))

	// Validate withdrawal request
	if err := request.Validate(); err != nil {
		u.logger.Error("Invalid withdrawal request", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Get wallet
	w, err := u.GetDriverWallet(request.DriverID)
	if err != nil {
		return nil, err
	}

	request.WalletID = w.ID

	// Check if can withdraw
	if err := w.CanWithdraw(request.Amount); err != nil {
		u.logger.Error("Cannot withdraw", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Set status to pending
	request.Status = wallet.TransactionStatusPending
	request.RequestedAt = time.Now()

	// Create withdrawal request
	createdRequest, err := u.walletRepo.CreateWithdrawalRequest(request)
	if err != nil {
		u.logger.Error("Error creating withdrawal request", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Withdrawal request created successfully", zap.Int("id", createdRequest.ID))
	return createdRequest, nil
}

func (u *WalletUseCase) GetWithdrawalRequests(driverID int) (*[]wallet.WithdrawalRequest, error) {
	u.logger.Info("Getting driver withdrawal requests", zap.Int("driverID", driverID))

	requests, err := u.walletRepo.GetDriverWithdrawals(driverID)
	if err != nil {
		u.logger.Error("Error getting withdrawal requests", zap.Error(err))
		return nil, err
	}

	return requests, nil
}

func (u *WalletUseCase) GetPendingWithdrawals() (*[]wallet.WithdrawalRequest, error) {
	u.logger.Info("Getting pending withdrawal requests")

	requests, err := u.walletRepo.GetPendingWithdrawals()
	if err != nil {
		u.logger.Error("Error getting pending withdrawals", zap.Error(err))
		return nil, err
	}

	return requests, nil
}

func (u *WalletUseCase) ApproveWithdrawal(id int, processedBy int) error {
	u.logger.Info("Approving withdrawal", zap.Int("id", id), zap.Int("processedBy", processedBy))

	// Get withdrawal request
	req, err := u.walletRepo.GetWithdrawalByID(id)
	if err != nil {
		return err
	}

	if req.Status != wallet.TransactionStatusPending {
		return domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Get wallet
	w, err := u.walletRepo.GetByID(req.WalletID)
	if err != nil {
		return err
	}

	// Record balance before
	balanceBefore := w.CurrentBalance

	// Update wallet balance (deduct withdrawal)
	if err := u.walletRepo.UpdateBalance(req.WalletID, req.Amount, wallet.TransactionTypeWithdrawal); err != nil {
		u.logger.Error("Error processing withdrawal", zap.Error(err))
		return err
	}

	// Get updated balance
	w, _ = u.walletRepo.GetByID(req.WalletID)
	balanceAfter := w.CurrentBalance

	// Create transaction record
	now := time.Now()
	transaction := &wallet.WalletTransaction{
		WalletID:      req.WalletID,
		Type:          wallet.TransactionTypeWithdrawal,
		Amount:        -req.Amount, // Negative for debit
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Status:        wallet.TransactionStatusCompleted,
		WithdrawalID:  &req.ID,
		Description:   "Withdrawal approved",
		ProcessedAt:   &now,
	}

	if err := u.walletRepo.CreateTransaction(transaction); err != nil {
		u.logger.Error("Error creating transaction", zap.Error(err))
	}

	// Approve withdrawal request
	req.Approve(processedBy)

	updates := map[string]interface{}{
		"status":       req.Status,
		"processed_at": req.ProcessedAt,
		"processed_by": req.ProcessedBy,
	}

	_, err = u.walletRepo.UpdateWithdrawal(id, updates)
	if err != nil {
		u.logger.Error("Error approving withdrawal", zap.Error(err))
		return err
	}

	u.logger.Info("Withdrawal approved successfully", zap.Int("id", id))
	return nil
}

func (u *WalletUseCase) RejectWithdrawal(id int, reason string, processedBy int) error {
	u.logger.Info("Rejecting withdrawal", 
		zap.Int("id", id),
		zap.String("reason", reason),
		zap.Int("processedBy", processedBy))

	// Get withdrawal request
	req, err := u.walletRepo.GetWithdrawalByID(id)
	if err != nil {
		return err
	}

	if req.Status != wallet.TransactionStatusPending {
		return domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Reject withdrawal request
	if err := req.Reject(reason, processedBy); err != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	updates := map[string]interface{}{
		"status":         req.Status,
		"processed_at":   req.ProcessedAt,
		"processed_by":   req.ProcessedBy,
		"failure_reason": req.FailureReason,
	}

	_, err = u.walletRepo.UpdateWithdrawal(id, updates)
	if err != nil {
		u.logger.Error("Error rejecting withdrawal", zap.Error(err))
		return err
	}

	u.logger.Info("Withdrawal rejected successfully", zap.Int("id", id))
	return nil
}

func (u *WalletUseCase) GetTransactionHistory(driverID int, limit int) (*[]wallet.WalletTransaction, error) {
	u.logger.Info("Getting transaction history", zap.Int("driverID", driverID))

	// Get wallet
	w, err := u.GetDriverWallet(driverID)
	if err != nil {
		return nil, err
	}

	transactions, err := u.walletRepo.GetTransactions(w.ID, limit)
	if err != nil {
		u.logger.Error("Error getting transactions", zap.Error(err))
		return nil, err
	}

	return transactions, nil
}

func (u *WalletUseCase) GetWalletBalance(driverID int) (float64, error) {
	u.logger.Info("Getting wallet balance", zap.Int("driverID", driverID))

	w, err := u.GetDriverWallet(driverID)
	if err != nil {
		return 0, err
	}

	return w.CurrentBalance, nil
}
