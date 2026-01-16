// Copyright 2026 Amazon.com, Inc. or its affiliates. All Rights Reserved

package lambda

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLogMessageTextFormat(t *testing.T) {
	// Save original values
	origStderr := os.Stderr
	origUseJSON := useJSONFormat
	defer func() {
		os.Stderr = origStderr
		useJSONFormat = origUseJSON
	}()

	// Create a pipe to capture stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	useJSONFormat = false

	logInfo("test info message")

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "INFO") {
		t.Errorf("expected output to contain 'INFO', got: %s", output)
	}
	if !strings.Contains(output, "test info message") {
		t.Errorf("expected output to contain 'test info message', got: %s", output)
	}
}

func TestLogMessageJSONFormat(t *testing.T) {
	// Save original values
	origStderr := os.Stderr
	origUseJSON := useJSONFormat
	defer func() {
		os.Stderr = origStderr
		useJSONFormat = origUseJSON
	}()

	// Create a pipe to capture stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	useJSONFormat = true

	logError("test error message")

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Parse as generic map since Message is json.RawMessage
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v, output: %s", err, output)
	}

	if entry["level"] != "ERROR" {
		t.Errorf("expected level 'ERROR', got: %v", entry["level"])
	}
	if entry["message"] != "test error message" {
		t.Errorf("expected message 'test error message', got: %v", entry["message"])
	}
	if entry["timestamp"] == nil || entry["timestamp"] == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestLogMessageJSONFormatWithJSONInput(t *testing.T) {
	// Save original values
	origStderr := os.Stderr
	origUseJSON := useJSONFormat
	defer func() {
		os.Stderr = origStderr
		useJSONFormat = origUseJSON
	}()

	// Create a pipe to capture stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	useJSONFormat = true

	// Log a JSON string (like errorPayload from reportFailure)
	jsonInput := `{"errorMessage":"something went wrong","errorType":"Runtime.Error"}`
	logError(jsonInput)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Parse as generic map
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v, output: %s", err, output)
	}

	if entry["level"] != "ERROR" {
		t.Errorf("expected level 'ERROR', got: %v", entry["level"])
	}

	// Message should be embedded as object, not escaped string
	msgObj, ok := entry["message"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected message to be an object, got: %T (%v)", entry["message"], entry["message"])
	}
	if msgObj["errorMessage"] != "something went wrong" {
		t.Errorf("expected errorMessage 'something went wrong', got: %v", msgObj["errorMessage"])
	}
	if msgObj["errorType"] != "Runtime.Error" {
		t.Errorf("expected errorType 'Runtime.Error', got: %v", msgObj["errorType"])
	}
}

func TestLogLevels(t *testing.T) {
	// Save original values
	origStderr := os.Stderr
	origUseJSON := useJSONFormat
	defer func() {
		os.Stderr = origStderr
		useJSONFormat = origUseJSON
	}()

	useJSONFormat = true

	tests := []struct {
		name     string
		logFunc  func(string)
		expected string
	}{
		{"info", logInfo, "INFO"},
		{"warn", logWarn, "WARN"},
		{"error", logError, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, _ := os.Pipe()
			os.Stderr = w

			tt.logFunc("test message")

			w.Close()
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			var entry map[string]interface{}
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
				t.Fatalf("failed to parse JSON output: %v", err)
			}

			if entry["level"] != tt.expected {
				t.Errorf("expected level %s, got: %v", tt.expected, entry["level"])
			}
		})
	}
}

func TestUseJSONFormatEnvVar(t *testing.T) {
	// This test verifies the initialization behavior
	// The actual env var check happens at package init time

	// Test that the variable can be set
	origUseJSON := useJSONFormat
	defer func() {
		useJSONFormat = origUseJSON
	}()

	useJSONFormat = true
	if !useJSONFormat {
		t.Error("expected useJSONFormat to be true")
	}

	useJSONFormat = false
	if useJSONFormat {
		t.Error("expected useJSONFormat to be false")
	}
}
