package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"belajar-golang-context/service/client"
	"belajar-golang-context/service/handler"
	"belajar-golang-context/service/middleware"
	"belajar-golang-context/service/repository"
	"belajar-golang-context/service/worker"
)

func setupTestServer() http.Handler {
	userRepo := repository.NewUserRepository()
	paymentClient := client.NewPaymentClient(150 * time.Millisecond)
	orderHandler := handler.NewOrderHandler(userRepo, paymentClient, 300*time.Millisecond)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /order", orderHandler.CreateOrder)
	mux.HandleFunc("GET /order", orderHandler.CreateOrder)

	return middleware.ContextMiddleware(mux)
}

func TestServiceNormalOrder(t *testing.T) {
	app := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/order", nil)
	req.Header.Set("X-User-ID", "1001")
	req.Header.Set("X-User-Role", "vip")
	req.Header.Set("X-Request-ID", "custom-req-001")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.OrderResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.RequestID != "custom-req-001" {
		t.Errorf("expected request_id 'custom-req-001', got '%s'", resp.RequestID)
	}
	if resp.UserID != "1001" {
		t.Errorf("expected user_id '1001', got '%s'", resp.UserID)
	}
	if resp.UserRole != "vip" {
		t.Errorf("expected user_role 'vip', got '%s'", resp.UserRole)
	}
	if resp.Status != "SUCCESS" {
		t.Errorf("expected status 'SUCCESS', got '%s'", resp.Status)
	}
}

func TestServiceUnauthorized(t *testing.T) {
	app := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/order", nil)
	// Tanpa header X-User-ID
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestServiceSlowDBTimeout(t *testing.T) {
	app := setupTestServer()

	// Kirim parameter simulate_slow_db=true
	req := httptest.NewRequest(http.MethodGet, "/order?simulate_slow_db=true", nil)
	req.Header.Set("X-User-ID", "1002")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504 (Gateway Timeout), got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp handler.OrderResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "TIMEOUT_OR_CANCELLED" {
		t.Errorf("expected status 'TIMEOUT_OR_CANCELLED', got '%s'", resp.Status)
	}
}

func TestServiceSlowPaymentTimeout(t *testing.T) {
	app := setupTestServer()

	// Kirim parameter simulate_slow_payment=true
	req := httptest.NewRequest(http.MethodGet, "/order?simulate_slow_payment=true", nil)
	req.Header.Set("X-User-ID", "1003")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504 (Gateway Timeout), got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestBatchProcessorWithDeadline(t *testing.T) {
	items := []worker.BatchItem{
		{ID: 1, Title: "Item 1"},
		{ID: 2, Title: "Item 2"},
		{ID: 3, Title: "Item 3"},
	}

	// 1. Uji deadline cukup panjang -> semua sukses selesai
	longDeadline := time.Now().Add(500 * time.Millisecond)
	resSuccess, err := worker.RunNightlySettlementTask(context.Background(), longDeadline, items)
	if err != nil {
		t.Fatalf("expected no error with long deadline, got %v", err)
	}
	if !resSuccess.Completed || resSuccess.ProcessedCount != 3 {
		t.Fatalf("expected all 3 items processed, got %d (completed: %t)", resSuccess.ProcessedCount, resSuccess.Completed)
	}

	// 2. Uji deadline sempit -> terhenti di tengah jalan
	shortDeadline := time.Now().Add(40 * time.Millisecond) // hanya cukup untuk ~1 item (delay 30ms/item)
	resCutoff, err := worker.RunNightlySettlementTask(context.Background(), shortDeadline, items)
	if err == nil {
		t.Fatal("expected cutoff error with short deadline, got nil")
	}
	if resCutoff.Completed {
		t.Fatal("expected batch to NOT be completed when deadline is exceeded")
	}
	if resCutoff.ProcessedCount >= 3 {
		t.Fatalf("expected items processed < 3, got %d", resCutoff.ProcessedCount)
	}
}

func TestBatchProcessorManualCancel(t *testing.T) {
	items := []worker.BatchItem{
		{ID: 1, Title: "Item 1"},
		{ID: 2, Title: "Item 2"},
		{ID: 3, Title: "Item 3"},
		{ID: 4, Title: "Item 4"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	processor := worker.NewBatchProcessor()

	// Jalankan batch di goroutine
	done := make(chan worker.BatchResult)
	go func() {
		res := processor.ProcessWithDeadline(ctx, items, 30*time.Millisecond)
		done <- res
	}()

	// Beri waktu proses item 1, lalu batalkan manual
	time.Sleep(45 * time.Millisecond)
	cancel()

	res := <-done
	if res.Completed {
		t.Fatal("expected batch to be cancelled before completion")
	}
}
