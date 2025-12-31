# percentiles

```go
import "github.com/cuberat-go/percentiles"
```

The percentile module computes percentile metrics (e.g., p50, p75, p95) on the input without assuming the data will fit in RAM.

```go
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
```

Note that the `Add()` and `AddSlice()` methods are thread-safe, but
`Percentiles()` and `Percentile()` are not (yet). However, for better performance
when adding data from multiple threads, it is recommended to write to a channel
that is then processed by another goroutine that calls `Add()` or `AddSlice()`.
