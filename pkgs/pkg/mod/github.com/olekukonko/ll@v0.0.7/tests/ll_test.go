package tests

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/olekukonko/ll"
	"github.com/olekukonko/ll/lh"
	"github.com/olekukonko/ll/lm"
	"github.com/olekukonko/ll/lx"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain sets up the test environment and runs the test suite.
// It resets the defaultLogger to a clean state to prevent state leakage between tests.
func TestMain(m *testing.M) {
	// Run tests and exit with the appropriate status code
	os.Exit(m.Run())
}

// TestLoggerConfiguration verifies the basic configuration methods of the Logger.
func TestLoggerConfiguration(t *testing.T) {
	// Create a new logger with namespace "test"
	logger := ll.New("test").Enable()

	// Test Enable/Disable functionality
	logger = logger.Disable()
	logger.Infof("Should not log") // Should be ignored since logger is disabled
	if logger.Enabled() {
		t.Errorf("Expected enabled=false, got %v", logger.Enabled())
	}
	logger = logger.Enable()
	if !logger.Enabled() {
		t.Errorf("Expected enabled=true, got %v", logger.Enabled())
	}

	// Test Level functionality
	logger = logger.Level(lx.LevelWarn)
	if logger.GetLevel() != lx.LevelWarn {
		t.Errorf("Expected level=%v, got %v", lx.LevelWarn, logger.GetLevel())
	}
	logger.Infof("Should not log") // Below Warn level, should be ignored
	logger.Warnf("Should log")     // At Warn level, should be processed

	// Test Style functionality
	logger = logger.Style(lx.NestedPath)
	if logger.GetStyle() != lx.NestedPath {
		t.Errorf("Expected style=%v, got %v", lx.NestedPath, logger.GetStyle())
	}
	logger = logger.Style(lx.FlatPath)
	if logger.GetStyle() != lx.FlatPath {
		t.Errorf("Expected style=%v, got %v", lx.FlatPath, logger.GetStyle())
	}
}

// TestLoggingMethods verifies the core logging methods (Debug, Info, Warn, Error, Stack).
func TestLoggingMethods(t *testing.T) {
	// Set up a logger with a buffer for capturing output
	buf := &bytes.Buffer{}
	logger := ll.New("test").Enable().Handler(lh.NewTextHandler(buf)).Level(lx.LevelDebug)

	// Test Debug logging
	buf.Reset()
	logger.Fields("key", "value").Debugf("Debug message")
	if !strings.Contains(buf.String(), "[test] DEBUG: Debug message [key=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] DEBUG: Debug message [key=value]")
	}

	// Test Info logging
	buf.Reset()
	logger.Fields("key", "value").Infof("Info message")
	if !strings.Contains(buf.String(), "[test] INFO: Info message [key=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Info message [key=value]")
	}

	// Test Warn logging
	buf.Reset()
	logger.Fields("key", "value").Warnf("Warn message")
	if !strings.Contains(buf.String(), "[test] WARN: Warn message [key=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] WARN: Warn message [key=value]")
	}

	// Test Error logging
	buf.Reset()
	logger.Fields("key", "value").Errorf("Error message")
	if !strings.Contains(buf.String(), "[test] ERROR: Error message [key=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] ERROR: Error message [key=value]")
	}

	// Test Stack logging with stack trace
	buf.Reset()
	logger.Fields("key", "value").Stackf("Error with stack")
	output := buf.String()
	if !strings.Contains(output, "[test] ERROR: Error with stack") {
		t.Errorf("Expected %q to contain %q", output, "[test] ERROR: Error with stack")
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected %q to contain %q", output, "key=value")
	}
	if !strings.Contains(output, "stack") {
		t.Errorf("Expected %q to contain %q", output, "stack")
	}
}

