// Package main demonstrates the retry functionality of github.com/olekukonko/errors.
// It simulates flaky database and external service operations with configurable retries,
// exponential backoff, jitter, and context timeouts, showcasing error handling, retry policies,
// and result capturing in various failure scenarios.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/olekukonko/errors"
)

// DatabaseClient simulates a flaky database connection with a recovery point.
// It fails until a specified number of attempts is reached, then succeeds.
type DatabaseClient struct {
	healthyAfterAttempt int // Number of attempts before becoming healthy
}

// Query attempts a database operation, failing until healthyAfterAttempt reaches zero.
// It returns a retryable error with remaining attempts context during failure.
func (db *DatabaseClient) Query() error {
	if db.healthyAfterAttempt > 0 {
		db.healthyAfterAttempt-- // Decrement failure counter
		return errors.New("database connection failed").
			With("attempt_remaining", db.healthyAfterAttempt). // Add remaining attempts context
			WithRetryable()                                    // Mark error as retryable
	}
	return nil // Success when attempts exhausted
}

// ExternalService simulates an unreliable external API with random failures.
// It fails 30% of the time, returning a retryable error with a 503 status code.
func ExternalService() error {
	if rand.Intn(100) < 30 { // 30% failure probability
		return errors.New("service unavailable").
			WithCode(503).  // Set HTTP 503 Service Unavailable status
			WithRetryable() // Mark error as retryable
	}
	return nil // Success on remaining 70%
}

// main is the entry point, demonstrating retry scenarios with database, external service, and timeout.
// It configures retries with backoff, jitter, and context, executing operations and reporting outcomes.
func main() {
	// Configure retry with exponential backoff and jitter
	// Set up a retry policy with custom parameters and logging
	retry := errors.NewRetry(
		errors.WithMaxAttempts(5),                       // Allow up to 5 attempts
		errors.WithDelay(200*time.Millisecond),          // Base delay of 200ms
		errors.WithMaxDelay(2*time.Second),              // Cap delay at 2s
		errors.WithJitter(true),                         // Add randomness to delays
		errors.WithBackoff(errors.ExponentialBackoff{}), // Use exponential backoff strategy
		errors.WithOnRetry(func(attempt int, err error) { // Callback on each retry
			// Calculate delay for logging, mirroring Execute logic
			baseDelay := 200 * time.Millisecond
			maxDelay := 2 * time.Second
			delay := errors.ExponentialBackoff{}.Backoff(attempt, baseDelay)
			if delay > maxDelay {
				delay = maxDelay
			}
			fmt.Printf("Attempt %d failed: %v (retrying in %v)\n",
				attempt, err.Error(), delay)
		}),
	)

	// Scenario 1: Database connection with known recovery point
	// Test retrying a database operation that recovers after 3 failures
	db := &DatabaseClient{healthyAfterAttempt: 3}
	fmt.Println("Starting database operation...")
	err := retry.Execute(func() error {
		return db.Query() // Attempt database query
	})
	if err != nil {
		fmt.Printf("Database operation failed after %d attempts: %v\n", retry.Attempts(), err)
	} else {
		fmt.Println("Database operation succeeded!") // Expect success after 4 attempts
	}

	// Scenario 2: External service with random failures
	// Test retrying an external service call with a 30% failure rate
	fmt.Println("\nStarting external service call...")
	var lastAttempts int // Track total attempts manually
	start := time.Now()  // Measure duration

	// Using ExecuteReply to capture both result and error
	result, err := errors.ExecuteReply[string](retry, func() (string, error) {
		lastAttempts++ // Increment attempt counter
		if err := ExternalService(); err != nil {
			return "", err // Return error on failure
		}
		return "service response data", nil // Return success data
	})

	duration := time.Since(start) // Calculate elapsed time
	if err != nil {
		fmt.Printf("Service call failed after %d attempts (%.2f sec): %v\n",
			lastAttempts, duration.Seconds(), err)
	} else {
		fmt.Printf("Service call succeeded after %d attempts (%.2f sec): %s\n",
			lastAttempts, duration.Seconds(), result) // Expect variable attempts
	}

	// Scenario 3: Context cancellation with more visibility
	// Test retrying an operation with a short timeout
	fmt.Println("\nStarting operation with timeout...")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond) // 500ms timeout
	defer cancel()                                                                 // Ensure context cleanup

	// Transform retry configuration with context and increased visibility
	timeoutRetry := retry.Transform(
		errors.WithContext(ctx),    // Apply timeout context
		errors.WithMaxAttempts(10), // Increase to 10 attempts
		errors.WithOnRetry(func(attempt int, err error) { // Log each retry attempt
			fmt.Printf("Timeout scenario attempt %d: %v\n", attempt, err)
		}),
	)

	startTimeout := time.Now() // Measure timeout scenario duration
	err = timeoutRetry.Execute(func() error {
		time.Sleep(300 * time.Millisecond)       // Simulate a long operation
		return errors.New("operation timed out") // Return consistent error
	})

	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Printf("Operation cancelled by timeout after %.2f sec: %v\n",
			time.Since(startTimeout).Seconds(), err) // Expect timeout cancellation
	} else if err != nil {
		fmt.Printf("Operation failed: %v\n", err)
	} else {
		fmt.Println("Operation succeeded (unexpected)") // Unlikely with 500ms timeout
	}
}
