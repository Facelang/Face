package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/olekukonko/ll/lh"
	"github.com/olekukonko/ll/lx"
)

// errorWriter is an io.Writer that always returns an error.
// It simulates a failing writer for testing error handling in log handlers.
// Used in tests like ErrorLogging to verify error logging behavior.
type errorWriter struct {
	err error // Error to return on Write
}

// Write implements io.Writer, always returning the configured error.
// It returns 0 bytes written and the specified error, simulating a write failure.
// Example (internal usage):
//
//	w := &errorWriter{err: errors.New("write error")}
//	n, err := w.Write([]byte("test")) // Returns n=0, err="write error"
func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

// TestBufferedHandler tests the functionality of the Buffered handler.
// It includes subtests for basic buffering, periodic flushing, overflow handling,
// explicit flushing, shutdown behavior, concurrent access, error handling, and finalizer cleanup.
// Each subtest verifies specific aspects of the Buffered handler's behavior under various conditions.
func TestBufferedHandler(t *testing.T) {
	// Test basic buffering: flushing when batch size is reached
	t.Run("BasicFunctionality", func(t *testing.T) {
		buf := &bytes.Buffer{}                 // Buffer to capture output
		textHandler := lh.NewTextHandler(buf)  // Create text handler
		handler := lh.NewBuffered(textHandler, // Create buffered handler
			lh.WithBatchSize(2),                        // Flush every 2 entries
			lh.WithFlushInterval(100*time.Millisecond)) // Flush every 100ms
		defer handler.Close() // Ensure cleanup

		// Send two log entries
		handler.Handle(&lx.Entry{Message: "test1"})
		handler.Handle(&lx.Entry{Message: "test2"})

		// Wait for batch flush
		time.Sleep(150 * time.Millisecond) // Allow time for flush to occur
		output := buf.String()             // Get captured output
		// Verify both messages are present
		if !strings.Contains(output, "test1") || !strings.Contains(output, "test2") {
			t.Errorf("Expected both messages in output, got: %q", output)
		}
	})

	// Test periodic flushing: flushing after a time interval
	t.Run("PeriodicFlushing", func(t *testing.T) {
		buf := &bytes.Buffer{}
		textHandler := lh.NewTextHandler(buf)
		handler := lh.NewBuffered(textHandler,
			lh.WithBatchSize(100),                     // High batch size to avoid batch flush
			lh.WithFlushInterval(50*time.Millisecond)) // Flush every 50ms
		defer handler.Close()

		// Send one log entry
		handler.Handle(&lx.Entry{Message: "test"})

		// Should flush after interval even though batch size not reached
		time.Sleep(75 * time.Millisecond) // Wait longer than flush interval
		// Verify message is flushed
		if !strings.Contains(buf.String(), "test") {
			t.Error("Expected message to be flushed after interval")
		}
	})

	// Test overflow handling: behavior when buffer exceeds MaxBuffer
	t.Run("OverflowHandling", func(t *testing.T) {
		buf := &bytes.Buffer{}
		textHandler := lh.NewTextHandler(buf)
		var overflowCalled bool // Track overflow handler calls
		handler := lh.NewBuffered(textHandler,
			lh.WithBatchSize(2), // Flush every 2 entries
			lh.WithMaxBuffer(2), // Max buffer size of 2
			lh.WithOverflowHandler(func(int) { overflowCalled = true }), // Set overflow callback
		)
		defer handler.Close()

		// Fill buffer
		handler.Handle(&lx.Entry{Message: "test1"})
		handler.Handle(&lx.Entry{Message: "test2"})

		// This should trigger overflow
		err := handler.Handle(&lx.Entry{Message: "test3"})
		// Verify overflow error
		if err == nil {
			t.Error("Expected error on overflow")
		}
		// Verify overflow handler was called
		if !overflowCalled {
			t.Error("Expected overflow handler to be called")
		}
	})

	// Test explicit flush: flushing on demand with Flush()
	t.Run("ExplicitFlush", func(t *testing.T) {
		buf := &bytes.Buffer{}
		textHandler := lh.NewTextHandler(buf)
		handler := lh.NewBuffered(textHandler, lh.WithBatchSize(100)) // High batch size
		defer handler.Close()

		// Send one log entry
		handler.Handle(&lx.Entry{Message: "test"})
		handler.Flush()                   // Trigger immediate flush
		time.Sleep(10 * time.Millisecond) // Allow worker to process
		// Verify message is flushed
		if !strings.Contains(buf.String(), "test") {
			t.Error("Expected message to be flushed after explicit flush")
		}
	})

	// Test shutdown: flushing remaining entries on Close()
	t.Run("ShutdownDrainsBuffer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		textHandler := lh.NewTextHandler(buf)
		handler := lh.NewBuffered(textHandler, lh.WithBatchSize(100)) // High batch size
		defer handler.Close()

		// Send one log entry
		handler.Handle(&lx.Entry{Message: "test"})
		handler.Close() // Close should flush remaining entries

		// Verify message is flushed
		if !strings.Contains(buf.String(), "test") {
			t.Error("Expected message to be flushed on shutdown")
		}
	})

	// Test concurrent access: handling multiple goroutines writing logs
	t.Run("ConcurrentAccess", func(t *testing.T) {
		buf := &bytes.Buffer{}
		textHandler := lh.NewTextHandler(buf)
		handler := lh.NewBuffered(textHandler,
			lh.WithBatchSize(100),                     // High batch size
			lh.WithFlushInterval(10*time.Millisecond), // Frequent flushes
			lh.WithMaxBuffer(1000))                    // Large buffer for concurrent writes
		defer handler.Close()

		var wg sync.WaitGroup // Synchronize goroutines
		// Send 100 log entries concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				handler.Handle(&lx.Entry{Message: fmt.Sprintf("test%d", i)})
			}(i)
		}
		wg.Wait()                         // Wait for all goroutines to finish
		handler.Flush()                   // Trigger final flush
		time.Sleep(50 * time.Millisecond) // Allow worker to process
		output := buf.String()
		t.Logf("Buffer output length: %d", len(output)) // Debug output
		// Verify all messages are present
		for i := 0; i < 100; i++ {
			if !strings.Contains(output, fmt.Sprintf("test%d", i)) {
				t.Errorf("Missing message test%d in output", i)
			}
		}
	})

	// Test error handling: behavior with a failing writer
	t.Run("ErrorHandling", func(t *testing.T) {
		errWriter := &errorWriter{err: errors.New("write error")} // Failing writer
		textHandler := lh.NewTextHandler(errWriter)
		handler := lh.NewBuffered(textHandler, lh.WithBatchSize(1)) // Flush every entry
		defer handler.Close()

		// Send one log entry
		err := handler.Handle(&lx.Entry{Message: "test"})
		// Verify no error on Handle (buffered, not yet written)
		if err != nil {
			t.Errorf("Unexpected error on Handle: %v", err)
		}

		// Wait for flush to occur
		time.Sleep(50 * time.Millisecond) // Allow worker to attempt write
	})

	// Test finalizer: flushing entries during garbage collection
	t.Run("Finalizer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		textHandler := lh.NewTextHandler(buf)
		handler := lh.NewBuffered(textHandler, lh.WithBatchSize(100)) // High batch size
		// Send one log entry
		handler.Handle(&lx.Entry{Message: "test"})

		// Simulate GC (in real code this would happen automatically)
		runtime.SetFinalizer(handler, nil) // Remove the finalizer for test
		handler.Final()                    // Trigger finalizer behavior

		// Verify message is flushed
		if !strings.Contains(buf.String(), "test") {
			t.Error("Expected message to be flushed by finalizer")
		}
	})
}

