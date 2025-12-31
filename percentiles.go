package percentiles

import (
	// Built-in/core modules.
	"bufio"
	"iter"
	"os"
	"slices"
	"strconv"
)

// Structure for object tracking percentile statistics.
type Stats struct {
	// Slice holding the data. This will not be populated if we are writing to a
	// file.
	data []float64

	// File handle for data. Only non-nil if we are storing data in a file.
	dataFh *os.File

	// Indicates whether the input data has been sorted.
	dataSorted bool

	// Length of the data added.
	dataLen int64

	// Data length threshold at which we write to a file instead of keeping data
	// in RAM.
	fileThreshold int64
}

// Creates and returns a new Stats instance.
func NewStats() *Stats {
	return &Stats{}
}

// Add a value to be tracked.
func (s *Stats) Add(value float64) {
	s.data = append(s.data, value)
	s.dataLen++
}

// Add a slice of values to be tracked.
func (s *Stats) AddSlice(values []float64) {
	s.data = append(s.data, values...)
	s.dataLen += int64(len(values))
}

// Returns commonly-used percentiles as a slice of float64: p50, p75, p90, p95,
// and p99.
func (s *Stats) Percentiles() []float64 {
	s.sortData()

	cntThresholds := []int64{
		int64(float64(s.dataLen) * 0.5),
		int64(float64(s.dataLen) * 0.75),
		int64(float64(s.dataLen) * 0.90),
		int64(float64(s.dataLen) * 0.95),
		int64(float64(s.dataLen) * 0.99),
	}

	percentiles := make([]float64, len(cntThresholds))

	cnt := int64(0)
	threshIndex := 0
	lastVal := float64(0)
	lastThreshIndex := -1
	for v := range s.iterateData() {
		cnt++
		if cnt >= cntThresholds[threshIndex] {
			if cnt > cntThresholds[threshIndex] {
				percentiles[threshIndex] = lastVal
			} else {
				percentiles[threshIndex] = v
			}
			lastThreshIndex = threshIndex
			threshIndex++
			if threshIndex >= len(cntThresholds) {
				break
			}
		}
		lastVal = v
	}

	if lastThreshIndex < len(cntThresholds) {
		for i := threshIndex; i < len(cntThresholds); i++ {
			percentiles[i] = percentiles[lastThreshIndex]
		}
	}

	return percentiles
}

// Returns the specified percentile (p) value. E.g., p=90.0 for the 90th
// percentile.
func (s *Stats) Percentile(p float64) float64 {
	s.sortData()

	cnt := int64(0)
	threshold := int64(float64(s.dataLen) * p / 100.0)
	lastVal := float64(0)
	for v := range s.iterateData() {
		cnt++
		if cnt >= threshold {
			if cnt > threshold {
				return lastVal
			} else {
				return v
			}
		}
		lastVal = v
	}

	return lastVal
}

// Clean up resources.
func (s *Stats) Done() {
	if s.dataFh != nil {
		s.dataFh.Close()
		os.Remove(s.dataFh.Name())
	}
}

func (s *Stats) iterateData() iter.Seq[float64] {
	if s.dataFh != nil {
		s.dataFh.Seek(0, 0)

		return func(yield func(float64) bool) {
			for scanner := bufio.NewScanner(s.dataFh); scanner.Scan(); {
				v, _ := strconv.ParseFloat(scanner.Text(), 64)
				if !yield(v) {
					return
				}
			}
		}
	}

	return func(yield func(float64) bool) {
		for _, v := range s.data {
			if !yield(v) {
				return
			}
		}
	}
}

func (s *Stats) sortData() {
	if s.dataFh != nil {
		// TODO: Implement sorting from file.
	}

	if !s.dataSorted {
		slices.Sort(s.data)
		s.dataSorted = true
	}
}
