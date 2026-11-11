package basics_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestContextWithCancelBasic menguji pembatalan manual secara langsung
func TestContextWithCancelBasic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Sebelum cancel dipanggil:
	if ctx.Err() != nil {
		t.Fatalf("expected ctx.Err() to be nil before cancel, got %v", ctx.Err())
	}

	select {
	case <-ctx.Done():
		t.Fatal("ctx.Done() should not be readable before cancel()")
	default:
		// Normal, channel masih open
	}

	// Panggil cancel
	cancel()

	// Setelah cancel dipanggil:
	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", ctx.Err())
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected ctx.Done() channel to be closed immediately after cancel()")
	}
}

// TestWorkerGoroutineCancellation mendemonstrasikan bagaimana sebuah worker goroutine
// mendengarkan sinyal pembatalan melalui channel ctx.Done() untuk berhenti bekerja dengan rapi
func TestWorkerGoroutineCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var processedCount int64
	done := make(chan struct{})

	// Jalankan worker goroutine
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				// Menerima sinyal pembatalan -> keluar dari loop dengan rapi
				return
			default:
				// Simulasi pemrosesan data
				atomic.AddInt64(&processedCount, 1)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Biarkan worker bekerja sebentar
	time.Sleep(50 * time.Millisecond)

	// Batalkan worker
	cancel()

	// Tunggu goroutine selesai keluar
	select {
	case <-done:
		// Berhasil berhenti
	case <-time.After(200 * time.Millisecond):
		t.Fatal("worker goroutine did not exit in response to cancellation")
	}

	countAfterCancel := atomic.LoadInt64(&processedCount)
	time.Sleep(30 * time.Millisecond)

	// Pastikan worker benar-benar berhenti dan tidak menambah hitungan lagi
	if atomic.LoadInt64(&processedCount) != countAfterCancel {
		t.Fatal("worker continued processing after cancellation")
	}
}

// TestCancelFanOutWorkers mendemonstrasikan satu context pembatalan yang membatalkan
// banyak worker goroutine secara bersamaan (broadcast cancellation)
func TestCancelFanOutWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const numWorkers = 5
	var wg sync.WaitGroup
	stoppedWorkers := make([]bool, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		workerID := i
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				stoppedWorkers[workerID] = true
			case <-time.After(2 * time.Second):
				// Timeout jika tidak menerima sinyal cancel
			}
		}()
	}

	// Beri jeda kecil agar semua goroutine aktif mendengarkan
	time.Sleep(20 * time.Millisecond)

	// Broadcast cancel ke semua worker
	cancel()

	// Tunggu semua worker keluar
	wg.Wait()

	for i, stopped := range stoppedWorkers {
		if !stopped {
			t.Fatalf("worker %d was not stopped by cancel broadcast", i)
		}
	}
}

// TestContextWithCancelCause mendemonstrasikan fitur Go 1.20+ WithCancelCause
// untuk menyematkan alasan spesifik di balik pembatalan context
func TestContextWithCancelCause(t *testing.T) {
	ctx, cancelCause := context.WithCancelCause(context.Background())

	errCustomCause := errors.New("upstream service connection dropped")

	// Batalkan dengan alasan spesifik
	cancelCause(errCustomCause)

	<-ctx.Done()

	// ctx.Err() standar tetap context.Canceled
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("expected ctx.Err() to be context.Canceled, got %v", ctx.Err())
	}

	// context.Cause(ctx) mengembalikan error spesifik yang menjadi penyebabnya
	cause := context.Cause(ctx)
	if !errors.Is(cause, errCustomCause) {
		t.Fatalf("expected cause to be '%v', got '%v'", errCustomCause, cause)
	}
}
