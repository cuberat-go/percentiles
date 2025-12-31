package percentiles

import (
	// Built-in/core modules.
	"bufio"
	"fmt"
	"iter"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"sync"
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

	// Mutex used to make data additions thread-safe.
	dataLock *sync.Mutex

	// Data length threshold at which we write to a file instead of keeping data
	// in RAM.
	fileThreshold int64
}

// Creates and returns a new Stats instance.
func NewStats() *Stats {
	return &Stats{fileThreshold: 100000, dataLock: &sync.Mutex{}}
}

// Set the data length threshold at which added data will be stored in a file,
// rather than in RAM. The default is 100,000 data points.
func (s *Stats) SetFileThreshold(threshold int64) {
	s.fileThreshold = threshold
}

// Add a value to be tracked. Add() is thread-safe, but for better performance
// when adding data from multiple threads, it is recommended to write to a
// channel that is then processed by another goroutine that calls Add().
func (s *Stats) Add(value float64) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	s.dataLen++

	if s.dataFh != nil || s.dataLen > s.fileThreshold {
		if s.dataFh == nil {
			s.setupDataFile()
		}
		fmt.Fprintf(s.dataFh, "%f\n", value)
		return
	}

	s.data = append(s.data, value)

}

// Add a slice of values to be tracked. AddSlice() is thread-safe, but for
// better performance when adding data from multiple threads, it is recommended
// to write to a channel that is then processed by another goroutine that calls
// AddSlice().
func (s *Stats) AddSlice(values []float64) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	s.dataLen += int64(len(values))

	if s.dataFh != nil || s.dataLen > s.fileThreshold {
		if s.dataFh == nil {
			s.setupDataFile()
		}
		for _, value := range values {
			fmt.Fprintf(s.dataFh, "%f\n", value)
		}
		return
	}

	s.data = append(s.data, values...)
}

func (s *Stats) setupDataFile() error {
	if s.dataFh != nil {
		// Already set up.
		return nil
	}

	var err error

	s.dataFh, err = os.CreateTemp("", "pct_data_")
	for _, value := range s.data {
		fmt.Fprintf(s.dataFh, "%f\n", value)
	}
	s.data = nil

	return err
}

// Returns commonly-used percentiles as a slice of float64: p50 (median), p75,
// p90, p95, and p99.
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

// Free up resources.
func (s *Stats) Done() {
	if s.dataFh != nil {
		s.dataFh.Close()
		os.Remove(s.dataFh.Name())
	}
	s.data = nil
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

// Sorts the input data. If the data is in a file, the external sort command is
// called.
func (s *Stats) sortData() error {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	if s.dataSorted {
		return nil
	}

	if s.dataFh != nil {
		s.dataFh.Close()
		oldFileName := s.dataFh.Name()
		defer os.Remove(oldFileName)

		newFh, err := os.CreateTemp("", "pct_data_sort_")
		if err != nil {
			return err
		}

		if err = s.callSort(s.dataFh.Name(), newFh.Name()); err != nil {
			return err
		}

		s.dataFh = newFh
		s.dataSorted = true

		return nil
	}

	if !s.dataSorted {
		slices.Sort(s.data)
		s.dataSorted = true
	}

	return nil
}

// Calls the external sort command to sort floating point numbers in infile and
// write the results to outfile.
func (s *Stats) callSort(infile, outfile string) error {
	sort_path := "/usr/bin/sort"
	cmd := &exec.Cmd{
		Path: sort_path,
		Args: []string{sort_path, "--mergesort", "-n", "-o", outfile, infile},
	}

	return cmd.Run()
}
