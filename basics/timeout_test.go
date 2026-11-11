package basics_test

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestContextWithTimeoutExceeded menguji bahwa context otomatis dibatalkan
// saat durasi timeout telah tercapai dan mengembalikan context.DeadlineExceeded
func TestContextWithTimeoutExceeded(t *testing.T) {
	timeout := 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // Best practice: selalu defer cancel() untuk membersihkan timer

	// Periksa apakah deadline tersedia
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set on context")
	}
	if time.Until(deadline) > timeout {
		t.Fatalf("expected deadline to be at most %v, got %v", timeout, time.Until(deadline))
	}

	// Tunggu hingga timeout terpicu
	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("expected error to be context.DeadlineExceeded, got %v", ctx.Err())
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context timed out too late or did not trigger")
	}
}

// TestContextWithTimeoutCompletesBeforeTimeout menguji operasi yang berhasil selesai
// sebelum batas timeout tercapai
func TestContextWithTimeoutCompletesBeforeTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200 * time.Millisecond)
	defer cancel()

	// Fungsi simulasi operasi cepat
	quickOperation := func(ctx context.Context) (string, error) {
		select {
		case <-time.After(20 * time.Millisecond):
			return "success", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	result, err := quickOperation(ctx)
	if err != nil {
		t.Fatalf("expected operation to succeed, got error: %v", err)
	}
	if result != "success" {
		t.Fatalf("expected 'success', got '%s'", result)
	}

	// Context seharusnya belum kadaluarsa
	if ctx.Err() != nil {
		t.Fatalf("context should not have error yet, got %v", ctx.Err())
	}
}

// TestContextWithTimeoutCancelsSlowOperation menguji pembatalan operasi lambat
// oleh batas waktu (timeout), mencegah proses menggantung (hanging)
func TestContextWithTimeoutCancelsSlowOperation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40 * time.Millisecond)
	defer cancel()

	// Fungsi simulasi query/request lambat (misal 500ms)
	slowOperation := func(ctx context.Context) error {
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			// Operasi dibatalkan oleh timeout
			return ctx.Err()
		}
	}

	start := time.Now()
	err := slowOperation(ctx)
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}

	// Operasi harus berhenti di sekitar 40ms, bukan menunggu hingga 500ms
	if elapsed > 150 * time.Millisecond {
		t.Fatalf("slow operation took too long to abort (%v)", elapsed)
	}
}

// TestManualCancelBeforeTimeout menguji pemanggilan cancel() manual sebelum waktu timeout habis
func TestManualCancelBeforeTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	// Batalkan secara manual lebih awal
	cancel()

	select {
	case <-ctx.Done():
		// Karena dibatalkan secara manual sebelum batas waktu habis,
		// error yang dihasilkan adalah context.Canceled, bukan context.DeadlineExceeded
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", ctx.Err())
		}
	default:
		t.Fatal("expected ctx.Done() to be closed after manual cancel")
	}
}
