package wallet

import (
	"net/http"
	"strconv"

	"github.com/gbrayhan/microservices-go/src/application/usecases/wallet"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainWallet "github.com/gbrayhan/microservices-go/src/domain/wallet"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IWalletController interface {
	GetDriverWallet(ctx *gin.Context)
	GetWalletBalance(ctx *gin.Context)
	GetTransactionHistory(ctx *gin.Context)
	RequestWithdrawal(ctx *gin.Context)
	GetWithdrawalRequests(ctx *gin.Context)
	GetPendingWithdrawals(ctx *gin.Context)
	ApproveWithdrawal(ctx *gin.Context)
	RejectWithdrawal(ctx *gin.Context)
}

type WalletController struct {
	walletUseCase wallet.IWalletUseCase
	logger        *logger.Logger
}

func NewWalletController(walletUseCase wallet.IWalletUseCase, logger *logger.Logger) *WalletController {
	return &WalletController{
		walletUseCase: walletUseCase,
		logger:        logger,
	}
}

type RequestWithdrawalRequest struct {
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Method          string  `json:"method" binding:"required"`
	BankAccountInfo string  `json:"bankAccountInfo"`
	PayPalEmail     string  `json:"paypalEmail"`
	Notes           string  `json:"notes"`
}

type ApproveWithdrawalRequest struct {
	ProcessedBy int `json:"processedBy" binding:"required"`
}

type RejectWithdrawalRequest struct {
	Reason      string `json:"reason" binding:"required"`
	ProcessedBy int    `json:"processedBy" binding:"required"`
}

func (c *WalletController) GetDriverWallet(ctx *gin.Context) {
	driverIDParam := ctx.Param("driverId")
	driverID, err := strconv.Atoi(driverIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting driver wallet", zap.Int("driverID", driverID))

	wallet, err := c.walletUseCase.GetDriverWallet(driverID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    wallet,
	})
}

func (c *WalletController) GetWalletBalance(ctx *gin.Context) {
	driverIDParam := ctx.Param("driverId")
	driverID, err := strconv.Atoi(driverIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting wallet balance", zap.Int("driverID", driverID))

	balance, err := c.walletUseCase.GetWalletBalance(driverID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"driverId": driverID,
			"balance":  balance,
		},
	})
}

func (c *WalletController) GetTransactionHistory(ctx *gin.Context) {
	driverIDParam := ctx.Param("driverId")
	driverID, err := strconv.Atoi(driverIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	limitParam := ctx.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitParam)

	c.logger.Info("Getting transaction history", 
		zap.Int("driverID", driverID),
		zap.Int("limit", limit))

	transactions, err := c.walletUseCase.GetTransactionHistory(driverID, limit)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transactions,
	})
}

func (c *WalletController) RequestWithdrawal(ctx *gin.Context) {
	driverIDParam := ctx.Param("driverId")
	driverID, err := strconv.Atoi(driverIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var req RequestWithdrawalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Processing withdrawal request", 
		zap.Int("driverID", driverID),
		zap.Float64("amount", req.Amount))

	withdrawalReq := &domainWallet.WithdrawalRequest{
		DriverID:        driverID,
		Amount:          req.Amount,
		Method:          domainWallet.WithdrawalMethod(req.Method),
		BankAccountInfo: req.BankAccountInfo,
		PayPalEmail:     req.PayPalEmail,
		Notes:           req.Notes,
	}

	created, err := c.walletUseCase.RequestWithdrawal(withdrawalReq)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Withdrawal request submitted successfully",
		"data":    created,
	})
}

func (c *WalletController) GetWithdrawalRequests(ctx *gin.Context) {
	driverIDParam := ctx.Param("driverId")
	driverID, err := strconv.Atoi(driverIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting withdrawal requests", zap.Int("driverID", driverID))

	requests, err := c.walletUseCase.GetWithdrawalRequests(driverID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
	})
}

func (c *WalletController) GetPendingWithdrawals(ctx *gin.Context) {
	c.logger.Info("Getting pending withdrawal requests")

	requests, err := c.walletUseCase.GetPendingWithdrawals()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
	})
}

func (c *WalletController) ApproveWithdrawal(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var req ApproveWithdrawalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Approving withdrawal", 
		zap.Int("id", id),
		zap.Int("processedBy", req.ProcessedBy))

	err = c.walletUseCase.ApproveWithdrawal(id, req.ProcessedBy)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal approved successfully",
	})
}

func (c *WalletController) RejectWithdrawal(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var req RejectWithdrawalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Rejecting withdrawal", 
		zap.Int("id", id),
		zap.String("reason", req.Reason))

	err = c.walletUseCase.RejectWithdrawal(id, req.Reason, req.ProcessedBy)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal rejected",
	})
}
