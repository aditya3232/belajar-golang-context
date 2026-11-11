# 🚀 Belajar Golang Context: Panduan Komprehensif & Real-World Service

Selamat datang di repositori **Belajar Golang Context**. Repositori ini dirancang sebagai panduan lengkap, terstruktur, dan aplikatif untuk memahami package `context` pada bahasa pemrograman Go — mulai dari konsep teoritis dasar hingga implementasi nyata pada arsitektur web service modern.

---

## 📚 Daftar Modul Pembelajaran

Seluruh materi dijelaskan secara mendalam dengan contoh kode dan diagram alir pada folder [`docs/`](./docs/):

1. **[01. Parent & Child Context](./docs/01-parent-child.md)**
   - Konsep pohon konteks (*Context Tree*) dari `context.Background()`.
   - Propagasi pembatalan top-down (pembatalan parent otomatis membatalkan seluruh child).
   - Isolasi pembatalan child (tidak mempengaruhi parent atau saudara).

2. **[02. Context with Value (`context.WithValue`)](./docs/02-context-with-value.md)**
   - Menyimpan data berumur pendek terikat request (*request-scoped*).
   - Bahaya tabrakan key (*key collision*) dan solusi menggunakan unexported custom type.
   - Pola Setter & Getter yang aman (*type-safe*).
   - Antipattern: hal-hal yang dilarang disimpan di dalam context.

3. **[03. Context with Cancel (`context.WithCancel`)](./docs/03-context-with-cancel.md)**
   - Mekanisme pembatalan manual dan penutupan channel `<-ctx.Done()`.
   - Mencegah memory dan goroutine leak dengan idiom `defer cancel()`.
   - Fitur modern Go 1.20+: `context.WithCancelCause` dan `context.Cause(ctx)`.
   - Implementasi Graceful Shutdown pada HTTP Server saat menerima sinyal OS (`SIGINT`/`SIGTERM`).

4. **[04. Context with Timeout (`context.WithTimeout`)](./docs/04-context-with-timeout.md)**
   - Membatasi durasi eksekusi operasi secara otomatis.
   - Penanganan error `context.DeadlineExceeded` vs `context.Canceled`.
   - Contoh nyata: SLA timeout pada panggilan HTTP client pihak ketiga dan query database lambat.

5. **[05. Context with Deadline (`context.WithDeadline`)](./docs/05-context-with-deadline.md)**
   - Perbedaan waktu absolut (`time.Time`) vs durasi relatif (`time.Duration`).
   - Aturan inheritansi deadline: child tidak bisa memperpanjang batas waktu parent.
   - Contoh nyata: Scheduled batch settlement dengan jendela cutoff waktu bank.

6. **[06. Best Practices & 7 Aturan Emas Context](./docs/06-best-practices.md)**
   - Aturan standar industri dalam penulisan context di Go.
   - Tabel ringkasan fungsi-fungsi inti package `context`.

---

## 🏗️ Struktur Proyek

```text
belajar-golang-context/
├── docs/                             # Modul panduan & materi edukasi
│   ├── 01-parent-child.md
│   ├── 02-context-with-value.md
│   ├── 03-context-with-cancel.md
│   ├── 04-context-with-timeout.md
│   ├── 05-context-with-deadline.md
│   └── 06-best-practices.md
├── basics/                           # Unit tests mandiri untuk masing-masing konsep
│   ├── parent_child_test.go          # Pengujian pohon konteks & hierarki
│   ├── value_test.go                 # Pengujian type safety context.WithValue
│   ├── cancel_test.go                # Pengujian cancel manual & cancel cause
│   ├── timeout_test.go               # Pengujian pembatalan timeout otomatis
│   └── deadline_test.go              # Pengujian titik waktu absolut & inheritansi deadline
├── service/                          # Implementasi Real-World Service
│   ├── middleware/
│   │   └── context_middleware.go     # Injeksi Request-ID & Auth Claims (WithValue)
│   ├── repository/
│   │   └── user_repository.go        # Simulasi query DB dengan pembatalan context
│   ├── client/
│   │   └── external_payment.go       # Simulasi 3rd party API dengan timeout SLA
│   ├── worker/
│   │   └── batch_worker.go           # Batch scheduler dengan WithDeadline & WithCancel
│   ├── handler/
│   │   └── order_handler.go          # HTTP Handler mengorkestrasi seluruh flow context
│   ├── server.go                     # HTTP Server dengan Graceful Shutdown (WithCancel)
│   └── service_test.go               # Pengujian integrasi end-to-end service
├── cmd/server/
│   └── main.go                       # Entry point aplikasi runnable
├── issue.md                          # Dokumen planning & issue GitHub
└── README.md                         # Dokumentasi utama proyek
```