// TestBufferedHandlerOptions tests the configuration options of the Buffered handler.
// It verifies default values and custom settings for BatchSize, FlushInterval, MaxBuffer,
// and the overflow handler, ensuring proper initialization and behavior.
func TestBufferedHandlerOptions(t *testing.T) {
	// Test default configuration values
	t.Run("DefaultValues", func(t *testing.T) {
		textHandler := lh.NewTextHandler(&bytes.Buffer{})
		handler := lh.NewBuffered(textHandler) // Create with defaults
		// Verify default BatchSize
		if handler.Config().BatchSize != 100 {
			t.Errorf("Expected default BatchSize=100, got %d", handler.Config().BatchSize)
		}
		// Verify default FlushInterval
		if handler.Config().FlushInterval != 10*time.Second {
			t.Errorf("Expected default FlushInterval=10s, got %v", handler.Config().FlushInterval)
		}
		// Verify default MaxBuffer
		if handler.Config().MaxBuffer != 1000 {
			t.Errorf("Expected default MaxBuffer=1000, got %d", handler.Config().MaxBuffer)
		}
	})

	// Test custom configuration options
	t.Run("CustomOptions", func(t *testing.T) {
		textHandler := lh.NewTextHandler(&bytes.Buffer{})
		var called bool // Track overflow handler calls
		handler := lh.NewBuffered(textHandler,
			lh.WithBatchSize(50),                                // Custom batch size
			lh.WithFlushInterval(5*time.Second),                 // Custom flush interval
			lh.WithMaxBuffer(500),                               // Custom max buffer
			lh.WithOverflowHandler(func(int) { called = true }), // Custom overflow handler
		)

		// Verify custom BatchSize
		if handler.Config().BatchSize != 50 {
			t.Errorf("Expected BatchSize=50, got %d", handler.Config().BatchSize)
		}
		// Verify custom FlushInterval
		if handler.Config().FlushInterval != 5*time.Second {
			t.Errorf("Expected FlushInterval=5s, got %v", handler.Config().FlushInterval)
		}
		// Verify custom MaxBuffer
		if handler.Config().MaxBuffer != 500 {
			t.Errorf("Expected MaxBuffer=500, got %d", handler.Config().MaxBuffer)
		}

		// Test overflow handler
		handler.Config().OnOverflow(1) // Trigger overflow callback
		// Verify overflow handler was called
		if !called {
			t.Error("Expected overflow handler to be called")
		}
	})
}