// TestBuilderFields verifies the Fields and Field methods for adding metadata to logs.
func TestBuilderFields(t *testing.T) {
	// Set up a logger with a buffer for capturing output
	buf := &bytes.Buffer{}
	logger := ll.New("test").Enable().Handler(lh.NewTextHandler(buf))

	// Test variadic Fields with multiple key-value pairs
	buf.Reset()
	logger.Fields("k1", "v1", "k2", "v2", "k3", 123).Infof("Test variadic")
	if !strings.Contains(buf.String(), "[test] INFO: Test variadic [k1=v1 k2=v2 k3=123]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test variadic [k1=v1 k2=v2 k3=123]")
	}

	// Test map-based Field with a pre-constructed map
	buf.Reset()
	fields := map[string]interface{}{"k1": "v1", "k2": "v2", "k3": 123}
	logger.Field(fields).Infof("Test map")
	if !strings.Contains(buf.String(), "[test] INFO: Test map [k1=v1 k2=v2 k3=123]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test map [k1=v1 k2=v2 k3=123]")
	}

	// Test variadic Fields with uneven key-value pairs
	buf.Reset()
	logger.Fields("k1", "v1", "k2").Infof("Test uneven")
	if !strings.Contains(buf.String(), "[test] INFO: Test uneven [error=uneven key-value pairs in Fields: [k2] k1=v1]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test uneven [error=uneven key-value pairs in Fields: [k2] k1=v1]")
	}

	// Test variadic Fields with a non-string key
	buf.Reset()
	logger.Fields("k1", "v1", 42, "v2").Infof("Test non-string")
	if !strings.Contains(buf.String(), "[test] INFO: Test non-string [error=non-string key in Fields: 42 k1=v1]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test non-string [error=non-string key in Fields: 42 k1=v1]")
	}
}

// TestRateLimiting verifies rate-limiting functionality for a log level.
func TestRateLimiting(t *testing.T) {
	// Set up a logger with a buffer for capturing output
	buf := &bytes.Buffer{}
	logger := ll.New("test").Enable().Handler(lh.NewTextHandler(buf))
	logger.Use(lm.NewRateLimiter(lx.LevelInfo, 2, time.Second))

	// Test logging within the rate limit (2 logs allowed)
	buf.Reset()
	logger.Infof("Log 1")
	logger.Infof("Log 2")
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected %d logs, got %d", 2, len(lines))
	}
	if !strings.Contains(buf.String(), "Log 1") {
		t.Errorf("Expected %q to contain %q", buf.String(), "Log 1")
	}
	if !strings.Contains(buf.String(), "Log 2") {
		t.Errorf("Expected %q to contain %q", buf.String(), "Log 2")
	}

	// Test exceeding the rate limit
	buf.Reset()
	logger.Infof("Log 3") // Should be blocked
	if buf.String() != "" {
		t.Errorf("Expected empty buffer, got %q", buf.String())
	}

	// Test logging after the interval resets
	time.Sleep(time.Second)
	buf.Reset()
	logger.Infof("Log 4")
	if !strings.Contains(buf.String(), "Log 4") {
		t.Errorf("Expected %q to contain %q", buf.String(), "Log 4")
	}
}

// TestSampling verifies sampling functionality for a log level.
func TestSampling(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := ll.New("test").Enable().Handler(lh.NewTextHandler(buf)).Clear() // Clear middleware
	logger.Use(lm.NewSampling(lx.LevelInfo, 0.0))                             // Never log

	// Test logging with 0.0 sampling rate
	buf.Reset()
	logger.Infof("Should not log")
	if buf.String() != "" {
		t.Errorf("Expected empty buffer, got %q", buf.String())
	}

	// Test logging with 1.0 sampling rate
	logger = ll.New("test").Enable().Handler(lh.NewTextHandler(buf)).Clear() // Fresh logger
	logger.Use(lm.NewSampling(lx.LevelInfo, 1.0))                            // Always log
	buf.Reset()
	logger.Infof("Should log")
	if !strings.Contains(buf.String(), "[test] INFO: Should log") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Should log")
	}
}

