// Package main demonstrates the use of context in the errors package from
// github.com/olekukonko/errors. It showcases adding context to custom errors,
// accessing context through wrapped and converted errors, handling standard library
// errors, and working with complex context data structures.
package main

import (
	"fmt"
	"github.com/olekukonko/errors"
)

// processData simulates a data processing operation that fails, returning an error with context.
// It attaches retry-related metadata to the error for demonstration purposes.
func processData(id string, attempt int) error {
	// Create an error with processing-specific context
	return errors.New("processing failed").
		With("id", id).           // Add data identifier
		With("attempt", attempt). // Add attempt number
		With("retryable", true)   // Mark as retryable
}

// main is the entry point, illustrating various ways to work with error context.
// It demonstrates basic context addition, context preservation through wrapping,
// handling standard errors, and managing complex context data.
func main() {
	// 1. Basic context example
	// Create and display an error with simple key-value context
	err := processData("123", 3)
	fmt.Println("Error:", err)                        // Print error message
	fmt.Println("Full context:", errors.Context(err)) // Print all context as a map

	// 2. Accessing context through conversion
	// Wrap the error with fmt.Errorf and show context preservation
	rawErr := fmt.Errorf("wrapped: %w", err)
	fmt.Println("\nAfter wrapping with fmt.Errorf:")
	fmt.Println("Direct context access:", errors.Context(rawErr)) // Show context is unavailable directly
	e := errors.Convert(rawErr)
	fmt.Println("After conversion - context:", e.Context()) // Show context restored via conversion

	// 3. Standard library errors
	// Demonstrate that standard errors lack context
	stdErr := fmt.Errorf("standard error")
	if errors.Context(stdErr) == nil {
		fmt.Println("\nStandard library errors have no context") // Confirm no context exists
	}

	// 4. Adding context to standard errors
	// Convert a standard error and enrich it with context
	converted := errors.Convert(stdErr).
		With("source", "legacy"). // Add source information
		With("severity", "high")  // Add severity level
	fmt.Println("\nConverted standard error:")
	fmt.Println("Message:", converted.Error())   // Print original message
	fmt.Println("Context:", converted.Context()) // Print added context

	// 5. Complex context example
	// Create an error with nested and varied context data
	complexErr := errors.New("database operation failed").
		With("query", "SELECT * FROM users"). // Add SQL query string
		With("params", map[string]interface{}{
			"limit":  100, // Nested parameter: result limit
			"offset": 0,   // Nested parameter: result offset
		}).
		With("duration_ms", 45.2) // Add execution time in milliseconds
	fmt.Println("\nComplex error context:")
	for k, v := range errors.Context(complexErr) {
		fmt.Printf("%s: %v (%T)\n", k, v, v) // Print each context key-value pair with type
	}
}
