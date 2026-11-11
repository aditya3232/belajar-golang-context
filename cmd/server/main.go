package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"belajar-golang-context/service"
	"belajar-golang-context/service/worker"
)

func main() {
	log.Println("==================================================")
	log.Println("🚀 Belajar Golang Context: Real-World Service Demo")
	log.Println("==================================================")

	// Demonstrasi background batch worker dengan WithDeadline
	go runBackgroundWorkerDemo()

	// Jalankan HTTP Server
	server := service.NewAppServer(8080)
	if err := server.Start(); err != nil {
		log.Fatalf("Fatal error pada server: %v", err)
	}
}

func runBackgroundWorkerDemo() {
	time.Sleep(1 * time.Second)
	log.Println("\n[Batch Worker] Menjalankan simulasi batch settlement dengan context.WithDeadline...")

	items := []worker.BatchItem{
		{ID: 1, Title: "Settlement Toko A"},
		{ID: 2, Title: "Settlement Toko B"},
		{ID: 3, Title: "Settlement Toko C"},
		{ID: 4, Title: "Settlement Toko D"},
		{ID: 5, Title: "Settlement Toko E"},
	}

	// Batas waktu deadline hanya 80ms, sementara 5 item butuh 5x30ms = 150ms
	// Ini akan mendemonstrasikan penghentian otomatis saat deadline tercapai
	deadline := time.Now().Add(80 * time.Millisecond)
	result, err := worker.RunNightlySettlementTask(context.Background(), deadline, items)

	if err != nil {
		log.Printf("[Batch Worker] Selesai dengan catatan: %v\n", err)
	}
	log.Printf("[Batch Worker] Total diproses: %d dari %d item (Completed: %t)\n\n",
		result.ProcessedCount, result.TotalItems, result.Completed)

	fmt.Println("--------------------------------------------------")
	fmt.Println("Endpoint yang tersedia:")
	fmt.Println("  1. GET/POST /health")
	fmt.Println("  2. GET/POST /order                (Skenario Normal)")
	fmt.Println("  3. GET/POST /order?simulate_slow_db=true      (DB Timeout)")
	fmt.Println("  4. GET/POST /order?simulate_slow_payment=true (Payment SLA Timeout)")
	fmt.Println("--------------------------------------------------")
}