// TestBufferedHandlerEdgeCases tests edge cases for Buffered handler configuration.
// It verifies that invalid or extreme configuration values (e.g., zero batch size,
// negative flush interval, small max buffer) are adjusted to sane defaults.
func TestBufferedHandlerEdgeCases(t *testing.T) {
	// Test zero batch size adjustment
	t.Run("ZeroBatchSize", func(t *testing.T) {
		textHandler := lh.NewTextHandler(&bytes.Buffer{})
		handler := lh.NewBuffered(textHandler, lh.WithBatchSize(0)) // Invalid batch size
		// Verify adjustment to minimum
		if handler.Config().BatchSize != 1 {
			t.Errorf("Expected BatchSize to be adjusted to 1, got %d", handler.Config().BatchSize)
		}
	})

	// Test negative flush interval adjustment
	t.Run("NegativeFlushInterval", func(t *testing.T) {
		textHandler := lh.NewTextHandler(&bytes.Buffer{})
		handler := lh.NewBuffered(textHandler, lh.WithFlushInterval(-1*time.Second)) // Invalid interval
		// Verify adjustment to default
		if handler.Config().FlushInterval != 10*time.Second {
			t.Errorf("Expected FlushInterval to be adjusted to 10s, got %v", handler.Config().FlushInterval)
		}
	})

	// Test small max buffer adjustment
	t.Run("SmallMaxBuffer", func(t *testing.T) {
		textHandler := lh.NewTextHandler(&bytes.Buffer{})
		handler := lh.NewBuffered(textHandler,
			lh.WithBatchSize(10), // Valid batch size
			lh.WithMaxBuffer(5),  // Invalid max buffer (less than batch size)
		)
		// Verify adjustment to ensure MaxBuffer >= BatchSize
		if handler.Config().MaxBuffer < handler.Config().BatchSize {
			t.Errorf("Expected MaxBuffer >= BatchSize, got %d < %d",
				handler.Config().MaxBuffer, handler.Config().BatchSize)
		}
	})
}

