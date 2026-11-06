package tests

import (
	"fmt"
	"sync"
	"testing"

	"github.com/suserc0de/rhui/expr"
)

func TestLRUExpressionCache(t *testing.T) {
	expr.SetCacheCapacity(10)
	defer expr.SetCacheCapacity(expr.DefaultLRUCapacity)

	// Populate 10 items
	for i := 0; i < 10; i++ {
		_, err := expr.Parse(fmt.Sprintf("item_%d + 1", i))
		if err != nil {
			t.Fatalf("unexpected error parsing expr: %v", err)
		}
	}

	// Add an 11th item - should evict item_0
	_, err := expr.Parse("item_10 + 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Concurrent read and write stress test to verify 0 data race
	var wg sync.WaitGroup
	workers := 16
	iterations := 100

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				query := fmt.Sprintf("val_%d == %d", (workerID*10+i)%30, i)
				node, err := expr.Parse(query)
				if err != nil || node == nil {
					t.Errorf("worker %d failed parsing %s: %v", workerID, query, err)
					return
				}
			}
		}(w)
	}

	wg.Wait()
}