---

## 🧪 Menjalankan Pengujian Otomatis (Automated Tests)

Seluruh konsep dasar dan alur layanan dilengkapi dengan unit & integration test. Jalankan perintah berikut untuk menguji seluruh suite dengan pemeriksaan *race condition*:

```bash
go test -v -race ./...
```

Contoh keluaran sukses:
```text
=== RUN   TestParentCancellationPropagatesToChildren
--- PASS: TestParentCancellationPropagatesToChildren (0.00s)
=== RUN   TestContextWithValueTypeSafety
--- PASS: TestContextWithValueTypeSafety (0.00s)
=== RUN   TestWorkerGoroutineCancellation
--- PASS: TestWorkerGoroutineCancellation (0.08s)
=== RUN   TestContextWithTimeoutCancelsSlowOperation
--- PASS: TestContextWithTimeoutCancelsSlowOperation (0.04s)
=== RUN   TestServiceNormalOrder
--- PASS: TestServiceNormalOrder (0.10s)
=== RUN   TestServiceSlowDBTimeout
--- PASS: TestServiceSlowDBTimeout (0.30s)
PASS
ok      belajar-golang-context/basics   1.332s
ok      belajar-golang-context/service  1.791s
```

---

## 🖥️ Menjalankan Real-World Service Demo

Aplikasi dapat dijalankan secara langsung menggunakan:

```bash
go run ./cmd/server
```

Saat aplikasi dijalankan, server HTTP akan aktif di port `8080`, sekaligus menjalankan demonstrasi background batch worker dengan `context.WithDeadline`.

### Skenario Pengujian Endpoint (cURL)

#### 1. Skenario Normal (Sukses)
Request normal dengan injeksi identitas pengguna:
```bash
curl -i -H "X-User-ID: 101" -H "X-User-Role: premium" http://localhost:8080/order
```
*Hasil: Status `200 OK`, `Request-ID` dihasilkan, transaksi berhasil.*

#### 2. Skenario Database Timeout (`context.WithTimeout`)
Simulasi query database yang macet / lambat:
```bash
curl -i -H "X-User-ID: 101" "http://localhost:8080/order?simulate_slow_db=true"
```
*Hasil: Status `504 Gateway Timeout` setelah melewati batas waktu handler.*

#### 3. Skenario Payment SLA Timeout
Simulasi panggilan ke 3rd-party Payment Gateway yang melebihi SLA:
```bash
curl -i -H "X-User-ID: 101" "http://localhost:8080/order?simulate_slow_payment=true"
```
*Hasil: Status `504 Gateway Timeout` dipicu oleh client-side timeout.*

#### 4. Skenario Missing Authentication
Request tanpa menyertakan user ID di header:
```bash
curl -i http://localhost:8080/order
```
*Hasil: Status `401 Unauthorized`.*

#### 5. Skenario Graceful Shutdown
Tekan `Ctrl+C` pada terminal tempat server berjalan:
```text
^CMenerima sinyal shutdown OS (interrupt). Memulai Graceful Shutdown...
Server berhasil dihentikan secara bersih (graceful shutdown complete).
```