// TestConditionalLogging verifies conditional logging using the If method.
func TestConditionalLogging(t *testing.T) {
	// Set up a logger with a buffer for capturing output
	buf := &bytes.Buffer{}
	logger := ll.New("VC").Enable().Handler(lh.NewTextHandler(buf)).Level(lx.LevelDebug)

	// Test false condition with variadic Fields
	buf.Reset()
	logger.If(false).Fields("key", "value").Infof("Should not log")
	if buf.String() != "" {
		t.Errorf("Expected empty buffer, got %q", buf.String())
	}

	// Test true condition with variadic Fields
	buf.Reset()
	logger.If(true).Fields("key", "value").Infof("Should log")
	if !strings.Contains(buf.String(), "[VC] INFO: Should log [key=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[VC] INFO: Should log [key=value]")
	}

	// Test false condition with map-based Field
	buf.Reset()
	fields := map[string]interface{}{"key": "value"}
	logger.If(false).Field(fields).Infof("Should not log")
	if buf.String() != "" {
		t.Errorf("Expected empty buffer, got %q", buf.String())
	}

	// Test true condition with map-based Field
	buf.Reset()
	logger.If(true).Field(fields).Infof("Should log")
	if !strings.Contains(buf.String(), "[VC] INFO: Should log [key=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[VC] INFO: Should log [key=value]")
	}

	// Test variadic Fields with uneven pairs under true condition
	buf.Reset()
	logger.If(true).Fields("key", "value", "odd").Infof("Test uneven")
	if !strings.Contains(buf.String(), "[VC] INFO: Test uneven [error=uneven key-value pairs in Fields: [odd] key=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[VC] INFO: Test uneven [error=uneven key-value pairs in Fields: [odd] key=value]")
	}

	// Test variadic Fields with non-string key under true condition
	buf.Reset()
	logger.If(true).Fields("key", "value", 42, "value2").Infof("Test non-string")
	if !strings.Contains(buf.String(), "[VC] INFO: Test non-string [error=non-string key in Fields: 42 key=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[VC] INFO: Test non-string [error=non-string key in Fields: 42 key=value]")
	}

	// Test Conditional Stack logging with stack trace
	t.Run("ConditionalStack", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := ll.New("test/app").Enable().Style(lx.NestedPath).Handler(lh.NewTextHandler(buf)).Level(lx.LevelDebug).Prefix("ERR: ").Indent(1)
		logger = logger.Context(map[string]interface{}{"ctx": "value"})
		logger.If(true).Stackf("error occurred: %v", "timeout")
		expectedStack := "[stack]"
		if !strings.Contains(buf.String(), expectedStack) {
			t.Errorf("Expected %q to contain %q; \ngot %q", buf.String(), expectedStack, buf.String())
		}
		buf.Reset()
		logger.If(false).Stackf("should not log: %v", "timeout")
		if buf.String() != "" {
			t.Errorf("Expected empty buffer, got %q", buf.String())
		}
	})
}

// TestMiddleware verifies the Use method for adding middleware to process log entries.
func TestMiddleware(t *testing.T) {
	// Set up a logger with a buffer for capturing output
	buf := &bytes.Buffer{}
	logger := ll.New("test").Enable().Handler(lh.NewTextHandler(buf)).Level(lx.LevelDebug)

	// Test middleware that adds a field
	logger = logger.Use(ll.Middle(func(e *lx.Entry) error {
		if e.Fields == nil {
			e.Fields = make(map[string]interface{})
		}
		e.Fields["extra"] = "value"
		return nil
	})).Logger()

	buf.Reset()
	logger.Infof("Test with extra field")
	if !strings.Contains(buf.String(), "[test] INFO: Test with extra field [extra=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test with extra field [extra=value]")
	}

	// Test middleware that filters logs by level
	logger = logger.Use(ll.Middle(func(e *lx.Entry) error {
		if e.Level >= lx.LevelWarn {
			return nil
		}
		return fmt.Errorf("level too low")
	})).Logger()
	buf.Reset()
	logger.Infof("Should not log") // Below Warn level, should be ignored
	if buf.String() != "" {
		t.Errorf("Expected empty buffer, got %q", buf.String())
	}
	buf.Reset()
	logger.Warnf("Should log")
	if !strings.Contains(buf.String(), "[test] WARN: Should log [extra=value]") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] WARN: Should log [extra=value]")
	}

	// Test middleware that skips all logs
	logger = logger.Use(ll.Middle(func(e *lx.Entry) error {
		return fmt.Errorf("skip all")
	})).Logger()
	buf.Reset()
	logger.Warnf("Should not log") // Should be ignored by middleware
	if buf.String() != "" {
		t.Errorf("Expected empty buffer, got %q", buf.String())
	}
}

