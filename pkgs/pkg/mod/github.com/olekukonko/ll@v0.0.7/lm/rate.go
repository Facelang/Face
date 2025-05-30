package lm

import (
	"fmt"
	"github.com/olekukonko/ll/lx"
	"sync"
	"time"
)

// RateLimiter is a middleware that limits the rate of log entries per level.
// It tracks log counts for each log level within a specified time interval,
// rejecting entries that exceed the allowed rate.
// Thread-safe with mutexes for concurrent access.
type RateLimiter struct {
	limits map[lx.LevelType]*rateLimit // Map of log levels to their rate limits
	mu     sync.Mutex                  // Protects concurrent access to limits map
}

// rateLimit holds rate limiting state for a specific log level.
// It tracks the current log count, maximum allowed logs, time interval,
// and the timestamp of the last log.
type rateLimit struct {
	count    int           // Current number of logs in the interval
	maxCount int           // Maximum allowed logs per interval
	interval time.Duration // Time window for rate limiting
	last     time.Time     // Time of the last log
	mu       sync.Mutex    // Protects concurrent access
}

// NewRateLimiter creates a new RateLimiter for a specific log level.
// It initializes the limiter with the given level, maximum log count, and time interval.
// The limiter can be extended to other levels using the Set method.
// Example:
//
//	limiter := NewRateLimiter(lx.LevelInfo, 10, time.Second)
//	logger := ll.New("app").Enable().Use(limiter)
//	logger.Info("Test") // Allowed up to 10 times per second
func NewRateLimiter(level lx.LevelType, count int, interval time.Duration) *RateLimiter {
	r := &RateLimiter{
		limits: make(map[lx.LevelType]*rateLimit), // Initialize empty limits map
	}
	// Set initial rate limit for the specified level
	r.Set(level, count, interval)
	return r
}

// Set configures a rate limit for a specific log level.
// It adds or updates the rate limit for the given level with the specified count and interval.
// Thread-safe with a mutex. Returns the RateLimiter for chaining.
// Example:
//
//	limiter := NewRateLimiter(lx.LevelInfo, 10, time.Second)
//	limiter.Set(lx.LevelWarn, 5, time.Minute) // Add limit for Warn level
func (rl *RateLimiter) Set(level lx.LevelType, count int, interval time.Duration) *RateLimiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	// Create or update rate limit for the level
	rl.limits[level] = &rateLimit{
		count:    0,          // Initialize count to zero
		maxCount: count,      // Set maximum allowed logs
		interval: interval,   // Set time window
		last:     time.Now(), // Set initial timestamp
	}
	return rl
}

// Handle processes a log entry and enforces rate limiting.
// It checks if the entry's level has a rate limit and verifies if the log count
// within the current interval exceeds the maximum allowed.
// Returns an error if the rate limit is exceeded, causing the log to be dropped.
// Thread-safe with mutexes for concurrent access.
// Example (internal usage):
//
//	err := limiter.Handle(&lx.Entry{Level: lx.LevelInfo}) // Returns error if limit exceeded
func (rl *RateLimiter) Handle(e *lx.Entry) error {
	rl.mu.Lock()
	limit, exists := rl.limits[e.Level] // Check if level has a rate limit
	rl.mu.Unlock()
	if !exists {
		return nil // No limit for this level, allow log
	}

	limit.mu.Lock()
	defer limit.mu.Unlock()
	now := time.Now()
	// Reset count if interval has passed
	if now.Sub(limit.last) >= limit.interval {
		limit.last = now
		limit.count = 0
	}
	limit.count++ // Increment log count
	// Check if limit is exceeded
	if limit.count > limit.maxCount {
		return fmt.Errorf("rate limit exceeded") // Drop log
	}
	return nil // Allow log
}
