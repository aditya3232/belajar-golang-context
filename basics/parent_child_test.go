package basics_test

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestParentCancellationPropagatesToChildren mendemonstrasikan bahwa ketika parent context
// dibatalkan, semua child context yang diturunkan darinya otomatis ikut dibatalkan.
func TestParentCancellationPropagatesToChildren(t *testing.T) {
	// 1. Buat parent context
	parentCtx, parentCancel := context.WithCancel(context.Background())

	// 2. Buat beberapa child context dari parent yang sama
	childCtx1, childCancel1 := context.WithCancel(parentCtx)
	defer childCancel1()

	childCtx2, childCancel2 := context.WithTimeout(parentCtx, 5*time.Second)
	defer childCancel2()

	// 3. Batalkan parent context
	parentCancel()

	// 4. Verifikasi bahwa parent context telah dibatalkan
	select {
	case <-parentCtx.Done():
		if !errors.Is(parentCtx.Err(), context.Canceled) {
			t.Fatalf("expected parent error to be context.Canceled, got %v", parentCtx.Err())
		}
	default:
		t.Fatal("expected parent context to be done")
	}

	// 5. Verifikasi bahwa semua child context otomatis ikut terbatalkan
	select {
	case <-childCtx1.Done():
		if !errors.Is(childCtx1.Err(), context.Canceled) {
			t.Fatalf("expected childCtx1 error to be context.Canceled, got %v", childCtx1.Err())
		}
	default:
		t.Fatal("expected childCtx1 to be cancelled when parent was cancelled")
	}

	select {
	case <-childCtx2.Done():
		if !errors.Is(childCtx2.Err(), context.Canceled) {
			t.Fatalf("expected childCtx2 error to be context.Canceled, got %v", childCtx2.Err())
		}
	default:
		t.Fatal("expected childCtx2 to be cancelled when parent was cancelled")
	}
}

// TestChildCancellationDoesNotAffectParentOrSibling mendemonstrasikan bahwa pembatalan
// pada satu child context terisolasi: tidak mempengaruhi parent maupun child context lainnya (sibling).
func TestChildCancellationDoesNotAffectParentOrSibling(t *testing.T) {
	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	childCtx1, childCancel1 := context.WithCancel(parentCtx)
	childCtx2, childCancel2 := context.WithCancel(parentCtx)
	defer childCancel2()

	// Batalkan HANYA child 1
	childCancel1()

	// Verifikasi child 1 selesai
	select {
	case <-childCtx1.Done():
		if !errors.Is(childCtx1.Err(), context.Canceled) {
			t.Fatalf("expected childCtx1 error to be context.Canceled, got %v", childCtx1.Err())
		}
	default:
		t.Fatal("expected childCtx1 to be done")
	}

	// Verifikasi parent TIDAK terpengaruh (masih aktif)
	select {
	case <-parentCtx.Done():
		t.Fatal("parent context should NOT be cancelled when child is cancelled")
	default:
		if parentCtx.Err() != nil {
			t.Fatalf("expected parentCtx.Err() to be nil, got %v", parentCtx.Err())
		}
	}

	// Verifikasi sibling (child 2) TIDAK terpengaruh
	select {
	case <-childCtx2.Done():
		t.Fatal("childCtx2 should NOT be cancelled when childCtx1 is cancelled")
	default:
		if childCtx2.Err() != nil {
			t.Fatalf("expected childCtx2.Err() to be nil, got %v", childCtx2.Err())
		}
	}
}

// TestMultilevelHierarchy mendemonstrasikan pohon konteks berlapis:
// Root -> Parent -> Child -> Grandchild.
// Membatalkan Child akan membatalkan Grandchild, namun Parent dan Root tetap aktif.
func TestMultilevelHierarchy(t *testing.T) {
	rootCtx := context.Background()

	parentCtx, parentCancel := context.WithCancel(rootCtx)
	defer parentCancel()

	childCtx, childCancel := context.WithCancel(parentCtx)

	grandchildCtx, grandchildCancel := context.WithCancel(childCtx)
	defer grandchildCancel()

	// Batalkan Child di tengah hierarki
	childCancel()

	// Child harus selesai
	<-childCtx.Done()

	// Grandchild harus ikut selesai karena child-nya dibatalkan
	select {
	case <-grandchildCtx.Done():
		// Sesuai harapan
	case <-time.After(100 * time.Millisecond):
		t.Fatal("grandchild context should have been cancelled by child cancellation")
	}

	// Parent di atasnya TIDAK boleh terpengaruh
	if parentCtx.Err() != nil {
		t.Fatalf("parentCtx should still be active, got %v", parentCtx.Err())
	}
}
