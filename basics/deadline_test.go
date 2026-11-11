package basics_test

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestContextWithDeadlineBasic menguji WithDeadline dengan target titik waktu absolut di masa depan
func TestContextWithDeadlineBasic(t *testing.T) {
	deadline := time.Now().Add(60 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Pastikan deadline sesuai dengan target waktu
	d, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}
	if !d.Equal(deadline) {
		t.Fatalf("expected deadline %v, got %v", deadline, d)
	}

	// Tunggu hingga deadline tercapai
	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got %v", ctx.Err())
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("deadline was not reached in expected timeframe")
	}
}

// TestContextWithAlreadyPassedDeadline menguji bahwa jika deadline yang diberikan
// sudah lewat (di masa lalu), context langsung ditutup seketika (immediately cancelled)
func TestContextWithAlreadyPassedDeadline(t *testing.T) {
	pastDeadline := time.Now().Add(-1 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), pastDeadline)
	defer cancel()

	// Channel Done() harus sudah langsung tertutup
	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got %v", ctx.Err())
		}
	default:
		t.Fatal("context with past deadline should be cancelled immediately")
	}
}

// TestChildCannotExtendParentDeadline membuktikan aturan penting Go context:
// Batas waktu child TIDAK BISA melebihi batas waktu parent.
// Jika child mencoba membuat deadline lebih lambat, Go akan tetap membatasi child
// dengan deadline parent yang lebih cepat.
func TestChildCannotExtendParentDeadline(t *testing.T) {
	now := time.Now()
	parentDeadline := now.Add(50 * time.Millisecond)
	attemptedChildDeadline := now.Add(500 * time.Millisecond) // Mencoba memperpanjang batas waktu

	parentCtx, parentCancel := context.WithDeadline(context.Background(), parentDeadline)
	defer parentCancel()

	childCtx, childCancel := context.WithDeadline(parentCtx, attemptedChildDeadline)
	defer childCancel()

	// Periksa deadline milik child
	childActualDeadline, ok := childCtx.Deadline()
	if !ok {
		t.Fatal("child context should have a deadline")
	}

	// Deadline anak HARUS sama dengan parentDeadline (karena parent lebih cepat)
	if !childActualDeadline.Equal(parentDeadline) {
		t.Fatalf("expected child to inherit earlier parent deadline (%v), but got %v", parentDeadline, childActualDeadline)
	}

	// Tunggu hingga parent deadline terlewati
	<-childCtx.Done()
	if !errors.Is(childCtx.Err(), context.DeadlineExceeded) {
		t.Fatalf("expected child to expire with DeadlineExceeded, got %v", childCtx.Err())
	}
}