// TestBufferedHandlerIntegration tests the Buffered handler's integration with various handlers.
// It verifies that the Buffered handler works correctly with TextHandler, JSONHandler,
// MultiHandler, and error-prone writers, ensuring proper flushing and error logging.
func TestBufferedHandlerIntegration(t *testing.T) {
	// Test integration with TextHandler
	t.Run("WithTextHandler", func(t *testing.T) {
		buf := &bytes.Buffer{}
		textHandler := lh.NewTextHandler(buf)
		handler := lh.NewBuffered(textHandler,
			lh.WithBatchSize(2),                       // Flush every 2 entries
			lh.WithFlushInterval(50*time.Millisecond)) // Flush every 50ms
		defer handler.Close()

		// Send two log entries
		handler.Handle(&lx.Entry{Message: "message1"})
		handler.Handle(&lx.Entry{Message: "message2"})

		// Wait for flush
		time.Sleep(75 * time.Millisecond) // Allow time for flush
		output := buf.String()
		// Verify both messages are present
		if !strings.Contains(output, "message1") || !strings.Contains(output, "message2") {
			t.Errorf("Expected both messages in output, got: %q", output)
		}
	})

	// Test integration with JSONHandler
	t.Run("WithJSONHandler", func(t *testing.T) {
		buf := &bytes.Buffer{}
		jsonHandler := lh.NewJSONHandler(buf)
		handler := lh.NewBuffered(jsonHandler, lh.WithBatchSize(2)) // Flush every 2 entries
		defer handler.Close()

		// Send two log entries
		handler.Handle(&lx.Entry{Message: "message1"})
		handler.Handle(&lx.Entry{Message: "message2"})
		handler.Flush()                   // Trigger flush
		time.Sleep(50 * time.Millisecond) // Allow worker to process
		var count int                     // Count JSON entries
		// Decode JSON output
		dec := json.NewDecoder(buf)
		for {
			var entry map[string]interface{}
			if err := dec.Decode(&entry); err == io.EOF {
				break
			} else if err != nil {
				t.Fatal(err)
			}
			count++
		}
		t.Logf("JSON entry count: %d", count) // Debug output
		// Verify two JSON entries
		if count != 2 {
			t.Errorf("Expected 2 JSON entries, got %d", count)
		}
	})

	// Test integration with MultiHandler
	t.Run("WithMultiHandler", func(t *testing.T) {
		buf1 := &bytes.Buffer{} // Buffer for text output
		buf2 := &bytes.Buffer{} // Buffer for JSON output
		multiHandler := lh.NewMultiHandler(
			lh.NewTextHandler(buf1), // Text handler
			lh.NewJSONHandler(buf2), // JSON handler
		)
		handler := lh.NewBuffered(multiHandler, lh.WithBatchSize(3)) // Flush every 3 entries
		defer handler.Close()

		// Send three log entries
		handler.Handle(&lx.Entry{Message: "test"})
		handler.Handle(&lx.Entry{Message: "test"})
		handler.Handle(&lx.Entry{Message: "test"})
		handler.Flush()                   // Trigger flush
		time.Sleep(10 * time.Millisecond) // Allow worker to process

		// Verify text output
		textOutput := buf1.String()
		t.Logf("Text output: %q", textOutput) // Debug output
		if strings.Count(textOutput, "test") != 3 {
			t.Error("Expected 3 messages in text output")
		}

		// Verify JSON output
		var count int
		dec := json.NewDecoder(buf2)
		for dec.More() {
			var entry map[string]interface{}
			if err := dec.Decode(&entry); err != nil {
				t.Fatal(err)
			}
			count++
		}
		t.Logf("JSON entry count: %d", count) // Debug output
		// Verify three JSON entries
		if count != 3 {
			t.Errorf("Expected 3 JSON entries, got %d", count)
		}
	})

	// Test error logging with a failing writer
	t.Run("ErrorLogging", func(t *testing.T) {
		errWriter := &errorWriter{err: errors.New("write error")} // Failing writer
		textHandler := lh.NewTextHandler(errWriter)

		// Set up stderr capture before creating the handler
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		handler := lh.NewBuffered(textHandler)

		var errOutput bytes.Buffer
		errChan := make(chan struct{})
		// Capture stderr in a goroutine
		go func() {
			defer close(errChan)
			io.Copy(&errOutput, r)
		}()

		// Send one log entry
		handler.Handle(&lx.Entry{Message: "message"})
		handler.Flush()

		// Give time for the flush to occur
		time.Sleep(50 * time.Millisecond)

		// Clean up
		handler.Close()
		w.Close()
		os.Stderr = oldStderr
		<-errChan // Wait for stderr capture to complete

		// Verify error is logged to stderr
		if !strings.Contains(errOutput.String(), "write error") {
			t.Errorf("Expected error to be logged to stderr, got: %q", errOutput.String())
		}
	})
}
