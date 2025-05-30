package errors

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// Custom error type to create a chain of errors
type customError struct {
	msg   string
	cause error
}

func (e *customError) Error() string {
	return e.msg
}

func (e *customError) Cause() error {
	return e.cause
}

// TestErrorNew verifies that New creates an error with the correct message and no stack trace.
func TestErrorNew(t *testing.T) {
	err := New("test error")
	defer err.Free()
	if err.Error() != "test error" {
		t.Errorf("New() error message = %v, want %v", err.Error(), "test error")
	}
	if len(err.Stack()) != 0 {
		t.Errorf("New() should not capture stack trace, got %d frames", len(err.Stack()))
	}
}

// TestErrorNewf checks that Newf formats the message correctly and includes no stack trace.
func TestErrorNewf(t *testing.T) {
	err := Newf("test %s %d", "error", 42)
	defer err.Free()
	want := "test error 42"
	if err.Error() != want {
		t.Errorf("Newf() error message = %v, want %v", err.Error(), want)
	}
	if len(err.Stack()) != 0 {
		t.Errorf("Newf() should not capture stack trace, got %d frames", len(err.Stack()))
	}
}

// TestErrorNamed ensures Named sets the name correctly and captures a stack trace.
func TestErrorNamed(t *testing.T) {
	err := Named("test_name")
	defer err.Free()
	if err.Error() != "test_name" {
		t.Errorf("Named() error message = %v, want %v", err.Error(), "test_name")
	}
	if len(err.Stack()) == 0 {
		t.Errorf("Named() should capture stack trace")
	}
}

// TestErrorMethods tests various methods on the Error type.
func TestErrorMethods(t *testing.T) {
	err := New("base error")
	defer err.Free()

	err = err.With("key", "value")
	if err.Context()["key"] != "value" {
		t.Errorf("With() failed, context[key] = %v, want %v", err.Context()["key"], "value")
	}

	cause := New("cause error")
	defer cause.Free()
	err = err.Wrap(cause)
	if err.Unwrap() != cause {
		t.Errorf("Wrap() failed, unwrapped = %v, want %v", err.Unwrap(), cause)
	}

	err = err.Msgf("new message %d", 123)
	if err.Error() != "new message 123: cause error" {
		t.Errorf("Msgf() failed, error = %v, want %v", err.Error(), "new message 123: cause error")
	}

	stackLen := len(err.Stack())
	if stackLen != 0 {
		t.Errorf("Initial stack length should be 0, got %d", stackLen)
	}
	err = err.Trace()
	if len(err.Stack()) == 0 {
		t.Errorf("Trace() should capture a stack trace, got no frames")
	}

	err = err.WithCode(400)
	if err.Code() != 400 {
		t.Errorf("WithCode() failed, code = %d, want 400", err.Code())
	}

	err = err.WithCategory("test_category")
	if Category(err) != "test_category" {
		t.Errorf("WithCategory() failed, category = %v, want %v", Category(err), "test_category")
	}

	err = err.Increment()
	if err.Count() != 1 {
		t.Errorf("Increment() failed, count = %d, want 1", err.Count())
	}
}

// TestErrorIs verifies the Is method for matching errors by name and wrapping.
func TestErrorIs(t *testing.T) {
	err := Named("test_error")
	defer err.Free()
	err2 := Named("test_error")
	defer err2.Free()
	err3 := Named("other_error")
	defer err3.Free()

	if !err.Is(err2) {
		t.Errorf("Is() failed, %v should match %v", err, err2)
	}
	if err.Is(err3) {
		t.Errorf("Is() failed, %v should not match %v", err, err3)
	}

	wrappedErr := Named("wrapper")
	defer wrappedErr.Free()
	cause := Named("cause_error")
	defer cause.Free()
	wrappedErr = wrappedErr.Wrap(cause)
	t.Logf("Before Is(cause): wrappedErr.cause = %p, cause = %p", wrappedErr.cause, cause)
	if !wrappedErr.Is(cause) {
		t.Errorf("Is() failed, wrapped error should match cause; wrappedErr = %+v, cause = %+v", wrappedErr, cause)
	}

	stdErr := errors.New("std error")
	wrappedErr = wrappedErr.Wrap(stdErr)
	t.Logf("Before Is(stdErr): wrappedErr.cause = %p, stdErr = %p", wrappedErr.cause, stdErr)
	if !wrappedErr.Is(stdErr) {
		t.Errorf("Is() failed, should match stdlib error")
	}
}

