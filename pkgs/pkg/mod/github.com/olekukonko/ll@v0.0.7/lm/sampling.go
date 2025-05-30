package lm

import (
	"fmt"
	"github.com/olekukonko/ll/lx"
	"math/rand"
	"sync"
)

// Sampling is a middleware that randomly samples log entries based on a rate per level.
// It allows logs to pass through with a specified probability, tracking rejected logs in stats.
// Thread-safe with a mutex for concurrent access to rates and stats maps.
type Sampling struct {
	rates map[lx.LevelType]float64 // Sampling rates per log level (0.0 to 1.0)
	stats map[lx.LevelType]int     // Count of rejected logs per level
	mu    sync.Mutex               // Protects concurrent access to rates and stats
}

// NewSampling creates a new Sampling middleware for a specific log level.
// It initializes the middleware with a sampling rate for the given level,
// allowing further configuration via the Set method.
// Example:
//
//	sampler := NewSampling(lx.LevelDebug, 0.1) // Sample 10% of Debug logs
//	logger := ll.New("app").Enable().Use(sampler)
//	logger.Debug("Test") // Passes with 10% probability
func NewSampling(level lx.LevelType, rate float64) *Sampling {
	s := &Sampling{
		rates: make(map[lx.LevelType]float64), // Initialize empty rates map
		stats: make(map[lx.LevelType]int),     // Initialize empty stats map
	}
	// Set initial sampling rate for the specified level
	s.Set(level, rate)
	return s
}

// Set configures a sampling rate for a specific log level.
// It adds or updates the sampling rate (0.0 to 1.0) for the given level,
// where 0.0 rejects all logs and 1.0 allows all logs.
// Thread-safe with a mutex. Returns the Sampling instance for chaining.
// Example:
//
//	sampler := NewSampling(lx.LevelDebug, 0.1)
//	sampler.Set(lx.LevelInfo, 0.5) // Sample 50% of Info logs
func (s *Sampling) Set(level lx.LevelType, rate float64) *Sampling {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rates[level] = rate // Set or update sampling rate
	return s
}

// Handle processes a log entry and applies sampling based on the level's rate.
// It generates a random number and compares it to the level's sampling rate,
// allowing the log if the random number is less than or equal to the rate.
// Rejected logs increment the stats counter. Returns an error for rejected logs.
// Thread-safe with a mutex for stats updates.
// Example (internal usage):
//
//	err := sampler.Handle(&lx.Entry{Level: lx.LevelDebug}) // Returns error if rejected
func (s *Sampling) Handle(e *lx.Entry) error {
	rate, exists := s.rates[e.Level] // Check if level has a sampling rate
	if !exists {
		// fmt.Printf("Sampling: No rate for level %v\n", e.Level)
		return nil // No sampling for this level, allow log
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	random := rand.Float64() // Generate random number (0.0 to 1.0)
	if random <= rate {
		// fmt.Printf("Sampling: rate=%v, random=%v, allowing log\n", rate, random)
		return nil // Allow log based on sampling rate
	}
	s.stats[e.Level]++ // Increment rejected log count
	// fmt.Printf("Sampling: rate=%v, random=%v, rejecting log\n", rate, random)
	return fmt.Errorf("sampling error") // Reject log
}

// GetStats returns a copy of the sampling statistics.
// It provides the count of rejected logs per level, ensuring thread-safety with a read lock.
// The returned map is safe for external use without affecting internal state.
// Example:
//
//	stats := sampler.GetStats() // Returns map of rejected log counts by level
func (s *Sampling) GetStats() map[lx.LevelType]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[lx.LevelType]int) // Create new map for copy
	// Copy stats to new map
	for k, v := range s.stats {
		result[k] = v
	}
	return result
}
