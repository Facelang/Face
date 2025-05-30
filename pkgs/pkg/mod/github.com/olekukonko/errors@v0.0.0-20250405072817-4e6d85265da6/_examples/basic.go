// Package main demonstrates basic usage of the errors package from github.com/olekukonko/errors.
// It showcases creating simple errors, formatted errors, errors with stack traces, and named errors,
// highlighting the package’s enhanced error handling capabilities.
package main

import (
	"fmt"
	"github.com/olekukonko/errors"
)

// main is the entry point of the program, illustrating various ways to create and use errors
// from the errors package, printing their outputs to demonstrate their behavior.
func main() {
	// Simple error (no stack trace, fast)
	// Creates a lightweight error without capturing a stack trace for optimal performance.
	err := errors.New("connection failed")
	fmt.Println(err) // Outputs: "connection failed"

	// Formatted error
	// Creates an error with a formatted message using a printf-style syntax, similar to fmt.Errorf.
	err = errors.Newf("user %s not found", "bob")
	fmt.Println(err) // Outputs: "user bob not found"

	// Error with stack trace
	// Creates an error and captures a stack trace at the point of creation for debugging purposes.
	err = errors.Trace("critical issue")
	fmt.Println(err)         // Outputs: "critical issue"
	fmt.Println(err.Stack()) // Outputs stack trace, e.g., ["main.go:15", "caller.go:42"]

	// Named error
	// Creates an error with a specific name and stack trace, useful for categorizing errors.
	err = errors.Named("InputError")
	fmt.Println(err.Name()) // Outputs: "InputError"
}
