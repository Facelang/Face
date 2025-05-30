// ll_ns_test.go
package tests

import (
	"bytes"
	"github.com/olekukonko/ll"
	"github.com/olekukonko/ll/lh"
	"github.com/olekukonko/ll/lx"
	"strings"
	"testing"
)

// TestNamespaceEnableWithCustomSeparator verifies that enabling a namespace with a custom separator
// enables logging for that namespace and its descendants, even if the logger is initially disabled.
func TestNamespaceEnableWithCustomSeparator(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := ll.New("base").Disable().Separator(lx.Dot).Handler(lh.NewTextHandler(buf))

	// Child loggers inherit disabled state
	a := logger.Namespace("a").Handler(lh.NewTextHandler(buf))       // base.a
	b := logger.Namespace("b").Handler(lh.NewTextHandler(buf))       // base.b
	c := logger.Namespace("c.1").Handler(lh.NewTextHandler(buf))     // base.c.1
	d := logger.Namespace("c.1.2.4").Handler(lh.NewTextHandler(buf)) // base.c.1.2.4

	// Enable "c.1" sub-namespace (full path: base.c.1)
	logger.NamespaceEnable("c.1")

	// Verify namespace enabled state
	if !logger.NamespaceEnabled("c.1") {
		t.Errorf("Expected namespace 'c.1' (base.c.1) to be enabled")
	}
	if !c.NamespaceEnabled("") { // Checks base.c.1
		t.Errorf("Expected logger 'c' (base.c.1) to be enabled")
	}

	// Test logging
	buf.Reset()
	a.Infof("hello a from custom sep")
	b.Infof("hello b from custom sep")
	c.Infof("hello c from custom sep")
	d.Infof("hello d from custom sep")
	output := buf.String()

	// Expected logs
	expectedLogs := []string{
		"[base.c.1] INFO: hello c from custom sep",
		"[base.c.1.2.4] INFO: hello d from custom sep",
	}
	for _, expected := range expectedLogs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got %q", expected, output)
		}
	}

	// Unexpected logs
	unexpectedLogs := []string{
		"hello a from custom sep",
		"hello b from custom sep",
	}
	for _, unexpected := range unexpectedLogs {
		if strings.Contains(output, unexpected) {
			t.Errorf("Unexpected log %q in output: %q", unexpected, output)
		}
	}
}

// TestNamespaces verifies namespace creation, logging styles, and enable/disable behavior.
func TestNamespaces(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := ll.New("parent").Enable().Handler(lh.NewTextHandler(buf)) // Default "/" separator

	// Child logger inherits enabled state
	child := logger.Namespace("child").Handler(lh.NewTextHandler(buf))

	// Test flat path logging
	child.Style(lx.FlatPath)
	buf.Reset()
	child.Infof("Child log")
	expectedFlatLog := "[parent/child] INFO: Child log"
	if !strings.Contains(buf.String(), expectedFlatLog) {
		t.Errorf("Expected %q, got %q", expectedFlatLog, buf.String())
	}

	// Test nested path logging
	logger.Style(lx.NestedPath)
	child.Style(lx.NestedPath)
	buf.Reset()
	child.Infof("Nested log")
	expectedNestedLog := expectedNestedLogPrefix(child, lx.Arrow) + "INFO: Nested log"
	if !strings.Contains(buf.String(), expectedNestedLog) {
		t.Errorf("Expected %q, got %q", expectedNestedLog, buf.String())
	}

	// Test NamespaceDisable
	logger.NamespaceDisable("child") // Disables parent/child
	if logger.NamespaceEnabled("child") {
		t.Errorf("Expected namespace 'child' (parent/child) to be disabled")
	}
	if child.NamespaceEnabled("") {
		t.Errorf("Expected namespace %q to be disabled", child.GetPath())
	}

	buf.Reset()
	child.Infof("Should not log this")
	if buf.String() != "" {
		t.Errorf("Expected empty output, got %q", buf.String())
	}

	// Test NamespaceEnable
	logger.NamespaceEnable("child") // Re-enables parent/child
	if !logger.NamespaceEnabled("child") {
		t.Errorf("Expected namespace 'child' (parent/child) to be enabled")
	}
	if !child.NamespaceEnabled("") {
		t.Errorf("Expected namespace %q to be enabled", child.GetPath())
	}

	buf.Reset()
	child.Infof("Should log this again")
	expectedReEnabledLog := expectedNestedLogPrefix(child, lx.Arrow) + "INFO: Should log this again"
	if !strings.Contains(buf.String(), expectedReEnabledLog) {
		t.Errorf("Expected %q, got %q", expectedReEnabledLog, buf.String())
	}
}

// expectedNestedLogPrefix generates the expected log prefix for nested path style.
func expectedNestedLogPrefix(l *ll.Logger, arrow string) string {
	separator := l.GetSeparator()
	if separator == "" {
		separator = lx.Slash
	}

	if l.GetPath() != "" {
		parts := strings.Split(l.GetPath(), separator)
		var builder strings.Builder
		for i, part := range parts {
			builder.WriteString(lx.LeftBracket)
			builder.WriteString(part)
			builder.WriteString(lx.RightBracket)
			if i < len(parts)-1 {
				builder.WriteString(arrow)
			}
		}
		builder.WriteString(lx.Colon)
		builder.WriteString(lx.Space)
		return builder.String()
	}
	return ""
}

