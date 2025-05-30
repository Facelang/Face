package tests

import (
	"bytes"
	"encoding/json"
	"github.com/olekukonko/ll"
	"github.com/olekukonko/ll/lh"
	"github.com/olekukonko/ll/lx"
	"log/slog"
	"strings"
	"testing"
)

// TestHandlers verifies the behavior of all log handlers (Text, Colorized, JSON, Slog, Multi).
func TestHandlers(t *testing.T) {
	// Test TextHandler for plain text output
	t.Run("TextHandler", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := ll.New("test").Enable().Handler(lh.NewTextHandler(buf))
		logger.Fields("key", "value").Infof("Test text")
		if !strings.Contains(buf.String(), "[test] INFO: Test text [key=value]") {
			t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test text [key=value]")
		}
	})

	// Test ColorizedHandler for ANSI-colored output
	t.Run("ColorizedHandler", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := ll.New("test").Enable().Handler(lh.NewColorizedHandler(buf))
		logger.Fields("key", "value").Infof("Test color")
		// Check for namespace presence, ignoring ANSI codes
		if !strings.Contains(buf.String(), "[test]") {
			t.Errorf("Expected %q to contain %q", buf.String(), "[test] INFO: Test color [key=value]")
		}
	})

	// Test JSONHandler for structured JSON output
	t.Run("JSONHandler", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := ll.New("test").Enable().Handler(lh.NewJSONHandler(buf))
		logger.Fields("key", "value").Infof("Test JSON")
		// Parse JSON output and verify fields
		var data lh.JsonOutput
		if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if data.Level != lx.LevelInfo.String() {
			t.Errorf("Expected level=%q, got %q", "INFO", data.Level)
		}
		if data.Msg != "Test JSON" {
			t.Errorf("Expected message=%q, got %q", "Test JSON", data.Msg)
		}
		if data.Namespace != "test" {
			t.Errorf("Expected namespace=%q, got %q", "test", data.Namespace)
		}
		f := data.Fields
		if f["key"] != "value" {
			t.Errorf("Expected key=%q, got %q", "value", f["key"])
		}
	})

	// Test SlogHandler for compatibility with slog
	t.Run("SlogHandler", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := ll.New("test").Enable().Handler(lh.NewSlogHandler(slog.NewTextHandler(buf, nil)))
		logger.Fields("key", "value").Infof("Test slog")
		output := buf.String()
		if !strings.Contains(output, "level=INFO") {
			t.Errorf("Expected %q to contain %q", output, "level=INFO")
		}
		if !strings.Contains(output, "msg=\"Test slog\"") {
			t.Errorf("Expected %q to contain %q", output, "msg=\"Test slog\"")
		}
		if !strings.Contains(output, "namespace=test") {
			t.Errorf("Expected %q to contain %q", output, "namespace=test")
		}
		if !strings.Contains(output, "key=value") {
			t.Errorf("Expected %q to contain %q", output, "key=value")
		}
	})

	// Test MultiHandler for combining multiple handlers
	t.Run("MultiHandler", func(t *testing.T) {
		buf1 := &bytes.Buffer{}
		buf2 := &bytes.Buffer{}
		logger := ll.New("test").Enable().Handler(lh.NewMultiHandler(
			lh.NewTextHandler(buf1),
			lh.NewJSONHandler(buf2),
		))
		logger.Fields("key", "value").Infof("Test multi")
		// Verify TextHandler output
		if !strings.Contains(buf1.String(), "[test] INFO: Test multi [key=value]") {
			t.Errorf("Expected %q to contain %q", buf1.String(), "[test] INFO: Test multi [key=value]")
		}
		// Verify JSONHandler output
		var data map[string]interface{}
		if err := json.Unmarshal(buf2.Bytes(), &data); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if data["msg"] != "Test multi" {
			t.Errorf("Expected message=%q, got %q", "Test multi", data["msg"])
		}
	})
}
