package errors

import (
	"database/sql"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var testMu sync.Mutex // Protect global state changes

// TestHelperWarmStackPool verifies that WarmStackPool pre-populates the stack pool correctly.
func TestHelperWarmStackPool(t *testing.T) {
	testMu.Lock()
	defer testMu.Unlock()

	// Save and restore original config
	originalConfig := currentConfig
	defer func() { currentConfig = originalConfig }()

	// Reinitialize stackPool with a nil-returning New function for this test
	stackPool = sync.Pool{
		New: func() interface{} {
			return nil // Return nil when pool is empty
		},
	}

	// Test disabled pooling
	currentConfig.disablePooling = true
	WarmStackPool(5)
	if got := stackPool.Get(); got != nil {
		t.Errorf("WarmStackPool should not populate when pooling is disabled, got %v", got)
	}

	// Reinitialize stackPool for enabled pooling test
	stackPool = sync.Pool{
		New: func() interface{} {
			return make([]uintptr, currentConfig.stackDepth)
		},
	}

	// Test enabled pooling
	currentConfig.disablePooling = false
	WarmStackPool(3)
	count := 0
	for i := 0; i < 3; i++ {
		if stackPool.Get() != nil {
			count++
		}
	}
	if count != 3 {
		t.Errorf("WarmStackPool should populate 3 items, got %d", count)
	}
}

// TestHelperCaptureStack verifies that captureStack captures the correct stack frames.
func TestHelperCaptureStack(t *testing.T) {
	stack := captureStack(0)
	if len(stack) == 0 {
		t.Error("captureStack should capture at least one frame")
	}
	found := false
	frames := runtime.CallersFrames(stack)
	for {
		frame, more := frames.Next()
		if frame == (runtime.Frame{}) {
			break
		}
		if strings.Contains(frame.Function, "TestHelperCaptureStack") {
			found = true
			break
		}
		if !more {
			break
		}
	}
	if !found {
		t.Error("captureStack should include TestHelperCaptureStack in the stack")
	}
}

// TestHelperMin verifies the min helper function returns the smaller integer.
func TestHelperMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}
	for _, tt := range tests {
		if got := min(tt.a, tt.b); got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// TestHelperClearMap verifies that clearMap empties a map.
func TestHelperClearMap(t *testing.T) {
	m := map[string]interface{}{
		"a": 1,
		"b": "test",
	}
	clearMap(m)
	if len(m) != 0 {
		t.Errorf("clearMap should empty the map, got %d items", len(m))
	}
}

// TestHelperSqlNull verifies sqlNull detects SQL null types correctly.
func TestHelperSqlNull(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil", nil, true},
		{"null string", sql.NullString{Valid: false}, true},
		{"valid string", sql.NullString{String: "test", Valid: true}, false},
		{"null time", sql.NullTime{Valid: false}, true},
		{"valid time", sql.NullTime{Time: time.Now(), Valid: true}, false},
		{"non-sql type", "test", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqlNull(tt.value); got != tt.expected {
				t.Errorf("sqlNull(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

// TestHelperGetFuncName verifies getFuncName extracts function names correctly.
func TestHelperGetFuncName(t *testing.T) {
	if got := getFuncName(nil); got != "unknown" {
		t.Errorf("getFuncName(nil) = %q, want 'unknown'", got)
	}
	if got := getFuncName(TestHelperGetFuncName); !strings.Contains(got, "TestHelperGetFuncName") {
		t.Errorf("getFuncName(TestHelperGetFuncName) = %q, want to contain 'TestHelperGetFuncName'", got)
	}
}

// TestHelperIsInternalFrame verifies isInternalFrame identifies internal frames.
func TestHelperIsInternalFrame(t *testing.T) {
	tests := []struct {
		frame    runtime.Frame
		expected bool
	}{
		{runtime.Frame{Function: "runtime.main"}, true},
		{runtime.Frame{Function: "reflect.ValueOf"}, true},
		{runtime.Frame{File: "github.com/olekukonko/errors/errors.go"}, true},
		{runtime.Frame{Function: "main.main"}, false},
	}
	for _, tt := range tests {
		if got := isInternalFrame(tt.frame); got != tt.expected {
			t.Errorf("isInternalFrame(%v) = %v, want %v", tt.frame, got, tt.expected)
		}
	}
}

// TestHelperFormatError verifies FormatError produces the expected string output.
func TestHelperFormatError(t *testing.T) {
	err := New("test").With("key", "value").Wrap(New("cause"))
	defer err.Free()
	formatted := FormatError(err)
	if !strings.Contains(formatted, "Error: test: cause") {
		t.Errorf("FormatError missing error message: %q", formatted)
	}
	if !strings.Contains(formatted, "Context:\n\tkey: value") {
		t.Errorf("FormatError missing context: %q", formatted)
	}
	if !strings.Contains(formatted, "Caused by:") {
		t.Errorf("FormatError missing cause: %q", formatted)
	}

	if FormatError(nil) != "<nil>" {
		t.Error("FormatError(nil) should return '<nil>'")
	}

	stdErr := errors.New("std error")
	if !strings.Contains(FormatError(stdErr), "Error: std error") {
		t.Errorf("FormatError for std error missing message: %q", FormatError(stdErr))
	}
}

// TestHelperCaller verifies Caller returns the correct caller information.
func TestHelperCaller(t *testing.T) {
	file, line, function := Caller(0)
	if !strings.Contains(file, "helper_test.go") {
		t.Errorf("Caller file = %q, want to contain 'helper_test.go'", file)
	}
	if line <= 0 {
		t.Errorf("Caller line = %d, want > 0", line)
	}
	if !strings.Contains(function, "TestHelperCaller") {
		t.Errorf("Caller function = %q, want to contain 'TestHelperCaller'", function)
	}
}

// TestHelperPackageIsEmpty verifies package-level IsEmpty behavior.
func TestHelperPackageIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, true},
		{"empty std error", errors.New(""), true},
		{"whitespace error", errors.New("   "), true},
		{"non-empty std error", errors.New("test"), false},
		{"empty custom error", New(""), true},
		{"non-empty custom error", New("test"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if e, ok := tt.err.(*Error); ok {
				defer e.Free()
			}
			if got := IsEmpty(tt.err); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestHelperPackageIsNull verifies package-level IsNull behavior.
func TestHelperPackageIsNull(t *testing.T) {
	nullTime := sql.NullTime{Valid: false}
	validTime := sql.NullTime{Time: time.Now(), Valid: true}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, true},
		{"std error", errors.New("test"), false},
		{"custom error with NULL", New("").With("time", nullTime), true},
		{"custom error with valid", New("").With("time", validTime), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if e, ok := tt.err.(*Error); ok {
				defer e.Free()
			}
			if got := IsNull(tt.err); got != tt.expected {
				t.Errorf("IsNull() = %v, want %v", got, tt.expected)
			}
		})
	}
}