// TestSharedNamespaces verifies that namespace state affects derived loggers.
func TestSharedNamespaces(t *testing.T) {
	buf := &bytes.Buffer{}
	parent := ll.New("parent").Enable().Handler(lh.NewTextHandler(buf))

	// Disable child namespace
	parent.NamespaceDisable("child") // Sets parent/child to false

	// Create child logger
	child := parent.Namespace("child").Handler(lh.NewTextHandler(buf)).Style(lx.FlatPath)

	// Verify disabled state
	if parent.NamespaceEnabled("child") {
		t.Errorf("Expected namespace 'child' (parent/child) to be disabled")
	}
	if child.NamespaceEnabled("") {
		t.Errorf("Expected namespace %q to be disabled", child.GetPath())
	}

	// Test logging (should be blocked)
	buf.Reset()
	child.Infof("Should not log from child")
	if buf.String() != "" {
		t.Errorf("Expected no output from %q, got %q", child.GetPath(), buf.String())
	}

	// Enable child namespace
	parent.NamespaceEnable("child") // Sets parent/child to true
	if !parent.NamespaceEnabled("child") {
		t.Errorf("Expected namespace 'child' (parent/child) to be enabled")
	}
	if !child.NamespaceEnabled("") {
		t.Errorf("Expected namespace %q to be enabled", child.GetPath())
	}

	// Test logging (should appear)
	buf.Reset()
	child.Infof("Should log from child")
	expectedLog := "[parent/child] INFO: Should log from child"
	if !strings.Contains(buf.String(), expectedLog) {
		t.Errorf("Expected %q, got %q", expectedLog, buf.String())
	}
}

// TestNamespaceHierarchicalOverride verifies hierarchical namespace rules with overrides.
func TestNamespaceHierarchicalOverride(t *testing.T) {
	l := ll.New("base").Disable() // Default "/" separator, instance disabled

	// Create buffers and loggers
	bufC1 := &bytes.Buffer{}
	bufC1D2 := &bytes.Buffer{}
	bufC1D2E3 := &bytes.Buffer{}
	bufC1D2E3F4 := &bytes.Buffer{}

	c1 := l.Namespace("c.1").Handler(lh.NewTextHandler(bufC1))                         // base/c.1
	c1d2 := l.Namespace("c.1/d.2").Handler(lh.NewTextHandler(bufC1D2))                 // base/c.1/d.2
	c1d2e3 := l.Namespace("c.1/d.2/e.3").Handler(lh.NewTextHandler(bufC1D2E3))         // base/c.1/d.2/e.3
	c1d2e3f4 := l.Namespace("c.1/d.2/e.3/f.4").Handler(lh.NewTextHandler(bufC1D2E3F4)) // base/c.1/d.2/e.3/f.4

	// Set namespace rules
	l.NamespaceDisable("c.1")        // base/c.1 -> false
	l.NamespaceDisable("c.1/d.2")    // base/c.1/d.2 -> false
	l.NamespaceEnable("c.1/d.2/e.3") // base/c.1/d.2/e.3 -> true

	// Verify namespace states
	if l.NamespaceEnabled("c.1") {
		t.Errorf("Expected namespace 'c.1' (base/c.1) to be disabled")
	}
	if l.NamespaceEnabled("c.1/d.2") {
		t.Errorf("Expected namespace 'c.1/d.2' (base/c.1/d.2) to be disabled")
	}
	if !l.NamespaceEnabled("c.1/d.2/e.3") {
		t.Errorf("Expected namespace 'c.1/d.2/e.3' (base/c.1/d.2/e.3) to be enabled")
	}
	if !l.NamespaceEnabled("c.1/d.2/e.3/f.4") {
		t.Errorf("Expected namespace 'c.1/d.2/e.3/f.4' (base/c.1/d.2/e.3/f.4) to be enabled")
	}
	if !c1d2e3f4.NamespaceEnabled("") {
		t.Errorf("Expected logger (base/c.1/d.2/e.3/f.4) to be enabled")
	}

	// Test logging
	c1.Infof("Log from c1")
	c1d2.Infof("Log from c1d2")
	c1d2e3.Infof("Log from c1d2e3")
	c1d2e3f4.Infof("Log from c1d2e3f4")

	// Verify outputs
	if strings.Contains(bufC1.String(), "Log from c1") {
		t.Errorf("Expected no log from c1 (base/c.1), got %q", bufC1.String())
	}
	if strings.Contains(bufC1D2.String(), "Log from c1d2") {
		t.Errorf("Expected no log from c1d2 (base/c.1/d.2), got %q", bufC1D2.String())
	}
	if !strings.Contains(bufC1D2E3.String(), "Log from c1d2e3") {
		t.Errorf("Expected log from c1d2e3 (base/c.1/d.2/e.3), got %q", bufC1D2E3.String())
	}
	if !strings.Contains(bufC1D2E3F4.String(), "Log from c1d2e3f4") {
		t.Errorf("Expected log from c1d2e3f4 (base/c.1/d.2/e.3/f.4), got %q", bufC1D2E3F4.String())
	}
}
