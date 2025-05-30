package errmgr

import (
	"fmt"
	"github.com/olekukonko/errors"
	"testing"
)

// BenchmarkTemplateError measures the performance of creating templated errors.
func BenchmarkTemplateError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ErrDBConnection(fmt.Sprintf("connection failed %d", i))
		err.Free()
	}
}

// BenchmarkCodedError measures the performance of creating coded errors.
func BenchmarkCodedError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ErrValidationFailed(fmt.Sprintf("field %d", i))
		err.Free()
	}
}

// BenchmarkCategorizedError measures the performance of creating categorized errors.
func BenchmarkCategorizedError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := NetworkError(fmt.Sprintf("host %d", i))
		err.Free()
	}
}

// BenchmarkCallableError measures the performance of creating custom callable errors.
func BenchmarkCallableError(b *testing.B) {
	fn := Tracked("custom", func(args ...interface{}) *errors.Error {
		return errors.New(fmt.Sprintf("custom %v", args[0]))
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := fn(i)
		err.Free()
	}
}

// BenchmarkMetrics measures the performance of retrieving error metrics.
func BenchmarkMetrics(b *testing.B) {
	for i := 0; i < 100; i++ {
		err := ErrDBConnection(fmt.Sprintf("test %d", i))
		err.Free()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Metrics()
	}
}

func BenchmarkMonitorWithClosedChannel(b *testing.B) {
	Reset()
	SetThreshold("BenchError", 1)

	// Create and close monitor to test closed channel case
	monitor := NewMonitor("BenchError")
	monitor.Close()

	errFunc := Define("BenchError", "bench test %d")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := errFunc(i)
		err.Free()
	}
}