// TestErrorAs checks the As method for unwrapping to the correct error type.
func TestErrorAs(t *testing.T) {
	err := New("base").Wrap(Named("target"))
	defer err.Free()
	var target *Error
	if !As(err, &target) {
		t.Errorf("As() failed, should unwrap to *Error")
	}
	if target.name != "target" {
		t.Errorf("As() unwrapped to wrong error, got %v, want %v", target.name, "target")
	}

	stdErr := errors.New("std error")
	err = New("wrapper").Wrap(stdErr)
	defer err.Free()
	var stdTarget error
	if !As(err, &stdTarget) {
		t.Errorf("As() failed, should unwrap to stdlib error")
	}
	if stdTarget != stdErr {
		t.Errorf("As() unwrapped to wrong error, got %v, want %v", stdTarget, stdErr)
	}
}

// TestErrorCount verifies the Count method for per-instance counting.
func TestErrorCount(t *testing.T) {
	err := New("unnamed")
	defer err.Free()
	if err.Count() != 0 {
		t.Errorf("Count() on new error should be 0, got %d", err.Count())
	}

	err = Named("test_count").Increment()
	if err.Count() != 1 {
		t.Errorf("Count() after Increment() should be 1, got %d", err.Count())
	}
}

// TestErrorCode checks the Code method for setting and retrieving HTTP status codes.
func TestErrorCode(t *testing.T) {
	err := New("unnamed")
	defer err.Free()
	if err.Code() != 0 {
		t.Errorf("Code() on new error should be 0, got %d", err.Code())
	}

	err = Named("test_code").WithCode(400)
	if err.Code() != 400 {
		t.Errorf("Code() after WithCode(400) should be 400, got %d", err.Code())
	}
}

// TestErrorMarshalJSON ensures JSON serialization includes all expected fields.
func TestErrorMarshalJSON(t *testing.T) {
	err := New("test").
		With("key", "value").
		WithCode(400).
		Wrap(Named("cause"))
	defer err.Free()
	data, e := json.Marshal(err)
	if e != nil {
		t.Fatalf("MarshalJSON() failed: %v", e)
	}

	want := map[string]interface{}{
		"message": "test",
		"context": map[string]interface{}{"key": "value"},
		"cause":   map[string]interface{}{"name": "cause"},
		"code":    float64(400),
	}
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got["message"] != want["message"] {
		t.Errorf("MarshalJSON() message = %v, want %v", got["message"], want["message"])
	}
	if !reflect.DeepEqual(got["context"], want["context"]) {
		t.Errorf("MarshalJSON() context = %v, want %v", got["context"], want["context"])
	}
	if cause, ok := got["cause"].(map[string]interface{}); !ok || cause["name"] != "cause" {
		t.Errorf("MarshalJSON() cause = %v, want %v", got["cause"], want["cause"])
	}
	if code, ok := got["code"].(float64); !ok || code != 400 {
		t.Errorf("MarshalJSON() code = %v, want %v", got["code"], 400)
	}

	t.Run("WithStack", func(t *testing.T) {
		err := New("test").WithStack().WithCode(500)
		defer err.Free()
		data, e := json.Marshal(err)
		if e != nil {
			t.Fatalf("MarshalJSON() failed: %v", e)
		}
		var got map[string]interface{}
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if _, ok := got["stack"].([]interface{}); !ok || len(got["stack"].([]interface{})) == 0 {
			t.Error("MarshalJSON() should include non-empty stack")
		}
		if code, ok := got["code"].(float64); !ok || code != 500 {
			t.Errorf("MarshalJSON() code = %v, want 500", got["code"])
		}
	})
}

