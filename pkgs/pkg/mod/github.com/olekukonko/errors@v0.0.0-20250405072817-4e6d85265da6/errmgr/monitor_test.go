package errmgr

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMonitorAlerts(t *testing.T) {
	Reset()
	monitor := NewMonitor("TestError")
	SetThreshold("TestError", 2)
	defer monitor.Close()

	errFunc := Define("TestError", "test error %d")
	for i := 0; i < 3; i++ {
		err := errFunc(i)
		if err.Name() != "TestError" {
			t.Errorf("Expected error name 'TestError', got %q", err.Name())
		}
		err.Free()
	}

	select {
	case alert := <-monitor.Alerts():
		if alert == nil {
			t.Fatal("Received nil alert after threshold exceeded")
		}
		if alert.Name() != "TestError" {
			t.Errorf("Expected alert name 'TestError', got %q", alert.Name())
		}
		if alert.Count() < 2 {
			t.Errorf("Expected alert count >= 2, got %d", alert.Count())
		}
		if !strings.Contains(alert.Error(), "threshold") {
			t.Errorf("Expected threshold message in alert, got %q", alert.Error())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("No alert received within 100ms timeout")
	}
}

func TestMonitorBuffered(t *testing.T) {
	Reset()
	monitor := NewMonitorBuffered("BufferedError", 2) // Buffer size 2
	SetThreshold("BufferedError", 1)
	defer monitor.Close()

	errFunc := Define("BufferedError", "buffered error %d")
	var wg sync.WaitGroup
	wg.Add(1) // Single goroutine

	go func() {
		defer wg.Done()
		for i := 0; i < 4; i++ { // Generate 4 errors
			err := errFunc(i)
			t.Logf("Generated error %d, count now %d", i, registry.counts.Value("BufferedError"))
			err.Free()
			time.Sleep(10 * time.Millisecond) // Slow down to fill buffer
		}
	}()

	// Wait for all errors to be generated
	wg.Wait()

	// Check metrics to confirm all 4 errors were counted
	counts := Metrics()
	if count, ok := counts["BufferedError"]; !ok || count != 4 {
		t.Errorf("Expected count 4 for BufferedError, got %v", counts)
	}

	// Consume alerts (expect up to 2 due to buffer size)
	received := 0
	timeout := time.After(200 * time.Millisecond)
	for received < 2 { // Expect at least 2 alerts
		select {
		case alert := <-monitor.Alerts():
			if alert == nil {
				t.Fatal("Received nil alert")
			}
			received++
			t.Logf("Received alert %d: %s", received, alert.Error())
			if alert.Name() != "BufferedError" {
				t.Errorf("Expected alert name 'BufferedError', got %q", alert.Name())
			}
		case <-timeout:
			t.Logf("Timeout waiting for alerts; received %d", received)
			break // Allow partial success if buffer limited alerts
		}
	}
}

func TestMonitorChannelCloseRace(t *testing.T) {
	Reset()
	SetThreshold("RaceError", 1)

	// Create and immediately close monitor to simulate quick close
	monitor := NewMonitor("RaceError")
	monitor.Close()

	// Ensure no panic when sending to closed channel
	errFunc := Define("RaceError", "race test %d")
	for i := 0; i < 3; i++ {
		err := errFunc(i)
		err.Free()
	}

	// Create new monitor and verify it works
	newMonitor := NewMonitor("RaceError")
	defer newMonitor.Close()

	err := errFunc(42)
	err.Free()

	select {
	case alert := <-newMonitor.Alerts():
		if alert == nil {
			t.Fatal("Received nil alert after reopening monitor")
		}
		if alert.Name() != "RaceError" {
			t.Errorf("Expected alert name 'RaceError', got %q", alert.Name())
		}
		if alert.Count() < 1 {
			t.Errorf("Expected alert count >= 1, got %d", alert.Count())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("No alert received within 100ms timeout")
	}

	if !monitor.IsClosed() {
		t.Error("Original monitor should be closed")
	}
	if newMonitor.IsClosed() {
		t.Error("New monitor should not be closed yet")
	}
}

func TestMonitorIsClosed(t *testing.T) {
	Reset()
	monitor := NewMonitor("CloseTest")
	if monitor.IsClosed() {
		t.Error("New monitor should not be closed")
	}

	monitor.Close()
	if !monitor.IsClosed() {
		t.Error("Monitor should be closed after Close()")
	}

	if ch := monitor.Alerts(); ch != nil {
		t.Error("Alerts should return nil after closure")
	}
}

func TestMonitorMultipleInstances(t *testing.T) {
	Reset()
	monitor1 := NewMonitor("MultiTest")
	monitor2 := NewMonitor("MultiTest") // Shares the same channel
	SetThreshold("MultiTest", 1)
	defer monitor1.Close()

	errFunc := Define("MultiTest", "multi test %d")
	err := errFunc(1)
	err.Free()

	// Consume from monitor1, expect monitor2 to see no alerts (single channel)
	select {
	case alert1 := <-monitor1.Alerts():
		if alert1 == nil {
			t.Fatal("Received nil alert from monitor1")
		}
		if alert1.Name() != "MultiTest" {
			t.Errorf("Expected alert name 'MultiTest', got %q", alert1.Name())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("No alert received from monitor1 within timeout")
	}

	// Verify monitor2 doesn't receive the same alert (already consumed)
	select {
	case alert2 := <-monitor2.Alerts():
		t.Errorf("Unexpected alert from monitor2: %v (channel should be drained)", alert2)
	case <-time.After(50 * time.Millisecond):
		// Expected: no alert since monitor1 consumed it
	}

	// Generate another error to ensure both monitors share the same channel
	err = errFunc(2)
	err.Free()

	select {
	case alert2 := <-monitor2.Alerts():
		if alert2 == nil {
			t.Fatal("Received nil alert from monitor2")
		}
		if alert2.Name() != "MultiTest" {
			t.Errorf("Expected alert name 'MultiTest', got %q", alert2.Name())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("No alert received from monitor2 within timeout")
	}
}
