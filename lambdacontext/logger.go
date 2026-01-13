//go:build go1.21
// +build go1.21

// Copyright 2026 Amazon.com, Inc. or its affiliates. All Rights Reserved.

package lambdacontext

import (
	"context"
	"log/slog"
	"os"
)

// Field represents an optional field to include in log records.
type Field struct {
	key   string
	value func(*LambdaContext) string
}

// FunctionArn includes the invoked function ARN in log records.
var FunctionArn = Field{"functionArn", func(lc *LambdaContext) string { return lc.InvokedFunctionArn }} //nolint: staticcheck

// TenantId includes the tenant ID in log records (for multi-tenant functions).
var TenantId = Field{"tenantId", func(lc *LambdaContext) string { return lc.TenantID }} //nolint: staticcheck

// Handler returns a [slog.Handler] for AWS Lambda structured logging.
// It reads AWS_LAMBDA_LOG_FORMAT and AWS_LAMBDA_LOG_LEVEL from environment,
// and injects requestId from Lambda context into each log record.
//
// By default, only requestId is injected. Pass optional fields to include more:
//
//	// Default: only requestId
//	slog.SetDefault(slog.New(lambdacontext.Handler()))
//
//	// With functionArn and tenantId
//	slog.SetDefault(slog.New(lambdacontext.Handler(lambdacontext.FunctionArn, lambdacontext.TenantId)))
func Handler(fields ...Field) slog.Handler {
	level := parseLogLevel()
	opts := &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: ReplaceAttr,
	}

	var h slog.Handler
	if LogFormatName == "JSON" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	return &lambdaHandler{handler: h, fields: fields}
}

// ReplaceAttr maps slog's default keys to AWS Lambda's log format (time->timestamp, msg->message).
func ReplaceAttr(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return attr
	}

	switch attr.Key {
	case slog.TimeKey:
		attr.Key = "timestamp"
	case slog.MessageKey:
		attr.Key = "message"
	}
	return attr
}

// Attrs returns Lambda context fields as slog-compatible key-value pairs.
// For most use cases, using [Handler] with slog.InfoContext is preferred.
func (lc *LambdaContext) Attrs() []any {
	return []any{"requestId", lc.AwsRequestID}
}

// lambdaHandler wraps a slog.Handler to inject Lambda context fields.
type lambdaHandler struct {
	handler slog.Handler
	fields  []Field
}

// Enabled implements slog.Handler.
func (h *lambdaHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (h *lambdaHandler) Handle(ctx context.Context, r slog.Record) error {
	if lc, ok := FromContext(ctx); ok {
		r.AddAttrs(slog.String("requestId", lc.AwsRequestID))

		for _, f := range h.fields {
			if v := f.value(lc); v != "" {
				r.AddAttrs(slog.String(f.key, v))
			}
		}
	}
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements slog.Handler.
func (h *lambdaHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &lambdaHandler{
		handler: h.handler.WithAttrs(attrs),
		fields:  h.fields,
	}
}

// WithGroup implements slog.Handler.
func (h *lambdaHandler) WithGroup(name string) slog.Handler {
	return &lambdaHandler{
		handler: h.handler.WithGroup(name),
		fields:  h.fields,
	}
}

func parseLogLevel() slog.Level {
	switch LogLevelName {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
