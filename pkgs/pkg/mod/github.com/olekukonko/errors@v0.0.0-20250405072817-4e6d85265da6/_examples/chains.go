// Package main demonstrates error wrapping and chaining using the errors package from
// github.com/olekukonko/errors. It simulates a layered application (API, business logic,
// database) where errors are created, enriched with context, and wrapped, then inspected
// to show the full error chain and specific error checks.
package main

import (
	"fmt"
	"github.com/olekukonko/errors"
)

// databaseQuery simulates a database operation that fails, returning an error with context.
// It represents the lowest layer where an error originates, enriched with database-specific details.
func databaseQuery() error {
	// Create a base error with database failure details
	return errors.New("connection timeout").
		With("timeout_sec", 5).     // Add timeout duration context
		With("server", "db01.prod") // Add server identifier context
}

// businessLogic processes a user request and handles database errors, wrapping them with business context.
// It represents an intermediate layer that adds its own context without modifying the original error.
func businessLogic(userID string) error {
	err := databaseQuery()
	if err != nil {
		// Create a new error specific to business logic failure
		return errors.New("failed to process user "+userID).
			With("user_id", userID).     // Add user ID context
			With("stage", "processing"). // Add processing stage context
			Wrap(err)                    // Wrap the database error for chaining
	}
	return nil
}

// apiHandler simulates an API request handler that wraps business logic errors with API context.
// It represents the top layer, adding a status code and stack trace for debugging.
func apiHandler() error {
	err := businessLogic("12345")
	if err != nil {
		// Create a new error specific to API failure
		return errors.New("API request failed").
			WithCode(500). // Set HTTP-like status code
			WithStack().   // Capture stack trace at API level
			Wrap(err)      // Wrap the business logic error
	}
	return nil
}

// main is the entry point, demonstrating error creation, wrapping, and inspection.
// It prints the combined error message, unwraps the error chain, and checks for a specific error.
func main() {
	err := apiHandler()

	// Print the full error message combining all wrapped errors
	fmt.Println("=== Combined Message ===")
	fmt.Println(err)

	// Unwrap and display each error in the chain with its details
	fmt.Println("\n=== Error Chain ===")
	for i, e := range errors.UnwrapAll(err) {
		fmt.Printf("%d. %T\n", i+1, e) // Show error index and type
		if err, ok := e.(*errors.Error); ok {
			fmt.Println(err.Format()) // Print formatted details for custom errors
		} else {
			fmt.Println(e) // Print standard error message for non-custom errors
		}
	}

	// Check if the error chain contains a specific error
	fmt.Println("\n=== Error Checks ===")
	if errors.Is(err, errors.New("connection timeout")) {
		fmt.Println("✓ Matches connection timeout error") // Confirm match with database error
	}
}
