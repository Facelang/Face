// Package main demonstrates the use of the IsNull method from github.com/olekukonko/errors.
// It tests various scenarios involving nil errors, empty errors, errors with null or non-null
// context, and MultiError instances, showcasing how IsNull determines nullability based on content
// and context, particularly with SQL null types.
package main

import (
	"database/sql"
	"fmt"
	"github.com/olekukonko/errors"
)

// main is the entry point, illustrating the behavior of IsNull across different error cases.
// It checks nil errors, empty errors, context with SQL null values, and MultiError instances.
func main() {
	// Case 1: Nil error
	// Test if a nil error is considered null
	var err error = nil
	if errors.IsNull(err) {
		fmt.Println("Nil error is null") // Expect true: nil errors are always null
	}

	// Case 2: Empty error
	// Test if an empty error (no message) is considered null
	err = errors.New("")
	if errors.IsNull(err) {
		fmt.Println("Empty error is null")
	} else {
		fmt.Println("Empty error is not null") // Expect false: empty message but no null context
	}

	// Case 3: Error with null context
	// Test if an error with a null SQL context value is considered null
	nullString := sql.NullString{Valid: false}
	err = errors.New("").With("data", nullString)
	if errors.IsNull(err) {
		fmt.Println("Error with null context is null") // Expect true: all context is null
	}

	// Case 4: Error with non-null context
	// Test if an error with a valid SQL context value is not null
	validString := sql.NullString{String: "test", Valid: true}
	err = errors.New("").With("data", validString)
	if errors.IsNull(err) {
		fmt.Println("Error with valid context is null")
	} else {
		fmt.Println("Error with valid context is not null") // Expect false: valid context present
	}

	// Case 5: Empty MultiError
	// Test if an empty MultiError is considered null
	multi := errors.NewMultiError()
	if multi.IsNull() {
		fmt.Println("Empty MultiError is null") // Expect true: no errors in MultiError
	}

	// Case 6: MultiError with null error
	// Test if a MultiError containing a null error is considered null
	multi.Add(errors.New("").With("data", nullString))
	if multi.IsNull() {
		fmt.Println("MultiError with null error is null") // Expect true: only null errors
	}

	// Case 7: MultiError with non-null error
	// Test if a MultiError with mixed errors (null and non-null) is not null
	multi.Add(errors.New("real error"))
	if multi.IsNull() {
		fmt.Println("MultiError with mixed errors is null")
	} else {
		fmt.Println("MultiError with mixed errors is not null") // Expect false: contains non-null error
	}
}
