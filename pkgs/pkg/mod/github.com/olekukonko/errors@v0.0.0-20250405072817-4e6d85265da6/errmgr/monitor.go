// Package errmgr provides error monitoring functionality.
package errmgr

import (
	"github.com/olekukonko/errors"
	"sync"
)

const (
	monitorSize = 10
)

// alertChannel wraps a channel with synchronization for safe closure.
// Used internally by Monitor to manage alert delivery.
type alertChannel struct {
	ch     chan *errors.Error
	closed bool
	mu     sync.Mutex
}

// Monitor represents an error monitoring channel for a specific error name.
// It receives alerts when the error count exceeds a configured threshold set via SetThreshold.
type Monitor struct {
	name string
	ac   *alertChannel
}

// Alerts returns the channel for receiving error alerts.
// Alerts are sent when the error count exceeds the threshold set by SetThreshold.
// Returns nil if the monitor has been closed.
func (m *Monitor) Alerts() <-chan *errors.Error {
	m.ac.mu.Lock()
	defer m.ac.mu.Unlock()
	if m.ac.closed {
		return nil
	}
	return m.ac.ch
}

// Close shuts down the monitor channel and removes it from the registry.
// Thread-safe and idempotent; subsequent calls have no effect.
func (m *Monitor) Close() {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if existing, ok := registry.alerts.Load(m.name); ok {
		if ac, ok := existing.(*alertChannel); ok && ac == m.ac {
			ac.mu.Lock()
			if !ac.closed {
				close(ac.ch)
				ac.closed = true
			}
			ac.mu.Unlock()
			registry.alerts.Delete(m.name)
		}
	}
}

// IsClosed reports whether the monitor’s channel has been closed.
// Thread-safe; useful for checking monitor status before use.
func (m *Monitor) IsClosed() bool {
	m.ac.mu.Lock()
	defer m.ac.mu.Unlock()
	return m.ac.closed
}

// NewMonitor creates a new Monitor for the given error name with a default buffer of 10.
// Reuses an existing channel if one is already registered; thread-safe.
// Use NewMonitorBuffered for a custom buffer size.
func NewMonitor(name string) *Monitor {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if existing, ok := registry.alerts.Load(name); ok {
		return &Monitor{name: name, ac: existing.(*alertChannel)}
	}

	ac := &alertChannel{
		ch:     make(chan *errors.Error, monitorSize),
		closed: false,
	}
	registry.alerts.Store(name, ac)
	return &Monitor{name: name, ac: ac}
}

// NewMonitorBuffered creates a new Monitor for the given error name with a specified buffer size.
// Reuses an existing channel if one is already registered; thread-safe.
// Buffer must be non-negative (0 means unbuffered); use NewMonitor for the default buffer of 10.
func NewMonitorBuffered(name string, buffer int) *Monitor {
	if buffer < 0 {
		buffer = 0
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if existing, ok := registry.alerts.Load(name); ok {
		return &Monitor{name: name, ac: existing.(*alertChannel)}
	}

	ac := &alertChannel{
		ch:     make(chan *errors.Error, buffer),
		closed: false,
	}
	registry.alerts.Store(name, ac)
	return &Monitor{name: name, ac: ac}
}
