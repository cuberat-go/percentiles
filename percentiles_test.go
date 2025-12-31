package percentiles_test

import (
	// Built-in/core modules
	"testing"

	// Third-party modules.
	"github.com/stretchr/testify/assert"

	// First-party modules.
	"github.com/cuberat-go/percentiles"
)

func TestPercent(t *testing.T) {
	p := percentiles.NewStats()
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
}
