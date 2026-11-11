# 02. Context with Value (`context.WithValue`)

`context.WithValue` digunakan untuk menyisipkan data berumur pendek yang terikat langsung pada satu siklus request (*request-scoped metadata*).

---

## 1. Cara Kerja & Format Dasar

Sintaks dasar:
```go
ctx = context.WithValue(parentCtx, key, value)
```
- Menghasilkan child context baru yang menyimpan pasangan `key` dan `value`.
- Pencarian nilai dilakukan secara hirarkis dari child ke atas menuju root (*upwards search*).

---

## 2. Bahaya Key Collision & Solusi Type Safety

> [!WARNING]
> **JANGAN PERNAH** menggunakan tipe data bawaan umum seperti `string` atau `int` sebagai context key!
>
> Jika Package A menyimpan `context.WithValue(ctx, "user_id", "123")` dan Package B (atau library 3rd party) juga menyimpan `context.WithValue(ctx, "user_id", 456)`, nilai tersebut akan saling menimpa (*key collision*).

### Pola Solusi: Unexported Custom Type

Buat tipe alias kustom yang tidak diekspor (*unexported*) di level package Anda, lalu sediakan fungsi pembungkus (Setter & Getter) yang type-safe:

```go
package auth

import "context"

// 1. Tipe kustom unexported (hanya dikenal di package ini)
type contextKey struct{}

// 2. Variabel key instansiasi unik
var userClaimsKey = contextKey{}

// Struct data pengguna
type UserClaims struct {
	UserID string
	Role   string
}

// 3. Setter function: menyuntikkan value dengan aman
func WithUserClaims(ctx context.Context, claims UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

// 4. Getter function: mengekstrak value dengan type assertion aman
func GetUserClaims(ctx context.Context) (UserClaims, bool) {
	val := ctx.Value(userClaimsKey)
	if val == nil {
		return UserClaims{}, false
	}
	claims, ok := val.(UserClaims)
	return claims, ok
}
```

Keuntungan pola ini:
1. **Bebas Tabrakan Key**: Package lain tidak mungkin mengakses atau menimpa key `userClaimsKey`.
2. **Type Safe**: Caller tidak perlu melakukan casting manual `interface{}` / `any`.
3. **Pemberian Nilai Default yang Jelas**: Mengembalikan status boolean `ok` jika data tidak ditemukan.

---

## 3. Data Apa Saja yang Boleh Disimpan di Context?

Gunakan `context.WithValue` **HANYA** untuk data transitif yang melintasi batasan API dan proses (*cross-cutting concerns*):
- **Request ID / Correlation ID**: Untuk log aggregation di ELK/Datadog/Grafana.
- **Identitas Autentikasi Pengguna**: User ID, Role, Tenant ID dari JWT token.
- **Distributed Tracing**: Trace ID, Span Context (misal OpenTelemetry).
- **Client Metadata**: IP address, user agent.

---

## 4. Antipattern (Hal yang DILARANG)

> [!CAUTION]
> Jangan jadikan context sebagai "tempat sampah variabel global"!

1. **Dilarang memasukkan parameter fungsi opsional ke context**:
   *Salah*: Memasukkan `sort_by` atau `page_size` ke context.
   *Benar*: Operasikan parameter ini secara eksplisit pada argumen fungsi atau struct request.
2. **Dilarang memasukkan koneksi database atau logger mutable**:
   *Salah*: `context.WithValue(ctx, "db", sqlDB)`.
   *Benar*: Lewatkan database handler melalui struct dependency injection (misal pada struct repository).
3. **Dilarang memasukkan pointer yang dimutasi bersamaan**:
   Context diakses oleh banyak goroutine; menyimpan pointer yang nilainya diubah-ubah di tengah jalan dapat menyebabkan *data race*.
