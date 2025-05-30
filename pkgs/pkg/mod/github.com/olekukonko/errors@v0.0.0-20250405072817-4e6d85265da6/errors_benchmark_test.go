package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"testing"
)

// Basic Error Creation Benchmarks
// These benchmarks measure the performance of creating basic errors with and without
// pooling, compared to standard library equivalents for baseline reference.

// BenchmarkBasic_New measures the creation and pooling of a new error.
func BenchmarkBasic_New(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := New("test error") // Create and pool a new error
		err.Free()
	}
}

// BenchmarkBasic_NewNoFree measures error creation without pooling.
func BenchmarkBasic_NewNoFree(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New("test error") // Create error without returning to pool
	}
}

// BenchmarkBasic_StdlibComparison measures standard library error creation as a baseline.
func BenchmarkBasic_StdlibComparison(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.New("test error") // Baseline using standard library errors.New
	}
}

// BenchmarkBasic_StdErrorComparison measures the package's Std wrapper for errors.New.
func BenchmarkBasic_StdErrorComparison(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Std("test error") // Baseline using package’s Std wrapper for errors.New
	}
}

// BenchmarkBasic_StdfComparison measures the package's Stdf wrapper for fmt.Errorf.
func BenchmarkBasic_StdfComparison(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Stdf("test error %d", i) // Baseline using package’s Stdf wrapper for fmt.Errorf
	}
}

// Stack Trace Benchmarks
// These benchmarks evaluate the performance of stack trace operations, including
// capturing and generating stack traces for error instances.

// BenchmarkStack_WithStack measures adding a stack trace to an error.
func BenchmarkStack_WithStack(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := New("test").WithStack() // Add stack trace to an error
		err.Free()
	}
}

// BenchmarkStack_Trace measures creating an error with a stack trace.
func BenchmarkStack_Trace(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Trace("test error") // Create error with stack trace
		err.Free()
	}
}

// BenchmarkStack_Capture measures generating a stack trace from an existing error.
func BenchmarkStack_Capture(b *testing.B) {
	err := New("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Stack() // Generate stack trace from existing error
	}
	err.Free()
}

// BenchmarkCaptureStack measures capturing a raw stack trace.
func BenchmarkCaptureStack(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack := captureStack(0) // Capture raw stack trace
		if stack != nil {
			runtime.KeepAlive(stack) // Ensure stack isn’t optimized away
		}
	}
}

// Context Operation Benchmarks
// These benchmarks assess the performance of adding context to errors, testing
// small context (array-based), map-based, and reuse scenarios.

// BenchmarkContext_Small measures adding context within the smallContext limit.
func BenchmarkContext_Small(b *testing.B) {
	err := New("base")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.With("key", i).With("key2", i+1) // Add two key-value pairs within smallContext limit
	}
	err.Free()
}

// BenchmarkContext_Map measures adding context exceeding smallContext capacity.
func BenchmarkContext_Map(b *testing.B) {
	err := New("base")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.With("k1", i).With("k2", i+1).With("k3", i+2) // Exceed smallContext, forcing map usage
	}
	err.Free()
}

// BenchmarkContext_Reuse measures adding to an existing context.
func BenchmarkContext_Reuse(b *testing.B) {
	err := New("base").With("init", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.With("key", i) // Add to existing context
	}
	err.Free()
}

// Error Wrapping Benchmarks
// These benchmarks measure the cost of wrapping errors, both shallow and deep chains.

// BenchmarkWrapping_Simple measures wrapping a single base error.
func BenchmarkWrapping_Simple(b *testing.B) {
	base := New("base")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := New("wrapper").Wrap(base) // Wrap a single base error
		err.Free()
	}
	base.Free()
}

// BenchmarkWrapping_Deep measures unwrapping a 10-level deep error chain.
func BenchmarkWrapping_Deep(b *testing.B) {
	var err *Error
	for i := 0; i < 10; i++ {
		err = New("level").Wrap(err) // Build a 10-level deep error chain
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Unwrap() // Unwrap the deep chain
	}
	err.Free()
}

// Type Assertion Benchmarks
// These benchmarks evaluate the performance of type assertions (Is and As) on wrapped errors.

// BenchmarkTypeAssertion_Is measures checking if an error matches a target.
func BenchmarkTypeAssertion_Is(b *testing.B) {
	target := Named("target")
	err := New("wrapper").Wrap(target)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Is(err, target) // Check if error matches target
	}
	target.Free()
}

// BenchmarkTypeAssertion_As measures extracting a target from an error chain.
func BenchmarkTypeAssertion_As(b *testing.B) {
	err := New("wrapper").Wrap(Named("target"))
	var target *Error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = As(err, &target) // Extract target from error chain
	}
	if target != nil {
		target.Free()
	}
}

// Serialization Benchmarks
// These benchmarks test JSON serialization performance with and without stack traces.

// BenchmarkSerialization_JSON measures serializing an error with context to JSON.
func BenchmarkSerialization_JSON(b *testing.B) {
	err := New("test").With("key", "value").With("num", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(err) // Serialize error with context
	}
}

// BenchmarkSerialization_JSONWithStack measures serializing an error with stack trace to JSON.
func BenchmarkSerialization_JSONWithStack(b *testing.B) {
	err := Trace("test").With("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(err) // Serialize error with stack trace
	}
}

// Concurrency Benchmarks
// These benchmarks assess performance under concurrent error creation and context modification.

// BenchmarkConcurrency_Creation measures concurrent error creation and pooling.
func BenchmarkConcurrency_Creation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := New("parallel error") // Create errors concurrently
			err.Free()
		}
	})
}

// BenchmarkConcurrency_Context measures concurrent context addition to a shared error.
func BenchmarkConcurrency_Context(b *testing.B) {
	base := New("base")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = base.With("key", "value") // Add context concurrently
		}
	})
	base.Free()
}

// BenchmarkContext_Concurrent measures concurrent context addition with unique keys.
func BenchmarkContext_Concurrent(b *testing.B) {
	err := New("base")
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			err.With(fmt.Sprintf("key%d", i%10), i) // Add unique keys concurrently
			i++
		}
	})
}

// Pool and Allocation Benchmarks
// These benchmarks evaluate pooling mechanisms and raw allocation costs.

// BenchmarkPoolGetPut measures the speed of pool get and put operations.
func BenchmarkPoolGetPut(b *testing.B) {
	e := &Error{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errorPool.Put(e)    // Return error to pool
		e = errorPool.Get() // Retrieve error from pool
	}
}

// BenchmarkPoolWarmup measures the cost of resetting and warming the error pool.
func BenchmarkPoolWarmup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		errorPool = NewErrorPool() // Recreate pool
		WarmPool(100)              // Pre-warm with 100 errors
	}
}

// BenchmarkStackAlloc measures the cost of allocating a stack slice.
func BenchmarkStackAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]uintptr, 0, currentConfig.stackDepth) // Allocate stack slice
	}
}

// Special Case Benchmarks
// These benchmarks test specialized error creation methods.

// BenchmarkSpecial_Named measures creating a named error with a stack trace.
func BenchmarkSpecial_Named(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Named("test_error") // Create named error with stack
		err.Free()
	}
}

// BenchmarkSpecial_Format measures creating a formatted error.
func BenchmarkSpecial_Format(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Errorf("formatted %s %d", "error", i) // Create formatted error
		err.Free()
	}
}
