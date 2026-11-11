package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"belajar-golang-context/service/client"
	"belajar-golang-context/service/handler"
	"belajar-golang-context/service/middleware"
	"belajar-golang-context/service/repository"
)

// AppServer membungkus http.Server dengan graceful shutdown berbasis context
type AppServer struct {
	server *http.Server
	port   int
}

// NewAppServer menginisialisasi router, dependency injection, dan http.Server
func NewAppServer(port int) *AppServer {
	userRepo := repository.NewUserRepository()
	paymentClient := client.NewPaymentClient(300 * time.Millisecond)
	orderHandler := handler.NewOrderHandler(userRepo, paymentClient, 1*time.Second)

	mux := http.NewServeMux()

	// Endpoint health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	// Endpoint pemrosesan pesanan
	mux.HandleFunc("POST /order", orderHandler.CreateOrder)
	mux.HandleFunc("GET /order", orderHandler.CreateOrder) // Kemudahan pengujian via browser / curl biasa

	// Bungkus router dengan middleware context
	handlerWithMiddleware := middleware.ContextMiddleware(mux)

	return &AppServer{
		port: port,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handlerWithMiddleware,
		},
	}
}

// Start menjalankan server dengan graceful shutdown menggunakan sinyal OS dan context.WithTimeout
func (s *AppServer) Start() error {
	// 1. Channel untuk menangkap error saat server listen & serve
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server berjalan di http://localhost:%d\n", s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// 2. Tangkap sinyal terminasi OS (SIGINT / SIGTERM)
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdownSignal:
		log.Printf("Menerima sinyal shutdown OS (%v). Memulai Graceful Shutdown...\n", sig)
	}

	// 3. Buat context dengan timeout untuk memberikan waktu pada request in-flight selesai
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown gagal: %w", err)
	}

	log.Println("Server berhasil dihentikan secara bersih (graceful shutdown complete).")
	return nil
}
