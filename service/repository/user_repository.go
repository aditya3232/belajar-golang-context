package repository

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// User merepresentasikan entitas pengguna di database
type User struct {
	ID      string
	Balance float64
	Status  string
}

// UserRepository mendefinisikan interface akses data pengguna
type UserRepository interface {
	GetUser(ctx context.Context, userID string, simulateSlow bool) (*User, error)
}

type userRepositoryImpl struct{}

// NewUserRepository membuat instance baru UserRepository
func NewUserRepository() UserRepository {
	return &userRepositoryImpl{}
}

// GetUser mensimulasikan query database yang menghormati context cancellation & timeout
func (r *userRepositoryImpl) GetUser(ctx context.Context, userID string, simulateSlow bool) (*User, error) {
	queryDuration := 50 * time.Millisecond
	if simulateSlow {
		// Simulasi database query yang lambat / lock table
		queryDuration = 2 * time.Second
	}

	// Simulasi pemrosesan I/O database dengan select ke ctx.Done()
	select {
	case <-time.After(queryDuration):
		// Query sukses selesai dalam batas waktu
		return &User{
			ID:      userID,
			Balance: 1500000.0,
			Status:  "ACTIVE",
		}, nil

	case <-ctx.Done():
		// Klien memutus koneksi atau durasi timeout tercapai
		// Pembatalan ini membebaskan thread/koneksi database pool
		return nil, fmt.Errorf("user_repository: query cancelled or timed out: %w", ctx.Err())
	}
}

var ErrUserNotFound = errors.New("user not found")