// TestErrorEdgeCases verifies behavior in unusual scenarios.
func TestErrorEdgeCases(t *testing.T) {
	var nilErr *Error
	if nilErr.Is(nil) {
		t.Errorf("nil.Is(nil) should be false, got true")
	}
	if Is(nilErr, New("test")) {
		t.Errorf("Is(nil, non-nil) should be false")
	}

	err := New("empty name")
	defer err.Free()
	if err.Is(Named("")) {
		t.Errorf("Error with empty name should not match unnamed error")
	}

	stdErr := errors.New("std error")
	customErr := New("custom").Wrap(stdErr)
	defer customErr.Free()
	if !Is(customErr, stdErr) {
		t.Errorf("Is() should match stdlib error through wrapping")
	}

	var nilTarget *Error
	if As(nilErr, &nilTarget) {
		t.Errorf("As(nil, &nilTarget) should return false")
	}
}

// TestErrorRetryWithCallback tests retry functionality with a callback.
func TestErrorRetryWithCallback(t *testing.T) {
	attempts := 0
	retry := NewRetry(
		WithMaxAttempts(3),
		WithDelay(1*time.Millisecond),
		WithOnRetry(func(attempt int, err error) {
			attempts++
		}),
	)

	err := retry.Execute(func() error {
		return New("retry me").WithRetryable()
	})

	if attempts != 3 {
		t.Errorf("Expected 3 retry attempts, got %d", attempts)
	}
	if err == nil {
		t.Error("Expected retry to exhaust with error, got nil")
	}
}

// TestErrorStackPresence verifies stack presence for New and Trace.
func TestErrorStackPresence(t *testing.T) {
	err := New("test")
	if len(err.Stack()) != 0 {
		t.Error("New() should not capture stack")
	}

	traced := Trace("test")
	if len(traced.Stack()) == 0 {
		t.Error("Trace() should capture stack")
	}
}

// TestErrorStackDepth verifies stack depth doesn’t exceed the configured maximum.
func TestErrorStackDepth(t *testing.T) {
	err := Trace("test")
	frames := err.Stack()
	if len(frames) > currentConfig.stackDepth {
		t.Errorf("Stack depth %d exceeds configured max %d", len(frames), currentConfig.stackDepth)
	}
}

// TestErrorTransform verifies Transform behavior with nil, non-*Error, and *Error inputs.
func TestErrorTransform(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		result := Transform(nil, func(e *Error) {})
		if result != nil {
			t.Error("Should return nil for nil input")
		}
	})

	t.Run("NonErrorType", func(t *testing.T) {
		stdErr := errors.New("standard")
		transformed := Transform(stdErr, func(e *Error) {})
		if transformed == nil {
			t.Error("Should not return nil for non-nil input")
		}
		if transformed.Error() != "standard" {
			t.Errorf("Should preserve original message, got %q, want %q", transformed.Error(), "standard")
		}
		if transformed == stdErr {
			t.Error("Should return a new *Error, not the original")
		}
	})

	t.Run("TransformError", func(t *testing.T) {
		orig := New("original")
		defer orig.Free()
		transformed := Transform(orig, func(e *Error) {
			e.With("key", "value")
		})
		defer transformed.Free()

		if transformed == orig {
			t.Error("Should return a copy, not the original")
		}
		if transformed.Error() != "original" {
			t.Errorf("Should preserve original message, got %q, want %q", transformed.Error(), "original")
		}
		if transformed.Context()["key"] != "value" {
			t.Error("Should apply transformations, context missing 'key'='value'")
		}
	})
}

// TestErrorWalk verifies Walk traverses the error chain correctly.
func TestErrorWalk(t *testing.T) {
	err1 := &customError{msg: "first error", cause: nil}
	err2 := &customError{msg: "second error", cause: err1}
	err3 := &customError{msg: "third error", cause: err2}

	var errorsWalked []string
	Walk(err3, func(e error) {
		errorsWalked = append(errorsWalked, e.Error())
	})

	expected := []string{"third error", "second error", "first error"}
	if !reflect.DeepEqual(errorsWalked, expected) {
		t.Errorf("Walk() = %v; want %v", errorsWalked, expected)
	}
}