// TestClone verifies the Clone method for creating a logger with the same namespace and isolated context.
func TestClone(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := ll.New("app").Enable().Handler(lh.NewTextHandler(buf)).Level(lx.LevelInfo).Style(lx.NestedPath)
	t.Logf("Initial logger enabled: %v, level: %v, style: %v", logger.Enabled(), logger.GetLevel(), logger.GetStyle())

	t.Run("Namespace", func(t *testing.T) {
		clone := logger.Clone()
		if clone.GetPath() != "app" {
			t.Errorf("Expected clone namespace %q, got %q", "app", clone.GetPath())
		}
	})

	t.Run("Configuration", func(t *testing.T) {
		clone := logger.Clone()
		if !clone.Enabled() {
			t.Errorf("Expected clone enabled=true, got %v", clone.Enabled())
		}
		if clone.GetLevel() != lx.LevelInfo {
			t.Errorf("Expected clone level=%v, got %v", lx.LevelInfo, clone.GetLevel())
		}
		if clone.GetStyle() != lx.NestedPath {
			t.Errorf("Expected clone style=%v, got %v", lx.NestedPath, clone.GetStyle())
		}
	})

	t.Run("ContextIsolation", func(t *testing.T) {
		logger = logger.Context(map[string]interface{}{"parent": "value"})
		t.Logf("Parent context: %v", logger.GetContext()) // Use GetContext
		clone := logger.Clone()
		buf.Reset()
		clone.Fields("clone", "value").Infof("Clone message")
		expected := "[app]" + lx.Colon + lx.Space + "INFO: Clone message [clone=value]"
		output := buf.String()
		t.Logf("Clone output: %q", output)
		if !strings.Contains(output, expected) {
			t.Errorf("Expected %q to contain %q; got %q", output, expected, output)
		}
		if strings.Contains(output, "parent=value") {
			t.Errorf("Expected %q not to contain %q", output, "parent=value")
		}
	})

	t.Run("ParentContext", func(t *testing.T) {
		t.Logf("Parent context before logging: %v", logger.GetContext()) // Use GetContext
		t.Logf("Parent enabled: %v, system active: %v", logger.Enabled(), ll.Active())
		buf.Reset()
		logger.Infof("Parent message")
		output := buf.String()
		expected := "[app]" + lx.Colon + lx.Space + "INFO: Parent message [parent=value]"
		t.Logf("Parent output: %q", output)
		if !strings.Contains(output, expected) {
			t.Errorf("Expected %q to contain %q; got %q", output, expected, output)
		}
		if strings.Contains(output, "clone=value") {
			t.Errorf("Expected %q not to contain %q", output, "clone=value")
		}
	})
}

// TestPrefix verifies the Prefix method for prepending a string to log messages.
func TestPrefix(t *testing.T) {
	// Set up a logger with a buffer for capturing output
	buf := &bytes.Buffer{}
	logger := ll.New("app").Enable().Handler(lh.NewTextHandler(buf)).Level(lx.LevelInfo)

	// Test setting a prefix
	t.Run("SetPrefix", func(t *testing.T) {
		buf.Reset()
		logger = logger.Prefix("INFO: ")
		logger.Infof("Test message")
		output := buf.String()
		//t.Logf("Buffer output: %q", output)
		if !strings.Contains(output, "INFO: Test message") {
			t.Errorf("Expected %q to contain %q", output, "INFO: Test message")
		}
	})

	// Test updating the prefix
	t.Run("UpdatePrefix", func(t *testing.T) {
		buf.Reset()
		logger = logger.Prefix("DEBUG: ")
		logger.Infof("Another message")
		output := buf.String()
		if !strings.Contains(output, "DEBUG: Another message") {
			t.Errorf("Expected %q to contain %q", output, "DEBUG: Another message")
		}
	})

	// Test removing the prefix
	t.Run("RemovePrefix", func(t *testing.T) {
		buf.Reset()
		logger = logger.Prefix("")
		logger.Infof("No prefix")
		output := buf.String()
		if !strings.Contains(output, "INFO: No prefix") {
			t.Errorf("Expected message without prefix, got: %q", output)
		}
	})
}

