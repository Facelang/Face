// Package main demonstrates the use of MultiError from github.com/olekukonko/errors to handle
// multiple validation and system errors. It showcases form validation with custom formatting,
// error filtering, and system error aggregation with retryable conditions, illustrating error
// management in a user registration and system monitoring context.
package main

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/olekukonko/errors"
)

// UserForm represents a user registration form with fields to validate.
type UserForm struct {
	Name     string
	Email    string
	Password string
	Birthday string
}

// validateUser validates a UserForm and returns a MultiError containing all validation failures.
// It checks name, email, password, and birthday fields, accumulating errors with a custom limit.
func validateUser(form UserForm) *errors.MultiError {
	// Initialize a MultiError with a limit of 10 errors and custom formatting
	multi := errors.NewMultiError(
		errors.WithLimit(10),               // Cap the number of errors at 10
		errors.WithFormatter(customFormat), // Use custom validation error format
	)

	// Name validation
	if form.Name == "" {
		multi.Add(errors.New("name is required")) // Add error for empty name
	} else if len(form.Name) > 50 {
		multi.Add(errors.New("name cannot exceed 50 characters")) // Add error for long name
	}

	// Email validation
	if form.Email == "" {
		multi.Add(errors.New("email is required")) // Add error for empty email
	} else {
		if _, err := mail.ParseAddress(form.Email); err != nil {
			multi.Add(errors.New("invalid email format")) // Add error for invalid email
		}
		if !strings.Contains(form.Email, "@") {
			multi.Add(errors.New("email must contain @ symbol")) // Add error for missing @
		}
	}

	// Password validation
	if len(form.Password) < 8 {
		multi.Add(errors.New("password must be at least 8 characters")) // Add error for short password
	}
	if !strings.ContainsAny(form.Password, "0123456789") {
		multi.Add(errors.New("password must contain at least one number")) // Add error for no digits
	}
	if !strings.ContainsAny(form.Password, "!@#$%^&*") {
		multi.Add(errors.New("password must contain at least one special character")) // Add error for no special chars
	}

	// Birthday validation
	if form.Birthday != "" {
		if _, err := time.Parse("2006-01-02", form.Birthday); err != nil {
			multi.Add(errors.New("birthday must be in YYYY-MM-DD format")) // Add error for invalid date format
		} else if bday, _ := time.Parse("2006-01-02", form.Birthday); time.Since(bday).Hours()/24/365 < 13 {
			multi.Add(errors.New("must be at least 13 years old")) // Add error for age under 13
		}
	}

	return multi
}

// customFormat formats a slice of validation errors into a user-friendly string.
// It adds a header, numbered list, and total count for display purposes.
func customFormat(errs []error) string {
	var sb strings.Builder
	sb.WriteString("🚨 Validation Errors:\n") // Add header with emoji
	for i, err := range errs {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err)) // List each error with number
	}
	sb.WriteString(fmt.Sprintf("\nTotal issues found: %d\n", len(errs))) // Append total count
	return sb.String()
}

// main is the entry point, demonstrating MultiError usage for validation and system errors.
// It validates a user form, analyzes errors, and aggregates system errors with retryable filtering.
func main() {
	fmt.Println("=== User Registration Validation ===")

	// Define a user form with intentional validation failures
	user := UserForm{
		Name:     "",              // Empty name to trigger error
		Email:    "invalid-email", // Invalid email format
		Password: "weak",          // Weak password
		Birthday: "2015-01-01",    // Date making user under 13
	}

	// Generate and display validation errors
	validationErrors := validateUser(user)

	if validationErrors.Has() {
		fmt.Println(validationErrors) // Print all validation errors

		// Detailed error analysis
		fmt.Println("\n🔍 Error Analysis:")
		fmt.Printf("Total errors: %d\n", validationErrors.Count()) // Show total error count
		fmt.Printf("First error: %v\n", validationErrors.First())  // Show first error
		fmt.Printf("Last error: %v\n", validationErrors.Last())    // Show last error

		// Categorize and display errors with consistent formatting
		fmt.Println("\n📋 Error Categories:")
		if emailErrors := validationErrors.Filter(contains("email")); emailErrors.Has() {
			fmt.Println("Email Issues:")
			if emailErrors.Count() == 1 {
				fmt.Println(customFormat([]error{emailErrors.First()})) // Format single email error
			} else {
				fmt.Println(emailErrors) // Print multiple email errors
			}
		}
		if pwErrors := validationErrors.Filter(contains("password")); pwErrors.Has() {
			fmt.Println("Password Issues:")
			if pwErrors.Count() == 1 {
				fmt.Println(customFormat([]error{pwErrors.First()})) // Format single password error
			} else {
				fmt.Println(pwErrors) // Print multiple password errors
			}
		}
		if ageErrors := validationErrors.Filter(contains("13 years")); ageErrors.Has() {
			fmt.Println("Age Restriction:")
			if ageErrors.Count() == 1 {
				fmt.Println(customFormat([]error{ageErrors.First()})) // Format single age error
			} else {
				fmt.Println(ageErrors) // Print multiple age errors
			}
		}
	}

	// System Error Aggregation Example
	fmt.Println("\n=== System Error Aggregation ===")
	// Initialize a MultiError for system errors with a limit and custom format
	systemErrors := errors.NewMultiError(
		errors.WithLimit(5),                     // Cap at 5 errors
		errors.WithFormatter(systemErrorFormat), // Use system error formatting
	)

	// Simulate various system errors
	systemErrors.Add(errors.New("database connection timeout").WithRetryable()) // Add retryable DB error
	systemErrors.Add(errors.New("API rate limit exceeded").WithRetryable())     // Add retryable API error
	systemErrors.Add(errors.New("disk space low"))                              // Add non-retryable error
	systemErrors.Add(errors.New("database connection timeout").WithRetryable()) // Add duplicate DB error
	systemErrors.Add(errors.New("cache miss"))                                  // Add another error
	systemErrors.Add(errors.New("database connection timeout").WithRetryable()) // Add over limit, ignored

	fmt.Println(systemErrors)                                               // Print system errors
	fmt.Printf("\nSystem Status: %d active issues\n", systemErrors.Count()) // Show active error count

	// Filter and display retryable errors
	if retryable := systemErrors.Filter(errors.IsRetryable); retryable.Has() {
		fmt.Println("\n🔄 Retryable Errors:")
		fmt.Println(retryable) // Print only retryable errors
	}
}

// systemErrorFormat formats a slice of system errors with retryable indicators.
// It creates a numbered list with a header, marking retryable errors explicitly.
func systemErrorFormat(errs []error) string {
	var sb strings.Builder
	sb.WriteString("⚠️ System Alerts:\n") // Add header with emoji
	for i, err := range errs {
		sb.WriteString(fmt.Sprintf("  %d. %s", i+1, err)) // List each error with number
		if errors.IsRetryable(err) {
			sb.WriteString(" (retryable)") // Mark as retryable if applicable
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// contains returns a predicate function to filter errors containing a substring.
// It’s used to categorize errors based on their message content.
func contains(substr string) func(error) bool {
	return func(err error) bool {
		return strings.Contains(err.Error(), substr) // Check if error message contains substring
	}
}