// TestErrorFind verifies Find locates the correct error in the chain.
func TestErrorFind(t *testing.T) {
	err1 := &customError{msg: "first error", cause: nil}
	err2 := &customError{msg: "second error", cause: err1}
	err3 := &customError{msg: "third error", cause: err2}

	found := Find(err3, func(e error) bool {
		return e.Error() == "second error"
	})

	if found == nil || found.Error() != "second error" {
		t.Errorf("Find() = %v; want 'second error'", found)
	}

	found = Find(err3, func(e error) bool {
		return e.Error() == "non-existent error"
	})

	if found != nil {
		t.Errorf("Find() = %v; want nil", found)
	}
}

// TestErrorTraceStackContent verifies Trace captures stack content correctly.
func TestErrorTraceStackContent(t *testing.T) {
	err := Trace("test")
	defer err.Free()
	frames := err.Stack()
	if len(frames) == 0 {
		t.Fatal("Trace() should capture stack frames")
	}
	found := false
	for _, frame := range frames {
		if strings.Contains(frame, "testing.tRunner") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Trace() stack does not contain testing.tRunner, got: %v", frames)
	}
}

// TestErrorWithStackContent verifies WithStack captures stack content correctly.
func TestErrorWithStackContent(t *testing.T) {
	err := New("test").WithStack()
	defer err.Free()
	frames := err.Stack()
	if len(frames) == 0 {
		t.Fatal("WithStack() should capture stack frames")
	}
	found := false
	for _, frame := range frames {
		if strings.Contains(frame, "testing.tRunner") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("WithStack() stack does not contain testing.tRunner, got: %v", frames)
	}
}

// TestErrorWrappingChain verifies a full error chain with wrapping.
func TestErrorWrappingChain(t *testing.T) {
	databaseErr := New("connection timeout").
		With("timeout_sec", 5).
		With("server", "db01.prod")
	defer databaseErr.Free()

	businessErr := New("failed to process user 12345").
		With("user_id", "12345").
		With("stage", "processing").
		Wrap(databaseErr)
	defer businessErr.Free()

	apiErr := New("API request failed").
		WithCode(500).
		WithStack().
		Wrap(businessErr)
	defer apiErr.Free()

	expectedFullMessage := "API request failed: failed to process user 12345: connection timeout"
	if apiErr.Error() != expectedFullMessage {
		t.Errorf("Full error message mismatch\ngot: %q\nwant: %q", apiErr.Error(), expectedFullMessage)
	}

	chain := UnwrapAll(apiErr)
	if len(chain) != 3 {
		t.Fatalf("Expected chain length 3, got %d", len(chain))
	}

	tests := []struct {
		index    int
		expected string
	}{
		{0, "API request failed"},
		{1, "failed to process user 12345"},
		{2, "connection timeout"},
	}
	for _, tt := range tests {
		if chain[tt.index].Error() != tt.expected {
			t.Errorf("Chain position %d mismatch\ngot: %q\nwant: %q", tt.index, chain[tt.index].Error(), tt.expected)
		}
	}

	if !errors.Is(apiErr, databaseErr) {
		t.Error("Is() should match the database error in the chain")
	}

	if ctx := businessErr.Context(); ctx["timeout_sec"] != nil {
		t.Error("Business error should not have database context")
	}

	if stack := apiErr.Stack(); len(stack) == 0 {
		t.Error("API error should have stack trace")
	}
	if stack := businessErr.Stack(); len(stack) != 0 {
		t.Error("Business error should not have stack trace")
	}

	if apiErr.Code() != 500 {
		t.Error("API error should have code 500")
	}
	if businessErr.Code() != 0 {
		t.Error("Business error should have no code")
	}
}

