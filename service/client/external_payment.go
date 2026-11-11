package client

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// PaymentRequest payload untuk memproses transaksi ke Payment Gateway
type PaymentRequest struct {
	OrderID string
	Amount  float64
}

// PaymentResponse response dari Payment Gateway
type PaymentResponse struct {
	TransactionID string
	Status        string
}

// PaymentClient mendefinisikan kontrak pemanggilan external API payment
type PaymentClient interface {
	Charge(ctx context.Context, req PaymentRequest, simulateSlow bool) (*PaymentResponse, error)
}

type paymentClientImpl struct {
	slaTimeout time.Duration
}

// NewPaymentClient menginisialisasi client dengan SLA timeout default
func NewPaymentClient(slaTimeout time.Duration) PaymentClient {
	if slaTimeout == 0 {
		slaTimeout = 300 * time.Millisecond
	}
	return &paymentClientImpl{
		slaTimeout: slaTimeout,
	}
}

// Charge memproses pembayaran ke external payment gateway dengan proteksi timeout ketat
func (c *paymentClientImpl) Charge(ctx context.Context, req PaymentRequest, simulateSlow bool) (*PaymentResponse, error) {
	// Buat child context dengan timeout khusus untuk panggilan eksternal ini (SLA)
	clientCtx, cancel := context.WithTimeout(ctx, c.slaTimeout)
	defer cancel() // Sangat penting agar timer child context segera dibersihkan

	latency := 50 * time.Millisecond
	if simulateSlow {
		// Simulasi gateway sedang lambat / mengalami latency spike
		latency = 1 * time.Second
	}

	select {
	case <-time.After(latency):
		// Pembayaran berhasil diproses
		return &PaymentResponse{
			TransactionID: fmt.Sprintf("TRX-%s-%d", req.OrderID, time.Now().UnixNano()),
			Status:        "SUCCESS",
		}, nil

	case <-clientCtx.Done():
		// Bisa terjadi karena:
		// 1. clientCtx mencapai SLA timeout (context.DeadlineExceeded)
		// 2. Parent ctx dibatalkan (misal: user membatalkan request browser)
		if errors.Is(clientCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("payment_client: external gateway SLA timeout (%v exceeded): %w", c.slaTimeout, clientCtx.Err())
		}
		return nil, fmt.Errorf("payment_client: request aborted by caller: %w", clientCtx.Err())
	}
}
