# 06. Best Practices & Aturan Emas Context di Go

Berikut adalah panduan standar industri (*industry best practices*) dan aturan emas dalam menggunakan package `context` di Go.

---

## 7 Aturan Emas Context di Go

### 1. Letakkan `ctx context.Context` sebagai Parameter Pertama
Selalu jadikan `ctx context.Context` sebagai parameter urutan pertama dalam penulisan fungsi atau method:

```go
// ✅ BENAR
func FetchUser(ctx context.Context, userID string) (*User, error)

// ❌ SALAH
func FetchUser(userID string, ctx context.Context) (*User, error)
```

---

### 2. Jangan Simpan Context di Dalam Struct
Context dirancang untuk bersifat sementara (*transient*) dan terikat pada alur panggilan fungsi (*call chain*), bukan disimpan sebagai field struct yang hidup lama:

```go
// ❌ SALAH: Menyimpan context di dalam struct
type Service struct {
	ctx context.Context
	db  *sql.DB
}

// ✅ BENAR: Lewatkan context melalui method pemanggilan
type Service struct {
	db *sql.DB
}

func (s *Service) Process(ctx context.Context, data Data) error { ... }
```

> **Pengecualian Langka**: Satu-satunya pengecualian standar adalah `http.Request` bawaan standard library Go yang memang merepresentasikan satu siklus HTTP request spesifik.

---

### 3. Jangan Pernah Mengirimkan `nil` Context
Mengirimkan `nil` sebagai context dapat menyebabkan `panic` saat fungsi mencoba memanggil method seperti `ctx.Done()`.

```go
// ❌ SALAH
result, err := repository.GetUser(nil, id)

// ✅ BENAR (jika belum pasti context mana yang dipakai)
result, err := repository.GetUser(context.TODO(), id)

// ✅ BENAR (jika merupakan root operasi independen)
result, err := repository.GetUser(context.Background(), id)
```

---

### 4. Selalu Panggil Fungsi `cancel()`
Setiap pemanggilan `WithCancel`, `WithTimeout`, atau `WithDeadline` menghasilkan fungsi `cancel`. Wajib memanggil fungsi ini untuk membebaskan alokasi timer dan dereferensi memori:

```go
ctx, cancel := context.WithTimeout(parentCtx, 2*time.Second)
defer cancel() // ✅ Selalu gunakan defer cancel()
```

---

### 5. Gunakan Context Value Hanya untuk Data Request-Scoped
`context.WithValue` bukan tempat penyimpanan variabel global atau konfigurasi aplikasi.

- ✅ **Boleh**: Request ID, Correlation ID, User Token Claims, Tenant ID, Distributed Trace Spans.
- ❌ **Dilarang**: Koneksi database, repository instance, logger struct, parameter fungsi opsional, atau pointer/map mutable.

---

### 6. Gunakan Tipe Kustom Unexported untuk Key Context
Hindari tabrakan key (*key collision*) antar package dengan mendefinisikan tipe kustom unexported:

```go
// ✅ BENAR
type contextKey struct{}
var userRoleKey = contextKey{}

// ❌ SALAH
const userRoleKey = "user_role" // Rentan tabrakan dengan package lain
```

---

### 7. Selalu Dengarkan `<-ctx.Done()` pada Operasi Blocking & Loop
Jika Anda menulis goroutine worker atau fungsi yang memproses loop panjang, periksa channel pembatalan secara berkala:

```go
for item := range taskQueue {
	select {
	case <-ctx.Done():
		return ctx.Err() // Berhenti dengan rapi saat dibatalkan
	default:
		process(item)
	}
}
```

---

## Ringkasan Fungsi Utama Package `context`

| Fungsi | Kegunaan Utama | Kapan Digunakan? |
| :--- | :--- | :--- |
| `context.Background()` | Root context kosong | Pada entry point aplikasi (`main()`, test root, background worker). |
| `context.TODO()` | Placeholder context | Saat refactoring atau belum jelas context mana yang harus dipassing. |
| `context.WithValue(parent, k, v)` | Menyimpan metadata | Passing Request ID, Auth claims, atau tracing context. |
| `context.WithCancel(parent)` | Pembatalan manual | Graceful shutdown, manual abort, fan-out worker stop. |
| `context.WithTimeout(parent, duration)` | Batas durasi relatif | HTTP Client calls, Database queries, RPC requests. |
| `context.WithDeadline(parent, time)` | Batas titik waktu absolut | Scheduled batch cutoffs, hard deadline synchronization. |
| `context.WithCancelCause(parent)` *(Go 1.20+)* | Pembatalan dengan alasan error | Debugging dan pencatatan alasan spesifik pembatalan context. |
