package wallet

import (
	"encoding/json"

	
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/wallet"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WalletRepository handles wallet and transaction data access
type WalletRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

type WalletRepositoryInterface interface {
	CreateWallet(w *wallet.DriverWallet) (*wallet.DriverWallet, error)
	GetByDriverID(driverID int) (*wallet.DriverWallet, error)
	GetByID(id int) (*wallet.DriverWallet, error)
	UpdateBalance(walletID int, amount float64, transactionType wallet.TransactionType) error
	CreateTransaction(tx *wallet.WalletTransaction) error
	GetTransactions(walletID int, limit int) (*[]wallet.WalletTransaction, error)
	CreateWithdrawalRequest(req *wallet.WithdrawalRequest) (*wallet.WithdrawalRequest, error)
	GetWithdrawalByID(id int) (*wallet.WithdrawalRequest, error)
	GetDriverWithdrawals(driverID int) (*[]wallet.WithdrawalRequest, error)
	GetPendingWithdrawals() (*[]wallet.WithdrawalRequest, error)
	UpdateWithdrawal(id int, updates map[string]interface{}) (*wallet.WithdrawalRequest, error)
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *gorm.DB, logger *logger.Logger) *WalletRepository {
	return &WalletRepository{
		DB:     db,
		Logger: logger,
	}
}

func (r *WalletRepository) CreateWallet(w *wallet.DriverWallet) (*wallet.DriverWallet, error) {
	result := r.DB.Create(w)

	if result.Error != nil {
		var gormErr domainErrors.GormErr
		if err := json.Unmarshal([]byte(result.Error.Error()), &gormErr); err == nil {
			if gormErr.Number == 1062 {
				return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			}
		}
		r.Logger.Error("Error creating wallet", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return w, nil
}

func (r *WalletRepository) GetByDriverID(driverID int) (*wallet.DriverWallet, error) {
	var w wallet.DriverWallet
	result := r.DB.Where("driver_id = ?", driverID).First(&w)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &w, nil
}

func (r *WalletRepository) GetByID(id int) (*wallet.DriverWallet, error) {
	var w wallet.DriverWallet
	result := r.DB.First(&w, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &w, nil
}

func (r *WalletRepository) UpdateBalance(walletID int, amount float64, transactionType wallet.TransactionType) error {
	// Get current wallet
	var w wallet.DriverWallet
	if err := r.DB.First(&w, walletID).Error; err != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}

	// Calculate new balance
	updates := make(map[string]interface{})

	switch transactionType {
	case wallet.TransactionTypeEarning:
		updates["current_balance"] = w.CurrentBalance + amount
		updates["total_earnings"] = w.TotalEarnings + amount
	case wallet.TransactionTypeWithdrawal:
		updates["current_balance"] = w.CurrentBalance - amount
		updates["total_withdrawals"] = w.TotalWithdrawals + amount
	case wallet.TransactionTypeBonus:
		updates["current_balance"] = w.CurrentBalance + amount
	case wallet.TransactionTypeFee:
		updates["current_balance"] = w.CurrentBalance - amount
	case wallet.TransactionTypeRefund:
		updates["current_balance"] = w.CurrentBalance + amount
	case wallet.TransactionTypeAdjustment:
		updates["current_balance"] = w.CurrentBalance + amount
	}

	result := r.DB.Model(&w).Updates(updates)
	if result.Error != nil {
		r.Logger.Error("Error updating wallet balance", zap.Int("walletID", walletID), zap.Error(result.Error))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return nil
}

func (r *WalletRepository) CreateTransaction(tx *wallet.WalletTransaction) error {
	result := r.DB.Create(tx)

	if result.Error != nil {
		r.Logger.Error("Error creating wallet transaction", zap.Error(result.Error))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return nil
}

func (r *WalletRepository) GetTransactions(walletID int, limit int) (*[]wallet.WalletTransaction, error) {
	var transactions []wallet.WalletTransaction

	query := r.DB.Where("wallet_id = ?", walletID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&transactions)

	if result.Error != nil {
		r.Logger.Error("Error getting wallet transactions", 
			zap.Int("walletID", walletID), 
			zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &transactions, nil
}

func (r *WalletRepository) CreateWithdrawalRequest(req *wallet.WithdrawalRequest) (*wallet.WithdrawalRequest, error) {
	result := r.DB.Create(req)

	if result.Error != nil {
		r.Logger.Error("Error creating withdrawal request", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return req, nil
}

func (r *WalletRepository) GetWithdrawalByID(id int) (*wallet.WithdrawalRequest, error) {
	var req wallet.WithdrawalRequest
	result := r.DB.First(&req, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &req, nil
}

func (r *WalletRepository) GetDriverWithdrawals(driverID int) (*[]wallet.WithdrawalRequest, error) {
	var requests []wallet.WithdrawalRequest
	result := r.DB.Where("driver_id = ?", driverID).
		Order("requested_at DESC").
		Find(&requests)

	if result.Error != nil {
		r.Logger.Error("Error getting driver withdrawal requests", 
			zap.Int("driverID", driverID), 
			zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &requests, nil
}

func (r *WalletRepository) GetPendingWithdrawals() (*[]wallet.WithdrawalRequest, error) {
	var requests []wallet.WithdrawalRequest
	result := r.DB.Where("status = ?", wallet.TransactionStatusPending).
		Order("requested_at ASC").
		Find(&requests)

	if result.Error != nil {
		r.Logger.Error("Error getting pending withdrawal requests", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &requests, nil
}

func (r *WalletRepository) UpdateWithdrawal(id int, updates map[string]interface{}) (*wallet.WithdrawalRequest, error) {
	var req wallet.WithdrawalRequest
	if err := r.DB.First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	result := r.DB.Model(&req).Updates(updates)
	if result.Error != nil {
		r.Logger.Error("Error updating withdrawal request", zap.Int("id", id), zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &req, nil
}

