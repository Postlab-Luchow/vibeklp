package llm

import "fmt"

// BatchProcessor handles batching of extraction tasks
type BatchProcessor struct {
	BatchSize int
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize int) *BatchProcessor {
	if batchSize <= 0 {
		batchSize = 5
	}
	return &BatchProcessor{
		BatchSize: batchSize,
	}
}

// CreateBatches splits items into batches of specified size
func (bp *BatchProcessor) CreateBatches(items []string) [][]string {
	var batches [][]string
	for i := 0; i < len(items); i += bp.BatchSize {
		end := i + bp.BatchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}
	return batches
}

// CombineHTMLBlocks combines multiple HTML blocks with separators for batch processing
func (bp *BatchProcessor) CombineHTMLBlocks(blocks []string) string {
	var combined string
	for i, block := range blocks {
		if i > 0 {
			combined += "\n\n---EVENT_SEPARATOR---\n\n"
		}
		combined += fmt.Sprintf("<!-- ITEM %d -->\n%s", i+1, block)
	}
	return combined
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
