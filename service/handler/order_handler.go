package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"belajar-golang-context/service/client"
	"belajar-golang-context/service/middleware"
	"belajar-golang-context/service/repository"
)

// OrderHandler mengelola endpoint order & transaksi
type OrderHandler struct {
	userRepo      repository.UserRepository
	paymentClient client.PaymentClient
	overallTimeout time.Duration
}

// NewOrderHandler membuat instance baru OrderHandler
func NewOrderHandler(userRepo repository.UserRepository, paymentClient client.PaymentClient, timeout time.Duration) *OrderHandler {
	if timeout == 0 {
		timeout = 1 * time.Second
	}
	return &OrderHandler{
		userRepo:       userRepo,
		paymentClient:  paymentClient,
		overallTimeout: timeout,
	}
}

// OrderResponse format response JSON
type OrderResponse struct {
	RequestID     string  `json:"request_id"`
	UserID        string  `json:"user_id"`
	UserRole      string  `json:"user_role"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
	TransactionID string  `json:"transaction_id,omitempty"`
	DurationMs    int64   `json:"duration_ms"`
}

// CreateOrder menangani pembuatan pesanan dengan propagasi context end-to-end
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	w.Header().Set("Content-Type", "application/json")

	// 1. Ekstraksi data request-scoped dari context yang disuntikkan middleware (context.WithValue)
	reqID, _ := middleware.GetRequestID(r.Context())
	userAuth, hasAuth := middleware.GetUserAuth(r.Context())

	if !hasAuth {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(OrderResponse{
			RequestID: reqID,
			Status:    "FAILED",
			Message:   "Unauthorized: header X-User-ID wajib diisi",
		})
		return
	}

	// 2. Buat child context dengan timeout menyeluruh (overall timeout SLA)
	// Context ini akan dibatalkan otomatis jika handler memakan waktu melebihi h.overallTimeout
	ctx, cancel := context.WithTimeout(r.Context(), h.overallTimeout)
	defer cancel()

	// Parameter query untuk simulasi pengujian
	query := r.URL.Query()
	slowDB := query.Get("simulate_slow_db") == "true"
	slowPayment := query.Get("simulate_slow_payment") == "true"

	// 3. Langkah 1: Panggil User Repository dengan context
	user, err := h.userRepo.GetUser(ctx, userAuth.UserID, slowDB)
	if err != nil {
		h.handleContextError(w, err, reqID, userAuth.UserID, start)
		return
	}

	// 4. Langkah 2: Panggil Payment Client dengan context
	paymentReq := client.PaymentRequest{
		OrderID: "ORD-" + reqID,
		Amount:  250000.0,
	}
	paymentResp, err := h.paymentClient.Charge(ctx, paymentReq, slowPayment)
	if err != nil {
		h.handleContextError(w, err, reqID, user.ID, start)
		return
	}

	// 5. Response sukses
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(OrderResponse{
		RequestID:     reqID,
		UserID:        user.ID,
		UserRole:      userAuth.Role,
		Status:        "SUCCESS",
		Message:       "Order berhasil dibuat dan pembayaran selesai diproses",
		TransactionID: paymentResp.TransactionID,
		DurationMs:    time.Since(start).Milliseconds(),
	})
}

// handleContextError memetakan error context menjadi status HTTP yang representatif
func (h *OrderHandler) handleContextError(w http.ResponseWriter, err error, reqID, userID string, start time.Time) {
	statusCode := http.StatusInternalServerError
	message := err.Error()

	if errors.Is(err, context.DeadlineExceeded) {
		statusCode = http.StatusGatewayTimeout // 504
		message = "Request Gateway Timeout: pemrosesan melebihi batas waktu maksimal"
	} else if errors.Is(err, context.Canceled) {
		statusCode = 499 // Nginx Client Closed Request convention
		message = "Request dibatalkan oleh pengguna sebelum selesai"
	}

	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(OrderResponse{
		RequestID:  reqID,
		UserID:     userID,
		Status:     "TIMEOUT_OR_CANCELLED",
		Message:    message,
		DurationMs: time.Since(start).Milliseconds(),
	})
}