// TestErrorExampleOutput verifies the formatted output matches expectations.
func TestErrorExampleOutput(t *testing.T) {
	databaseErr := New("connection timeout").
		With("timeout_sec", 5).
		With("server", "db01.prod")
	businessErr := New("failed to process user 12345").
		With("user_id", "12345").
		With("stage", "processing").
		Wrap(databaseErr)
	apiErr := New("API request failed").
		WithCode(500).
		WithStack().
		Wrap(businessErr)

	chain := UnwrapAll(apiErr)
	for _, err := range chain {
		if e, ok := err.(*Error); ok {
			formatted := e.Format()
			if formatted == "" {
				t.Error("Format() returned empty string")
			}
			if !strings.Contains(formatted, "Error: "+e.Error()) {
				t.Errorf("Format() output missing error message: %q", formatted)
			}
			if e == apiErr {
				if !strings.Contains(formatted, "Code: 500") {
					t.Error("Format() missing code for API error")
				}
				if !strings.Contains(formatted, "Stack:") {
					t.Error("Format() missing stack for API error")
				}
			}
			if e == businessErr {
				if ctx := e.Context(); ctx != nil {
					if !strings.Contains(formatted, "Context:") {
						t.Error("Format() missing context for business error")
					}
					for k := range ctx {
						if !strings.Contains(formatted, k) {
							t.Errorf("Format() missing context key %q", k)
						}
					}
				}
			}
		}
	}

	if !errors.Is(apiErr, errors.New("connection timeout")) {
		t.Error("Is() failed to match connection timeout error")
	}
}

// TestErrorFullChain verifies a complex error chain with mixed error types.
func TestErrorFullChain(t *testing.T) {
	stdErr := errors.New("file not found")
	authErr := Named("AuthError").WithCode(401)
	storageErr := Wrapf(stdErr, "storage failed")
	authErrWrapped := Wrap(storageErr, authErr)
	wrapped := Wrapf(authErrWrapped, "request failed")

	var targetAuth *Error
	expectedTopLevelMsg := "request failed: AuthError: storage failed: file not found"
	if !errors.As(wrapped, &targetAuth) || targetAuth.Error() != expectedTopLevelMsg {
		t.Errorf("stderrors.As(wrapped, &targetAuth) failed, got %v, want %q", targetAuth.Error(), expectedTopLevelMsg)
	}

	var targetAuthPtr *Error
	if !As(wrapped, &targetAuthPtr) || targetAuthPtr.Name() != "AuthError" || targetAuthPtr.Code() != 401 {
		t.Errorf("As(wrapped, &targetAuthPtr) failed, got name=%s, code=%d; want AuthError, 401", targetAuthPtr.Name(), targetAuthPtr.Code())
	}

	if !Is(wrapped, authErr) {
		t.Errorf("Is(wrapped, authErr) failed, expected true")
	}
	if !errors.Is(wrapped, authErr) {
		t.Errorf("stderrors.Is(wrapped, authErr) failed, expected true")
	}
	if !Is(wrapped, stdErr) {
		t.Errorf("Is(wrapped, stdErr) failed, expected true")
	}
	if !errors.Is(wrapped, stdErr) {
		t.Errorf("stderrors.Is(wrapped, stdErr) failed, expected true")
	}

	chain := UnwrapAll(wrapped)
	if len(chain) != 4 {
		t.Errorf("UnwrapAll(wrapped) length = %d, want 4", len(chain))
	}
	expected := []string{
		"request failed",
		"AuthError",
		"storage failed",
		"file not found",
	}
	for i, err := range chain {
		if err.Error() != expected[i] {
			t.Errorf("UnwrapAll[%d] = %v, want %v", i, err.Error(), expected[i])
		}
	}
}

// TestErrorUnwrapAllMessageIsolation verifies message isolation in UnwrapAll.
func TestErrorUnwrapAllMessageIsolation(t *testing.T) {
	inner := New("inner")
	middle := New("middle").Wrap(inner)
	outer := New("outer").Wrap(middle)

	chain := UnwrapAll(outer)
	if chain[0].Error() != "outer" {
		t.Errorf("Expected 'outer', got %q", chain[0].Error())
	}
	if chain[1].Error() != "middle" {
		t.Errorf("Expected 'middle', got %q", chain[1].Error())
	}
	if chain[2].Error() != "inner" {
		t.Errorf("Expected 'inner', got %q", chain[2].Error())
	}
}

