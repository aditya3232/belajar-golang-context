# Issue: Panduan Komprehensif Dasar-Dasar Context di Go & Implementasi Real-World Service

## 1. Deskripsi & Latar Belakang
Package `context` merupakan salah satu komponen paling fundamental dalam ekosistem Go modern, khususnya dalam pembuatan web service, microservices, dan sistem terdistribusi. Context bertanggung jawab atas:
- Manajemen lifecycle request dan goroutine.
- Propagasi pembatalan (*cancellation propagation*) antar layer aplikasi.
- Pembatasan waktu proses (*timeout & deadline*) untuk mencegah resource exhaustion dan request macet (*hanging*).
- Pengiriman data scoped-request (misal: Request ID, Authentication Claims, Tracing Context).

Tujuan project ini adalah menyusun dokumentasi edukatif yang komprehensif serta kode implementasi runnable yang memodelkan use-case dunia nyata (real-world service).

---

## 2. Cakupan Materi & Topik Utama

### A. Parent & Child Context
- **Konsep Tree Context**: Bagaimana konteks root (`context.Background()` / `context.TODO()`) menurunkan child context.
- **Cancellation Propagation**: Pembatalan pada parent otomatis membatalkan semua child, tetapi pembatalan pada child tidak membatalkan parent.
- **Visualisasi Hierarki**: Diagram alir pohon context dan penyebaran signal `Done()`.

### B. Context with Value (`context.WithValue`)
- **Fungsi & Konsep**: Menyimpan data berumur pendek yang terikat pada satu siklus request (*request-scoped*).
- **Best Practices & Type Safety**:
  - Menghindari penggunaan tipe data primitif (`string`) sebagai key guna mencegah benturan (*key collision*).
  - Menggunakan tipe kustom unexported (contoh: `type ctxKey string`).
  - Helper function untuk injeksi (*setter*) dan ekstraksi (*getter*) value yang aman (type assertion).
- **Antipattern**:
  - Dilarang melewatkan parameter fungsi opsional, koneksi database, atau mutable state ke dalam context.

### C. Context with Cancel (`context.WithCancel`)
- **Fungsi & Mekanisme**: Menghentikan eksekusi goroutine secara manual menggunakan fungsi `cancel()`.
- **Mencegah Goroutine Leak**:
  - Wajib memanggil `defer cancel()` untuk membebaskan alokasi timer dan child context.
  - Mendengarkan channel `<-ctx.Done()` di dalam goroutine worker/loop.
  - Memeriksa `ctx.Err()` (`context.Canceled`).
- **Use Case Nyata**: Graceful shutdown pada HTTP Server saat menerima sinyal OS (`SIGINT`/`SIGTERM`) atau pembatalan concurrent worker saat satu worker gagal (*fail-fast pattern*).

### D. Context with Timeout (`context.WithTimeout`)
- **Fungsi**: Membatalkan context secara otomatis setelah durasi tertentu berlalu (contoh: 2 detik).
- **Mekanisme**: Membaca `ctx.Err()` yang mengembalikan `context.DeadlineExceeded`.
- **Use Case Nyata**: HTTP client call ke 3rd party API, query lambat ke database, atau panggilan RPC yang memerlukan SLA ketat.

### E. Context with Deadline (`context.WithDeadline`)
- **Fungsi**: Membatalkan context pada titik waktu absolut tertentu (contoh: `time.Now().Add(...)` atau pukul 23:59:00).
- **Perbedaan dengan WithTimeout**: WithTimeout bersifat duratif (*relative duration*), sedangkan WithDeadline bersifat titik waktu absolut (*absolute timestamp*).
- **Use Case Nyata**: Batch job / cron scheduler yang memiliki hard cutoff window eksekusi sebelum proses berikutnya dimulai.

---

## 3. Rencana Arsitektur & Struktur Direktori

```text
belajar-golang-context/
├── issue.md                          # Dokumentasi planning (file ini)
├── README.md                         # Panduan utama & ringkasan pembelajaran
├── docs/                             # Materi panduan mendalam
│   ├── 01-parent-child.md
│   ├── 02-context-with-value.md
│   ├── 03-context-with-cancel.md
│   ├── 04-context-with-timeout.md
│   ├── 05-context-with-deadline.md
│   └── 06-best-practices.md
├── basics/                           # Unit tests & contoh dasar terisolasi
│   ├── parent_child_test.go
│   ├── value_test.go
│   ├── cancel_test.go
│   ├── timeout_test.go
│   └── deadline_test.go
└── service/                          # Contoh nyata: Mini Web Service
    ├── middleware/
    │   └── request_id.go             # Implementasi context.WithValue
    ├── repository/
    │   └── user_repository.go        # Simulasi query DB dengan context.WithTimeout
    ├── client/
    │   └── payment_client.go         # Simulasi HTTP Client 3rd party call dengan timeout
    ├── worker/
    │   └── batch_processor.go        # Simulasi batch worker dengan context.WithDeadline & WithCancel
    ├── handler/
    │   └── user_handler.go           # HTTP Handler yang mengorkestrasi context
    ├── server.go                     # Server HTTP dengan Graceful Shutdown (WithCancel)
    └── main.go                       # Entry point aplikasi runnable
```

---

## 4. Tahapan Eksekusi (Action Items)

- [ ] **Fase 1: Inisialisasi Modul**
  - Buat `go.mod` dengan nama modul `belajar-golang-context`.
- [ ] **Fase 2: Implementasi Contoh Dasar (`basics/`)**
  - Tulis unit tests idiomatis untuk kelima konsep agar mudah di-run dengan `go test -v ./basics/...`.
- [ ] **Fase 3: Implementasi Real-World Service (`service/`)**
  - Bangun pipeline HTTP lengkap:
    1. Middleware menyuntikkan `Request-ID` dan `Auth User` ke context (`WithValue`).
    2. Handler menerima request dan mem-pass context ke service dan repository (`Parent & Child`).
    3. Repository mendeteksi client disconnect atau query timeout (`WithTimeout`).
    4. Worker batch scheduler menjalankan task dengan hard deadline (`WithDeadline`).
    5. Graceful shutdown HTTP server dengan sinyal interrupt OS (`WithCancel`).
- [ ] **Fase 4: Penyusunan Dokumen Panduan (`docs/` & `README.md`)**
  - Tulis penjelasan konseptual, analogi mudah dipahami, diagram alir visual (Mermaid), serta panduan antipattern vs best practices.
- [ ] **Fase 5: Validasi & Pengujian**
  - Jalankan seluruh automated tests (`go test -v ./...`).
  - Uji server dan simulasi skenario timeout/cancellation secara manual (`curl` / script).

---

## 5. Kriteria Penerimaan (Acceptance Criteria)

1. Semua konsep (Parent-Child, WithValue, WithCancel, WithTimeout, WithDeadline) memiliki contoh kode runnable dan unit test yang lolos (PASS).
2. Penjelasan materi mudah dipahami oleh developer pemula hingga menengah, disajikan dalam Bahasa Indonesia yang lugas dan terstruktur.
3. Contoh real-world service merepresentasikan pola industri (clean layering: handler, middleware, repository, client, worker).
4. Tidak ada goroutine leak (semua `cancel()` di-defer dengan benar dan channel `Done()` di-handle).
