package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/wallet"
	"github.com/gin-gonic/gin"
)

func WalletRoutes(router *gin.RouterGroup, controller wallet.IWalletController) {
	walletGroup := router.Group("/wallet")
	{
		// Driver endpoints
		walletGroup.GET("/driver/:driverId", controller.GetDriverWallet)
		walletGroup.GET("/driver/:driverId/balance", controller.GetWalletBalance)
		walletGroup.GET("/driver/:driverId/transactions", controller.GetTransactionHistory)
		walletGroup.POST("/driver/:driverId/withdraw", controller.RequestWithdrawal)
		walletGroup.GET("/driver/:driverId/withdrawals", controller.GetWithdrawalRequests)

		// Admin endpoints (should be protected with auth + admin middleware)
		walletGroup.GET("/withdrawals/pending", controller.GetPendingWithdrawals)
		walletGroup.POST("/withdrawals/:id/approve", controller.ApproveWithdrawal)
		walletGroup.POST("/withdrawals/:id/reject", controller.RejectWithdrawal)
	}
}