// TestErrorIsEmpty verifies IsEmpty behavior for various error states.
func TestErrorIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected bool
	}{
		{"nil error", nil, true},
		{"empty error", New(""), true},
		{"named empty", Named(""), true},
		{"with empty template", New("").WithTemplate(""), true},
		{"with message", New("test"), false},
		{"with name", Named("test"), false},
		{"with template", New("").WithTemplate("template"), false},
		{"with cause", New("").Wrap(New("cause")), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				defer tt.err.Free()
			}
			if got := tt.err.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestErrorIsNull verifies IsNull behavior for null and non-null errors.
func TestErrorIsNull(t *testing.T) {
	nullString := sql.NullString{Valid: false}
	validString := sql.NullString{String: "test", Valid: true}

	tests := []struct {
		name     string
		err      *Error
		expected bool
	}{
		{"nil error", nil, true},
		{"empty error", New(""), false},
		{"with NULL context", New("").With("data", nullString), true},
		{"with valid context", New("").With("data", validString), false},
		{"with NULL cause", New("").Wrap(New("NULL value").With("data", nullString)), true},
		{"with valid cause", New("").Wrap(New("valid value").With("data", validString)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				defer tt.err.Free()
			}
			if got := tt.err.IsNull(); got != tt.expected {
				t.Errorf("IsNull() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestErrorFromContext verifies FromContext enhances errors with context info.
func TestErrorFromContext(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		ctx := context.Background()
		if FromContext(ctx, nil) != nil {
			t.Error("Expected nil for nil input error")
		}
	})

	t.Run("deadline exceeded", func(t *testing.T) {
		deadline := time.Now().Add(-1 * time.Hour)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		err := errors.New("operation failed")
		cerr := FromContext(ctx, err)

		if !IsTimeout(cerr) {
			t.Error("Expected timeout error")
		}
		if !HasContextKey(cerr, "deadline") {
			t.Error("Expected deadline in context")
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := errors.New("operation failed")
		cerr := FromContext(ctx, err)

		if !HasContextKey(cerr, "cancelled") {
			t.Error("Expected cancelled flag")
		}
	})
}

// TestContextStorage checks `smallContext`and it's expansion
func TestContextStorage(t *testing.T) {
	t.Run("stores first 4 items in smallContext", func(t *testing.T) {
		err := New("test")

		err.With("a", 1)
		err.With("b", 2)
		err.With("c", 3)
		err.With("d", 4)

		if err.smallCount != 4 {
			t.Errorf("expected smallCount=4, got %d", err.smallCount)
		}
		if err.context != nil {
			t.Error("expected context map to be nil")
		}
	})

	t.Run("switches to map on 5th item", func(t *testing.T) {
		err := New("test")

		err.With("a", 1)
		err.With("b", 2)
		err.With("c", 3)
		err.With("d", 4)
		err.With("e", 5)

		if err.context == nil {
			t.Error("expected context map to be initialized")
		}
		if len(err.context) != 5 {
			t.Errorf("expected 5 items in map, got %d", len(err.context))
		}
	})

	t.Run("preserves all context items", func(t *testing.T) {
		err := New("test")
		items := []struct {
			k string
			v interface{}
		}{
			{"a", 1}, {"b", 2}, {"c", 3},
			{"d", 4}, {"e", 5}, {"f", 6},
		}

		for _, item := range items {
			err.With(item.k, item.v)
		}

		ctx := err.Context()
		if len(ctx) != len(items) {
			t.Errorf("expected %d items, got %d", len(items), len(ctx))
		}
		for _, item := range items {
			if val, ok := ctx[item.k]; !ok || val != item.v {
				t.Errorf("missing item %s in context", item.k)
			}
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		err := New("test")
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			err.With("a", 1)
			err.With("b", 2)
			err.With("c", 3)
		}()

		go func() {
			defer wg.Done()
			err.With("d", 4)
			err.With("e", 5)
			err.With("f", 6)
		}()

		wg.Wait()
		ctx := err.Context()
		if len(ctx) != 6 {
			t.Errorf("expected 6 items, got %d", len(ctx))
		}
	})
}
