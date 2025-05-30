// Package main demonstrates the use of WithStack from github.com/olekukonko/errors.
// It showcases adding stack traces to errors using both package-level and method-style approaches,
// comparing their application to standard and enhanced errors, and combining them in a real-world
// scenario with additional context and error details.
package main

import (
	"fmt"
	"time"

	"github.com/olekukonko/errors"
	"math/rand"
)

// basicFunc simulates a simple function returning a standard error.
// It represents a legacy or external function without enhanced error features.
func basicFunc() error {
	return fmt.Errorf("basic error") // Return a basic fmt.Errorf error
}

// enhancedFunc simulates a function returning an enhanced *errors.Error.
// It represents a function utilizing the errors package's custom error type.
func enhancedFunc() *errors.Error {
	return errors.New("enhanced error") // Return a new *errors.Error
}

// main is the entry point, demonstrating WithStack usage in various contexts.
// It tests package-level WithStack on standard errors, method-style WithStack on enhanced errors,
// and a combined approach in a practical scenario.
func main() {
	// 1. Package-level WithStack - works with ANY error type
	// Demonstrate adding a stack trace to a standard error
	err1 := basicFunc()
	enhanced1 := errors.WithStack(err1) // Convert and add stack trace to any error
	fmt.Println("Package-level WithStack:")
	fmt.Println(enhanced1.Stack()) // Print stack trace from standard error

	// 2. Method-style WithStack - only for *errors.Error
	// Show adding a stack trace to an enhanced error using method chaining
	err2 := enhancedFunc()
	enhanced2 := err2.WithStack() // Add stack trace to *errors.Error via method
	fmt.Println("\nMethod-style WithStack:")
	fmt.Println(enhanced2.Stack()) // Print stack trace from enhanced error

	// 3. Combined usage in real-world scenario
	// Test a mixed error type with both WithStack approaches and additional context
	result := processData()
	if result != nil {
		// Use package-level WithStack when error type is unknown
		stackErr := errors.WithStack(result)

		// Chain method-style enhancements on the resulting *errors.Error
		finalErr := stackErr.
			With("timestamp", time.Now()). // Add timestamp context
			WithCode(500)                  // Set HTTP-like status code

		fmt.Println("\nCombined Usage:")
		fmt.Println("Message:", finalErr.Error())   // Print full error message
		fmt.Println("Context:", finalErr.Context()) // Print context map
		fmt.Println("Stack:")
		for _, frame := range finalErr.Stack() {
			fmt.Println(frame) // Print each stack frame
		}
	}
}

// processData simulates a data processing function with variable error types.
// It randomly returns either a standard error or an enhanced error with context.
func processData() error {
	// Randomly choose between standard and enhanced error
	if rand.Intn(2) == 0 {
		return fmt.Errorf("database error") // Return standard error
	}
	return errors.New("validation error").With("field", "email") // Return enhanced error with context
}
