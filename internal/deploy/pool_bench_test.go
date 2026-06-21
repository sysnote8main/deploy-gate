package deploy

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// BenchmarkPool benchmarks the bounded worker pool.
func BenchmarkPool(b *testing.B) {
	pool := NewPool(4) // 4 concurrent workers
	defer pool.Shutdown()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pool.Run("/usr/bin/true")
	}
	// Wait for all tasks to complete.
	pool.Shutdown()
}

// BenchmarkPoolUnbounded benchmarks unbounded goroutines (old behavior).
func BenchmarkPoolUnbounded(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		go func() {
			cmd := __execCommand("/usr/bin/true")
			_ = cmd
		}()
	}
}

// BenchmarkPoolThroughput measures tasks completed per second.
func BenchmarkPoolThroughput(b *testing.B) {
	for _, workers := range []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			pool := NewPool(workers)

			var completed atomic.Int64
			var wg sync.WaitGroup

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				wg.Add(1)
				pool.sem <- struct{}{}
				completed.Add(1)

				go func() {
					defer wg.Done()
					defer func() { <-pool.sem }()
					// Simulate work.
					_ = make([]byte, 1024)
				}()
			}

			wg.Wait()
			pool.Shutdown()
		})
	}
}

// BenchmarkPoolMemory benchmarks memory usage under load.
func BenchmarkPoolMemory(b *testing.B) {
	pool := NewPool(4)
	defer pool.Shutdown()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pool.Run("/usr/bin/true")
	}

	// Measure active count.
	_ = pool.Active()
}

// __execCommand is a placeholder for benchmark (avoids actual exec).
func __execCommand(script string) interface{ Run() error } {
	return &dummyCmd{script: script}
}

type dummyCmd struct {
	script string
}

func (d *dummyCmd) Run() error {
	return nil
}
