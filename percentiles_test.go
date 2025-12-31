package percentiles_test

import (
	// Built-in/core modules
	"fmt"
	"testing"

	// Third-party modules.
	"github.com/stretchr/testify/assert"

	// First-party modules.
	"github.com/cuberat-go/percentiles"
)

func TestPercent(t *testing.T) {
	t.Run("default_threshold", func(st *testing.T) {
		runWithThreshold(st, 0)
	})

	t.Run("threshold=5", func(st *testing.T) {
		runWithThreshold(st, 5)
	})

}

func runWithThreshold(t *testing.T, threshold int64) {
	p := percentiles.NewStats()
	p.SetFileThreshold(threshold)
	for i := range 100 {
		// Add values in reverse order to ensure sorting works.
		p.Add(float64(100 - i))
	}

	result := p.Percentiles()
	expected := []float64{50.0, 75.0, 90.0, 95.0, 99.0}

	assert.Equal(t, expected, result,
		"percentiles do not match expected values")

	result90 := p.Percentile(90.0)

	assert.Equal(t, 90.0, result90,
		"90th percentile does not match expected value")

	result100 := p.Percentile(100.0)

	assert.Equal(t, 100.0, result100)

	p.Done()
}

func ExampleStats_Percentiles() {
	p := percentiles.NewStats()

	for i := range 100 {
		p.Add(float64(100 - i))
	}

	// Returns p50, p75, p90, p95, and p99.
	results := p.Percentiles()
	fmt.Printf("percentiles: %v\n", results)

	// Returns a specific percentile:
	result87 := p.Percentile(87.0)

	fmt.Printf("87th percentile: %v\n", result87)

	p.Done()

	// Output:
	// percentiles: [50 75 90 95 99]
	// 87th percentile: 87
}
