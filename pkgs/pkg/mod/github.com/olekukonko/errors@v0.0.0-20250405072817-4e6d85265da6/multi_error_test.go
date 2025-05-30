// multi_error_test.go
package errors

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
)

// TestMultiError_Basic verifies basic MultiError functionality.
// Ensures empty creation, nil error handling, and single error addition work as expected.
func TestMultiError_Basic(t *testing.T) {
	m := NewMultiError()
	if m.Has() {
		t.Error("New MultiError should be empty")
	}

	m.Add(nil)
	if m.Has() {
		t.Error("Adding nil should not create error")
	}

	err1 := errors.New("error 1")
	m.Add(err1)
	if !m.Has() {
		t.Error("Should detect errors after adding one")
	}
	if m.Count() != 1 {
		t.Errorf("Count should be 1, got %d", m.Count())
	}
	if m.First() != err1 || m.Last() != err1 {
		t.Errorf("First() and Last() should both be %v, got First=%v, Last=%v", err1, m.First(), m.Last())
	}
}

// TestMultiError_Sampling tests the sampling behavior of MultiError.
// Adds many unique errors with a 50% sampling rate and checks the resulting ratio is within 45-55%.
func TestMultiError_Sampling(t *testing.T) {
	r := rand.New(rand.NewSource(42)) // Fixed seed for reproducible results
	m := NewMultiError(WithSampling(50), WithRand(r))
	total := 1000

	for i := 0; i < total; i++ {
		m.Add(errors.New(fmt.Sprintf("test%d", i))) // Use unique errors to avoid duplicate filtering
	}

	count := m.Count()
	ratio := float64(count) / float64(total)
	// Expect roughly 50% (±5%) due to sampling; adjust range if sampling logic changes
	if ratio < 0.45 || ratio > 0.55 {
		t.Errorf("Sampling ratio %v not within expected range (45-55%%), count=%d, total=%d", ratio, count, total)
	}
}

// TestMultiError_Limit tests the error limit enforcement of MultiError.
// Adds twice the limit of unique errors and verifies the count caps at the limit.
func TestMultiError_Limit(t *testing.T) {
	limit := 10
	m := NewMultiError(WithLimit(limit))

	for i := 0; i < limit*2; i++ {
		m.Add(errors.New(fmt.Sprintf("test%d", i))) // Use unique errors to avoid duplicate filtering
	}

	if m.Count() != limit {
		t.Errorf("Should cap at %d errors, got %d", limit, m.Count())
	}
}

// TestMultiError_Formatting verifies custom formatting in MultiError.
// Adds two errors and checks the custom formatter outputs the expected string.
func TestMultiError_Formatting(t *testing.T) {
	customFormat := func(errs []error) string {
		return fmt.Sprintf("custom: %d", len(errs))
	}

	m := NewMultiError(WithFormatter(customFormat))
	m.Add(errors.New("test1"))
	m.Add(errors.New("test2"))

	expected := "custom: 2"
	if m.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, m.Error())
	}
}

// TestMultiError_Filter tests the filtering functionality of MultiError.
// Adds three errors, filters out one, and verifies the resulting count is correct.
func TestMultiError_Filter(t *testing.T) {
	m := NewMultiError()
	m.Add(errors.New("error1"))
	m.Add(errors.New("skip"))
	m.Add(errors.New("error2"))

	filtered := m.Filter(func(err error) bool {
		return err.Error() != "skip"
	})

	if filtered.Count() != 2 {
		t.Errorf("Should filter out one error, leaving 2, got %d", filtered.Count())
	}
}

// TestMultiError_AsSingle tests the Single() method across different scenarios.
// Verifies behavior for empty, single-error, and multi-error cases.
func TestMultiError_AsSingle(t *testing.T) {
	// Subtest: Empty MultiError should return nil
	t.Run("Empty", func(t *testing.T) {
		m := NewMultiError()
		if m.Single() != nil {
			t.Errorf("Empty should return nil, got %v", m.Single())
		}
	})

	// Subtest: Single error should return that error
	t.Run("Single", func(t *testing.T) {
		m := NewMultiError()
		err := errors.New("test")
		m.Add(err)
		if m.Single() != err {
			t.Errorf("Should return single error %v, got %v", err, m.Single())
		}
	})

	// Subtest: Multiple errors should return the MultiError itself
	t.Run("Multiple", func(t *testing.T) {
		m := NewMultiError()
		m.Add(errors.New("test1"))
		m.Add(errors.New("test2"))
		if m.Single() != m {
			t.Errorf("Should return self for multiple errors, got %v", m.Single())
		}
	})
}
