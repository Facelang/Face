package tests

import (
	"fmt"
	"github.com/olekukonko/ll"
	"github.com/olekukonko/ll/lh"
	"io"
	"testing"
)

// BenchmarkNamespaceLoop measures the performance of calling Namespace in a loop.
// Each iteration creates a new child logger with a unique namespace and logs a message.
func BenchmarkNamespaceLoop(b *testing.B) {
	logger := ll.New("app")
	logger.Handler(lh.NewTextHandler(io.Discard))
	logger.Enable()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		child := logger.Namespace(fmt.Sprintf("child%d", i))
		child.Infof("Message")
	}
}

// BenchmarkNamespaceCached measures the performance of caching Namespace calls outside a loop.
// Loggers are pre-created, and only logging is performed in the loop.
func BenchmarkNamespaceCached(b *testing.B) {
	logger := ll.New("app")
	logger.Handler(lh.NewTextHandler(io.Discard))
	logger.Enable()
	childLoggers := make([]*ll.Logger, b.N)
	for i := 0; i < b.N; i++ {
		childLoggers[i] = logger.Namespace(fmt.Sprintf("child%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		childLoggers[i].Infof("Message")
	}
}

// BenchmarkCloneLoop measures the performance of calling Clone in a loop.
// Each iteration creates a new logger with the same namespace and logs a message.
func BenchmarkCloneLoop(b *testing.B) {
	logger := ll.New("app")
	logger.Handler(lh.NewTextHandler(io.Discard))
	logger.Enable()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clone := logger.Clone()
		clone.Infof("Message")
	}
}

// BenchmarkFieldsLoggerLoop measures the performance of calling Fields followed by Logger in a loop.
// Each iteration creates a new logger with embedded fields and logs a message.
func BenchmarkFieldsLoggerLoop(b *testing.B) {
	logger := ll.New("app")
	logger.Handler(lh.NewTextHandler(io.Discard))
	logger.Enable()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newLogger := logger.Fields("iteration", i).Logger()
		newLogger.Infof("Message")
	}
}

// BenchmarkPrefixLoop measures the performance of setting a prefix in a loop.
// Each iteration modifies the logger’s prefix and logs a message.
func BenchmarkPrefixLoop(b *testing.B) {
	logger := ll.New("app")
	logger.Handler(lh.NewTextHandler(io.Discard))
	logger.Enable()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Prefix(fmt.Sprintf("prefix%d: ", i))
		logger.Infof("Message")
	}
}

// BenchmarkIndentLoop measures the performance of setting indentation in a loop.
// Each iteration modifies the logger’s indentation level and logs a message.
func BenchmarkIndentLoop(b *testing.B) {
	logger := ll.New("app")
	logger.Handler(lh.NewTextHandler(io.Discard))
	logger.Enable()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Indent(i % 5) // Vary indentation up to 5 levels
		logger.Infof("Message")
	}
}
