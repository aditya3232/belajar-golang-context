package basics_test

import (
	"context"
	"testing"
)

// Tipe kustom unexported untuk mencegah collision / tabrakan key antar-package
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

// Helper functions (Setter & Getter) sebagai idiom rekomendasi Go

// WithRequestID menyuntikkan request ID ke dalam context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID mengambil request ID dari context dengan aman (type assertion)
func GetRequestID(ctx context.Context) (string, bool) {
	val := ctx.Value(requestIDKey)
	if val == nil {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// WithUserID menyuntikkan user ID ke dalam context
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID mengambil user ID dari context
func GetUserID(ctx context.Context) (int64, bool) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return 0, false
	}
	id, ok := val.(int64)
	return id, ok
}

// TestContextWithValueTypeSafety menguji penyimpanan dan ekstraksi data secara type-safe
func TestContextWithValueTypeSafety(t *testing.T) {
	ctx := context.Background()

	ctx = WithRequestID(ctx, "req-xyz-123")
	ctx = WithUserID(ctx, 42)

	// Verifikasi Request ID
	reqID, ok := GetRequestID(ctx)
	if !ok || reqID != "req-xyz-123" {
		t.Fatalf("expected request ID 'req-xyz-123', got '%s', ok=%v", reqID, ok)
	}

	// Verifikasi User ID
	userID, ok := GetUserID(ctx)
	if !ok || userID != 42 {
		t.Fatalf("expected user ID 42, got %d, ok=%v", userID, ok)
	}

	// Verifikasi Key yang tidak ada
	emptyCtx := context.Background()
	_, ok = GetRequestID(emptyCtx)
	if ok {
		t.Fatal("expected GetRequestID to return ok=false for empty context")
	}
}

// TestContextValueInheritance membuktikan bahwa child context mewarisi semua value dari parent
func TestContextValueInheritance(t *testing.T) {
	parentCtx := WithRequestID(context.Background(), "req-parent-001")

	// Child hanya menambahkan userID
	childCtx := WithUserID(parentCtx, 99)

	// Child harus memiliki requestID dari parent dan userID miliknya sendiri
	reqID, ok := GetRequestID(childCtx)
	if !ok || reqID != "req-parent-001" {
		t.Fatalf("child expected to inherit reqID 'req-parent-001', got '%s'", reqID)
	}

	userID, ok := GetUserID(childCtx)
	if !ok || userID != 99 {
		t.Fatalf("child expected to have userID 99, got %d", userID)
	}

	// Parent TIDAK boleh memiliki userID milik child (immutability)
	_, ok = GetUserID(parentCtx)
	if ok {
		t.Fatal("parent should not have userID set by child")
	}
}

// TestContextValueShadowing membuktikan bahwa child dapat menimpa (shadow) value parent
// tanpa mengubah nilai asli pada parent
func TestContextValueShadowing(t *testing.T) {
	parentCtx := WithRequestID(context.Background(), "original-id")
	childCtx := WithRequestID(parentCtx, "overridden-id")

	childID, _ := GetRequestID(childCtx)
	if childID != "overridden-id" {
		t.Fatalf("expected child to have overridden-id, got %s", childID)
	}

	parentID, _ := GetRequestID(parentCtx)
	if parentID != "original-id" {
		t.Fatalf("expected parent to maintain original-id, got %s", parentID)
	}
}

// TestKeyCollisionPrevention mendemonstrasikan bahwa tipe kustom berbeda mencegah
// tabrakan nilai meskipun string literal dasarnya sama
func TestKeyCollisionPrevention(t *testing.T) {
	type PackageAKey string
	type PackageBKey string

	const KeyA PackageAKey = "role"
	const KeyB PackageBKey = "role"

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyA, "admin")
	ctx = context.WithValue(ctx, KeyB, "moderator")

	valA := ctx.Value(KeyA).(string)
	valB := ctx.Value(KeyB).(string)

	if valA != "admin" || valB != "moderator" {
		t.Fatalf("key collision occurred: valA=%s, valB=%s", valA, valB)
	}
}
