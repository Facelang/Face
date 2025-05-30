// Package main demonstrates the use of MultiError with sampling from github.com/olekukonko/errors.
// It generates a large number of errors, applies sampling with a limit, and analyzes the results,
// showcasing error collection, custom formatting, and statistical reporting in a simulated error-heavy scenario.
package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/olekukonko/errors"
)

// main is the entry point, simulating error generation with sampling and reporting statistics.
// It creates a MultiError, populates it with sampled errors, and displays detailed analysis.
func main() {
	// Configuration
	totalErrors := 1000 // Total number of errors to generate
	sampleRate := 10    // Target sampling rate (10%)
	errorLimit := 50    // Maximum number of errors to store

	// Initialize with reproducible seed for demo purposes
	r := rand.New(rand.NewSource(42)) // Create a seeded random source for consistency
	start := time.Now()               // Record start time for performance measurement

	// Create MultiError with sampling
	// Configure MultiError with sampling rate, limit, random source, and custom formatter
	multi := errors.NewMultiError(
		errors.WithSampling(uint32(sampleRate)),            // Set sampling rate to 10%
		errors.WithLimit(errorLimit),                       // Cap stored errors at 50
		errors.WithRand(r),                                 // Use seeded random number generator
		errors.WithFormatter(createFormatter(totalErrors)), // Apply custom formatter with total
	)

	// Generate errors
	for i := 0; i < totalErrors; i++ {
		multi.Add(errors.Newf("operation %d failed", i)) // Add formatted error for each iteration
	}

	// Calculate statistics
	duration := time.Since(start)                                    // Calculate elapsed time
	sampledCount := multi.Count()                                    // Get number of sampled errors
	actualRate := float64(sampledCount) / float64(totalErrors) * 100 // Compute actual sampling percentage

	// Print results
	fmt.Println(multi)                                                           // Display sampled errors with custom format
	printStatistics(totalErrors, sampledCount, sampleRate, actualRate, duration) // Show statistical summary
	printErrorDistribution(multi, 5)                                             // Show distribution of first 5 errors
}

// createFormatter returns a formatter for MultiError that includes total error count.
// It generates a header for the error report, showing sampled vs. total errors.
func createFormatter(total int) errors.ErrorFormatter {
	return func(errs []error) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Sampled Error Report (%d/%d):\n", len(errs), total)) // Report sampled vs. total
		sb.WriteString("══════════════════════════════\n")                               // Add separator line
		return sb.String()
	}
}

// printStatistics displays statistical summary of error sampling.
// It reports total errors, sampled count, rates, duration, and analysis notes.
func printStatistics(total, sampled, targetRate int, actualRate float64, duration time.Duration) {
	fmt.Printf("\nStatistics:\n")
	fmt.Printf("├─ Total errors generated: %d\n", total)            // Show total errors created
	fmt.Printf("├─ Errors captured: %d (limit: %d)\n", sampled, 50) // Show sampled errors and limit
	fmt.Printf("├─ Target sampling rate: %d%%\n", targetRate)       // Show intended sampling rate
	fmt.Printf("├─ Actual sampling rate: %.1f%%\n", actualRate)     // Show achieved sampling rate
	fmt.Printf("├─ Processing time: %v\n", duration)                // Show time taken

	// Analyze sampling accuracy and limits
	switch {
	case sampled == 50 && actualRate < float64(targetRate):
		fmt.Printf("└─ Note: Hit storage limit - actual rate would be ~%.1f%% without limit\n",
			float64(targetRate)) // Note when limit caps sampling
	case actualRate < float64(targetRate)*0.8 || actualRate > float64(targetRate)*1.2:
		fmt.Printf("└─ ⚠️ Warning: Significant sampling deviation\n") // Warn on large deviation
	default:
		fmt.Printf("└─ Sampling within expected range\n") // Confirm normal sampling
	}
}

// printErrorDistribution displays a subset of errors with a progress bar visualization.
// It shows up to maxDisplay errors, indicating remaining count if truncated.
func printErrorDistribution(m *errors.MultiError, maxDisplay int) {
	errs := m.Errors() // Get all sampled errors
	if len(errs) == 0 {
		return // Skip if no errors
	}

	fmt.Printf("\nError Distribution (showing first %d):\n", maxDisplay) // Announce display limit
	for i, err := range errs {
		if i >= maxDisplay {
			fmt.Printf("└─ ... and %d more\n", len(errs)-maxDisplay) // Indicate remaining errors
			break
		}
		fmt.Printf("%s %v\n", getProgressBar(i, len(errs)), err) // Print error with progress bar
	}
}

// getProgressBar generates a visual progress bar for error distribution.
// It creates a fixed-width bar based on the index relative to total errors.
func getProgressBar(index, total int) string {
	const width = 10                                                                        // Set bar width to 10 characters
	pos := int(float64(index) / float64(total) * width)                                     // Calculate filled portion
	return fmt.Sprintf("├─%s%s┤", strings.Repeat("■", pos), strings.Repeat(" ", width-pos)) // Build bar with ■ and spaces
}
