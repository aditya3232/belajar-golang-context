package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// Tipe kustom unexported untuk mencegah benturan key pada context
type ctxKey string

const (
	requestIDKey ctxKey = "request_id"
	userAuthKey  ctxKey = "user_auth"
)

// UserAuth menyimpan informasi identitas pengguna yang terautentikasi
type UserAuth struct {
	UserID   string
	Username string
	Role     string
}

// WithRequestID menyuntikkan RequestID ke context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID mengambil RequestID dari context
func GetRequestID(ctx context.Context) (string, bool) {
	val := ctx.Value(requestIDKey)
	if val == nil {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// WithUserAuth menyuntikkan UserAuth ke context
func WithUserAuth(ctx context.Context, auth UserAuth) context.Context {
	return context.WithValue(ctx, userAuthKey, auth)
}

// GetUserAuth mengambil UserAuth dari context
func GetUserAuth(ctx context.Context) (UserAuth, bool) {
	val := ctx.Value(userAuthKey)
	if val == nil {
		return UserAuth{}, false
	}
	auth, ok := val.(UserAuth)
	return auth, ok
}

// ContextMiddleware menginjeksi metadata request-scoped (Request ID & User Auth) ke dalam context.Context
func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Ambil atau buat X-Request-ID
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = generateRandomID()
		}

		// Set header response agar klien dapat melacak log
		w.Header().Set("X-Request-ID", reqID)

		// 2. Suntikkan RequestID ke context
		ctx := WithRequestID(r.Context(), reqID)

		// 3. Simulasi autentikasi user dari header Authorization / X-User-ID
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			role := r.Header.Get("X-User-Role")
			if role == "" {
				role = "customer"
			}
			ctx = WithUserAuth(ctx, UserAuth{
				UserID:   userID,
				Username: "user-" + userID,
				Role:     role,
			})
		}

		// Teruskan request dengan child context yang telah diperkaya value
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRandomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