// TestIndent verifies the Indent method for adding double spaces to log messages.
func TestIndent(t *testing.T) {
	// Set up a logger with a buffer for capturing output
	buf := &bytes.Buffer{}
	logger := ll.New("app").Enable().Handler(lh.NewTextHandler(buf)).Level(lx.LevelInfo)

	// Test setting indentation to 2 (4 spaces)
	t.Run("SetIndent", func(t *testing.T) {
		logger = logger.Indent(2)
		buf.Reset()
		logger.Infof("Test message")
		output := buf.String()
		t.Logf("Buffer output: %q", output) // Debug output
		if !strings.Contains(output, "[app] INFO:     Test message") {
			t.Errorf("Expected %q to contain %q", output, "[app] INFO:     Test message")
		}
	})

	// Test updating indentation to 1 (2 spaces)
	t.Run("UpdateIndent", func(t *testing.T) {
		logger = logger.Indent(1)
		buf.Reset()
		logger.Infof("Another message")
		if !strings.Contains(buf.String(), "[app] INFO:   Another message") {
			t.Errorf("Expected %q to contain %q", buf.String(), "[app] INFO:   Another message")
		}
	})

	// Test removing indentation
	t.Run("RemoveIndent", func(t *testing.T) {
		logger = logger.Indent(0)
		buf.Reset()
		logger.Infof("No indent")
		if !strings.Contains(buf.String(), "[app] INFO: No indent") {
			t.Errorf("Expected %q to contain %q", buf.String(), "[app] INFO: No indent")
		}
	})
}

// failingWriter is a test writer that always fails to write, used to simulate handler errors.
type failingWriter struct{}

// Write implements io.Writer, always returning an error.
func (w *failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

// TestHandlerErrors verifies handler behavior when errors occur.
func TestHandlerErrors(t *testing.T) {
	// Test single TextHandler
	buf := &bytes.Buffer{}
	logger := ll.New("test").Enable().Level(lx.LevelDebug).Handler(lh.NewTextHandler(buf))

	logger.Infof("Test single handler")
	if !strings.Contains(buf.String(), "[test] INFO: Test single handler") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test single handler")
	}

	// Test MultiHandler with a failing TextHandler
	buf.Reset()
	logger = logger.Handler(lh.NewMultiHandler(
		lh.NewTextHandler(buf),
		lh.NewTextHandler(&failingWriter{}),
	))
	logger.Infof("Test multi error")
	if !strings.Contains(buf.String(), "[test] INFO: Test multi error") {
		t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test multi error")
	}
}

// TestNamespaceToggle verifies the NamespaceEnable and NamespaceDisable methods.
func TestNamespaceToggle(t *testing.T) {
	// Create a logger and test namespace toggling
	logger := ll.New("test").Disable()
	logger = logger.NamespaceEnable("parent/child")
	if !logger.NamespaceEnabled("parent/child") {
		t.Errorf("parent/child should be enabled")
	}
	logger = logger.NamespaceDisable("parent/child")
	if logger.Enabled() {
		t.Errorf("parent/child should be disabled")
	}
}

// TestTextHandler verifies the TextHandler’s output format.
func TestTextHandler(t *testing.T) {
	// Create a buffer and TextHandler
	var buf bytes.Buffer
	h := lh.NewTextHandler(&buf)
	// Create a test log entry
	e := &lx.Entry{
		Timestamp: time.Now(),
		Level:     lx.LevelInfo,
		Message:   "test",
		Namespace: "",
		Fields:    map[string]interface{}{"key": 1},
	}
	// Process the entry
	if err := h.Handle(e); err != nil {
		t.Errorf("Handle failed: %v", err)
	}
	// Verify the output format
	if !strings.Contains(buf.String(), "INFO: test [key=1]") {
		t.Errorf("Unexpected output: %s", buf.String())
	}
}
