# 03. Context with Cancel (`context.WithCancel`)

`context.WithCancel` digunakan untuk membatalkan operasi atau menghentikan sekumpulan goroutine secara manual ketika sebuah kondisi tertentu terpenuhi.

---

## 1. Mekanisme Kerja

Sintaks dasar:
```go
ctx, cancel := context.WithCancel(parentCtx)
defer cancel() // Wajib selalu dipanggil!
```

- Menghasilkan dua hal:
  1. `ctx`: context baru yang mendengarkan sinyal pembatalan.
  2. `cancel`: fungsi penutup (`CancelFunc`). Saat dipanggil, fungsi ini akan menutup channel `ctx.Done()`.
- Setiap goroutine yang mendengarkan `select { case <-ctx.Done(): ... }` akan segera terbangun dan dapat membersihkan resource sebelum keluar.

---

## 2. Mengapa Wajib Memanggil `defer cancel()`?

> [!IMPORTANT]
> Setiap kali Anda membuat context pembatalan (`WithCancel`, `WithTimeout`, `WithDeadline`), fungsi `cancel()` **HARUS SELALU** dipanggil sebelum fungsi pembuatnya keluar, biasanya menggunakan idiom `defer cancel()`.

Jika Anda tidak memanggil `cancel()`:
1. Child context akan tetap terdaftar di memori parent context selamanya (*memory leak*).
2. Timer internal (pada timeout/deadline) tidak akan dihentikan lebih awal (*goroutine leak*).

---

## 3. Pola Konsumsi Sinyal Pembatalan di Goroutine

Pola standar dalam mendengarkan sinyal pembatalan adalah menggunakan pernyataan `select`:

```go
func Worker(ctx context.Context, jobQueue <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			// Sinyal pembatalan diterima -> segera bersihkan resource dan keluar
			log.Println("Worker dihentikan:", ctx.Err())
			return

		case job, ok := <-jobQueue:
			if !ok {
				return
			}
			processJob(ctx, job)
		}
	}
}
```

---

## 4. Fitur Modern: `context.WithCancelCause` (Go 1.20+)

Mulai Go 1.20, Go memperkenalkan `context.WithCancelCause` yang memungkinkan Anda menyematkan alasan spesifik (error kustom) mengapa context dibatalkan:

```go
var ErrNetworkDown = errors.New("upstream service tidak dapat dijangkau")

ctx, cancelCause := context.WithCancelCause(context.Background())

// Batalkan dengan alasan spesifik
cancelCause(ErrNetworkDown)

// Di tempat lain:
<-ctx.Done()
fmt.Println(ctx.Err())            // context canceled
fmt.Println(context.Cause(ctx))   // upstream service tidak dapat dijangkau
```

Fitur ini sangat berguna untuk debugging dan tracing akar masalah di sistem terdistribusi.

---

## 5. Use Case Nyata: Graceful Shutdown HTTP Server

Saat server menerima sinyal `Ctrl+C` (`SIGINT`) atau sinyal shutdown dari Kubernetes (`SIGTERM`), kita ingin membatalkan context root dan memberikan waktu pada request yang sedang berjalan untuk menyelesaikan tugasnya:

```go
func main() {
	server := &http.Server{Addr: ":8080"}

	// Tangkap sinyal OS ke dalam context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Tunggu sinyal interupsi OS
	<-ctx.Done()
	log.Println("Memulai graceful shutdown...")

	// Beri batas waktu maksimal 5 detik untuk menutup koneksi aktif
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Shutdown gagal: %v", err)
	}
	log.Println("Server selesai dimatikan dengan aman.")
}
```
