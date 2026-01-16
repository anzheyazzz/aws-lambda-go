// Copyright 2026 Amazon.com, Inc. or its affiliates. All Rights Reserved

package lambda

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type logLevel string

const (
	logLevelInfo  logLevel = "INFO"
	logLevelWarn  logLevel = "WARN"
	logLevelError logLevel = "ERROR"
)

var useJSONFormat = os.Getenv("AWS_LAMBDA_LOG_FORMAT") == "JSON"

type logEntry struct {
	Timestamp string          `json:"timestamp"`
	Level     logLevel        `json:"level"`
	Message   json.RawMessage `json:"message"`
}

func logMessage(level logLevel, msg string) {
	if useJSONFormat {
		// Check if msg is already valid JSON
		var rawMsg json.RawMessage
		if json.Valid([]byte(msg)) {
			rawMsg = json.RawMessage(msg)
		} else {
			// Wrap plain text as JSON string
			rawMsg, _ = json.Marshal(msg)
		}

		entry := logEntry{
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Level:     level,
			Message:   rawMsg,
		}
		jsonBytes, _ := json.Marshal(entry)
		fmt.Fprintln(os.Stderr, string(jsonBytes))
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", level, msg)
	}
}

func logInfo(msg string) {
	logMessage(logLevelInfo, msg)
}

func logWarn(msg string) {
	logMessage(logLevelWarn, msg)
}

func logError(msg string) {
	logMessage(logLevelError, msg)
}

func logFatal(msg string) {
	logError(msg)
	os.Exit(1)
}
