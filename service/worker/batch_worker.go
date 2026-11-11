package worker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// BatchItem merepresentasikan data yang akan diproses dalam batch
type BatchItem struct {
	ID    int
	Title string
}

// BatchResult rangkuman hasil eksekusi batch
type BatchResult struct {
	ProcessedCount int
	TotalItems     int
	Completed      bool
	StoppedReason  error
}

// BatchProcessor memproses antrian tugas dengan batas waktu absolut (Deadline)
type BatchProcessor struct{}

// NewBatchProcessor membuat instance baru BatchProcessor
func NewBatchProcessor() *BatchProcessor {
	return &BatchProcessor{}
}

// ProcessWithDeadline memproses sekumpulan item sebelum titik waktu deadline tercapai
func (p *BatchProcessor) ProcessWithDeadline(ctx context.Context, items []BatchItem, itemDelay time.Duration) BatchResult {
	result := BatchResult{
		TotalItems: len(items),
	}

	for i, item := range items {
		// Periksa apakah context (deadline atau cancel) sudah terpicu sebelum memproses item berikutnya
		select {
		case <-ctx.Done():
			result.StoppedReason = ctx.Err()
			result.Completed = false
			return result
		default:
		}

		// Simulasi pengerjaan item
		select {
		case <-time.After(itemDelay):
			result.ProcessedCount++
			_ = fmt.Sprintf("Item %d (%s) diproses", item.ID, item.Title)
		case <-ctx.Done():
			// Terpotong di tengah pemrosesan item
			result.StoppedReason = ctx.Err()
			result.Completed = false
			return result
		}

		// Jika ini item terakhir, tandai completed
		if i == len(items)-1 {
			result.Completed = true
		}
	}

	return result
}

// RunNightlySettlementTask simulasi penjadwalan task batch dengan context.WithDeadline
func RunNightlySettlementTask(parentCtx context.Context, deadline time.Time, items []BatchItem) (BatchResult, error) {
	// Buat context dengan batas waktu absolut
	deadlineCtx, cancel := context.WithDeadline(parentCtx, deadline)
	defer cancel()

	processor := NewBatchProcessor()
	result := processor.ProcessWithDeadline(deadlineCtx, items, 30*time.Millisecond)

	if !result.Completed {
		if errors.Is(result.StoppedReason, context.DeadlineExceeded) {
			return result, fmt.Errorf("batch processing terhenti karena mencapai cutoff window deadline: %w", result.StoppedReason)
		}
		if errors.Is(result.StoppedReason, context.Canceled) {
			return result, fmt.Errorf("batch processing dibatalkan secara manual: %w", result.StoppedReason)
		}
	}

	return result, nil
}
