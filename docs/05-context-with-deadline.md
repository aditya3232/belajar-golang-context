# 05. Context with Deadline (`context.WithDeadline`)

`context.WithDeadline` menetapkan batas waktu absolut (*point in time*) berupa `time.Time` kapan sebuah context harus dibatalkan.

---

## 1. Perbedaan Mendasar: `WithDeadline` vs `WithTimeout`

Secara fungsional keduanya mirip, namun berbeda dalam cara mendefinisikan batas waktu:

| Fitur | `context.WithTimeout` | `context.WithDeadline` |
| :--- | :--- | :--- |
| **Tipe Parameter** | `time.Duration` (contoh: `2 * time.Second`) | `time.Time` (contoh: `time.Date(2026, 11, 11, 23, 59, 0, 0, time.UTC)`) |
| **Sifat Waktu** | Relatif (durasi sejak fungsi dipanggil) | Absolut (titik waktu kalender tertentu) |
| **Implementasi Internal Go** | Memanggil `WithDeadline(parent, time.Now().Add(timeout))` | Logika timer inti Go standar |

Faktanya, di dalam source code standard library Go, `WithTimeout` hanyalah pembungkus tipis di atas `WithDeadline`:
```go
// Dari package context standar Go:
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}
```

---

## 2. Aturan Emas Deadline: Child Tidak Bisa Memperpanjang Deadline Parent

> [!IMPORTANT]
> **Aturan Inheritansi Deadline**:
> Jika Parent Context memiliki deadline pukul **14:00**, dan sebuah Child Context dibuat dengan deadline pukul **15:00**, maka deadline Child Context akan **DIPAKSA menjadi pukul 14:00**.
> 
> Anak tidak pernah bisa hidup lebih lama daripada orang tuanya!

Namun, jika Child menentukan deadline yang **lebih cepat** (misal pukul **13:30**), maka Child akan selesai pada pukul 13:30 tanpa mempengaruhi Parent.

```go
func TestDeadlineInheritance() {
	now := time.Now()
	parentDeadline := now.Add(1 * time.Minute)
	childAttempted := now.Add(5 * time.Minute)

	parentCtx, parentCancel := context.WithDeadline(context.Background(), parentDeadline)
	defer parentCancel()

	childCtx, childCancel := context.WithDeadline(parentCtx, childAttempted)
	defer childCancel()

	childActualDeadline, _ := childCtx.Deadline()
	// childActualDeadline akan SAMA DENGAN parentDeadline (1 menit), BUKAN 5 menit!
}
```

---

## 3. Contoh Kasus Nyata: Batch Settlement dengan Cutoff Window

Misalkan Anda memiliki background task yang memproses ribuan data transaksi bank setiap malam. Bank memiliki hard cutoff pukul 23:55:00 WIB di mana seluruh transaksi setelah jam tersebut akan ditolak.

Dengan `WithDeadline`, kita dapat memastikan pemrosesan berhenti secara tertib sebelum jendela waktu ditutup:

```go
func RunDailySettlement(ctx context.Context, cutoffTime time.Time, transactions []Transaction) {
	// Buat context dengan deadline absolut
	deadlineCtx, cancel := context.WithDeadline(ctx, cutoffTime)
	defer cancel()

	for _, trx := range transactions {
		// Cek apakah cutoff sudah tercapai sebelum memproses item berikutnya
		select {
		case <-deadlineCtx.Done():
			log.Printf("Batas cutoff %v tercapai! Menghentikan sisa transaksi secara aman: %v\n", 
				cutoffTime, deadlineCtx.Err())
			return
		default:
			// Lanjutkan proses transaksi
			processTransaction(deadlineCtx, trx)
		}
	}

	log.Println("Seluruh transaksi berhasil diselesaikan sebelum cutoff.")
}
```

---

## 4. Kasus Khusus: Deadline di Masa Lalu (*Past Deadline*)

Jika Anda memasukkan deadline yang sudah lewat (misal `time.Now().Add(-1 * time.Second)`), Go tidak akan error atau panic, melainkan **langsung menutup channel `Done()` seketika** pada saat context tersebut dibuat.
