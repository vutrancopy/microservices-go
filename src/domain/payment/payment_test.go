package payment

import (
	"testing"

	"github.com/gbrayhan/microservices-go/src/domain/common"
)

func TestPayment_IsPending(t *testing.T) {
	payment := &Payment{Status: common.PaymentStatusPending}
	if !payment.IsPending() {
		t.Error("Expected payment to be pending")
	}
}

func TestPayment_IsCompleted(t *testing.T) {
	payment := &Payment{Status: common.PaymentStatusCompleted}
	if !payment.IsCompleted() {
		t.Error("Expected payment to be completed")
	}
}

func TestPayment_IsFailed(t *testing.T) {
	payment := &Payment{Status: common.PaymentStatusFailed}
	if !payment.IsFailed() {
		t.Error("Expected payment to be failed")
	}
}

func TestPayment_IsRefunded(t *testing.T) {
	payment := &Payment{Status: common.PaymentStatusRefunded}
	if !payment.IsRefunded() {
		t.Error("Expected payment to be refunded")
	}
}

func TestPayment_CanBeRefunded(t *testing.T) {
	tests := []struct {
		name   string
		status common.PaymentStatus
		want   bool
	}{
		{"Pending cannot be refunded", common.PaymentStatusPending, false},
		{"Completed can be refunded", common.PaymentStatusCompleted, true},
		{"Failed cannot be refunded", common.PaymentStatusFailed, false},
		{"Already refunded", common.PaymentStatusRefunded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{Status: tt.status}
			if got := payment.CanBeRefunded(); got != tt.want {
				t.Errorf("CanBeRefunded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPayment_MarkAsCompleted(t *testing.T) {
	tests := []struct {
		name          string
		status        common.PaymentStatus
		transactionID string
		wantError     bool
	}{
		{"Pending payment can be completed", common.PaymentStatusPending, "txn_123", false},
		{"Completed payment cannot be completed again", common.PaymentStatusCompleted, "txn_456", true},
		{"Failed payment cannot be completed", common.PaymentStatusFailed, "txn_789", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{Status: tt.status}
			err := payment.MarkAsCompleted(tt.transactionID)
			if (err != nil) != tt.wantError {
				t.Errorf("MarkAsCompleted() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil {
				if payment.TransactionID != tt.transactionID {
					t.Error("TransactionID not set correctly")
				}
				if !payment.IsCompleted() {
					t.Error("Payment status not set to completed")
				}
				if payment.ProcessedAt == nil {
					t.Error("ProcessedAt not set")
				}
			}
		})
	}
}

func TestPayment_MarkAsFailed(t *testing.T) {
	tests := []struct {
		name      string
		status    common.PaymentStatus
		reason    string
		wantError bool
	}{
		{"Pending payment can fail", common.PaymentStatusPending, "Insufficient funds", false},
		{"Completed payment cannot fail", common.PaymentStatusCompleted, "Some reason", true},
		{"Already failed payment", common.PaymentStatusFailed, "Another reason", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{Status: tt.status}
			err := payment.MarkAsFailed(tt.reason)
			if (err != nil) != tt.wantError {
				t.Errorf("MarkAsFailed() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil {
				if payment.FailureReason != tt.reason {
					t.Error("FailureReason not set correctly")
				}
				if !payment.IsFailed() {
					t.Error("Payment status not set to failed")
				}
				if payment.ProcessedAt == nil {
					t.Error("ProcessedAt not set")
				}
			}
		})
	}
}

func TestPayment_ProcessRefund(t *testing.T) {
	tests := []struct {
		name      string
		status    common.PaymentStatus
		amount    float64
		total     float64
		wantError bool
	}{
		{"Valid refund on completed payment", common.PaymentStatusCompleted, 50.0, 100.0, false},
		{"Full refund", common.PaymentStatusCompleted, 100.0, 100.0, false},
		{"Refund more than paid", common.PaymentStatusCompleted, 150.0, 100.0, true},
		{"Negative refund amount", common.PaymentStatusCompleted, -10.0, 100.0, true},
		{"Zero refund amount", common.PaymentStatusCompleted, 0.0, 100.0, true},
		{"Cannot refund pending payment", common.PaymentStatusPending, 50.0, 100.0, true},
		{"Cannot refund failed payment", common.PaymentStatusFailed, 50.0, 100.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{
				Status: tt.status,
				Amount: tt.total,
			}
			err := payment.ProcessRefund(tt.amount, "Test refund")
			if (err != nil) != tt.wantError {
				t.Errorf("ProcessRefund() error = %v, wantError %v", err, tt.wantError)
			}
			if err == nil {
				if payment.RefundAmount != tt.amount {
					t.Error("RefundAmount not set correctly")
				}
				if !payment.IsRefunded() {
					t.Error("Payment status not set to refunded")
				}
				if payment.RefundedAt == nil {
					t.Error("RefundedAt not set")
				}
				if payment.RefundReason != "Test refund" {
					t.Error("RefundReason not set correctly")
				}
			}
		})
	}
}

func TestPayment_GetNetAmount(t *testing.T) {
	tests := []struct {
		name         string
		amount       float64
		refundAmount float64
		isRefunded   bool
		want         float64
	}{
		{"No refund", 100.0, 0.0, false, 100.0},
		{"Partial refund", 100.0, 30.0, true, 70.0},
		{"Full refund", 100.0, 100.0, true, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := common.PaymentStatusCompleted
			if tt.isRefunded {
				status = common.PaymentStatusRefunded
			}
			payment := &Payment{
				Amount:       tt.amount,
				RefundAmount: tt.refundAmount,
				Status:       status,
			}
			if got := payment.GetNetAmount(); got != tt.want {
				t.Errorf("GetNetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}
