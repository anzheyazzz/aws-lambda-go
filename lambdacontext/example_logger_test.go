//go:build go1.21
// +build go1.21

package lambdacontext_test

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// ExampleLogHandler demonstrates basic usage of LogHandler for structured logging.
// The handler automatically injects requestId from Lambda context into each log record.
func ExampleLogHandler() {
	// Set up the Lambda-aware slog handler
	slog.SetDefault(slog.New(lambdacontext.LogHandler()))

	lambda.Start(func(ctx context.Context) (string, error) {
		// Use slog.InfoContext to include Lambda context in logs
		slog.InfoContext(ctx, "processing request", "action", "example")
		return "success", nil
	})
}

// ExampleLogHandler_withFields demonstrates LogHandler with additional fields.
// Use WithFields with FieldFunctionARN() and FieldTenantID() to include extra context.
func ExampleLogHandler_withFields() {
	// Set up handler with function ARN and tenant ID fields
	slog.SetDefault(slog.New(lambdacontext.LogHandler(
		lambdacontext.WithFields(lambdacontext.FieldFunctionARN(), lambdacontext.FieldTenantID()),
	)))

	lambda.Start(func(ctx context.Context) (string, error) {
		slog.InfoContext(ctx, "multi-tenant request", "tenant", "acme-corp")
		return "success", nil
	})
}

// ExampleWithFields demonstrates using WithFields to include specific Lambda context fields.
func ExampleWithFields() {
	// Include only function ARN
	handler := lambdacontext.LogHandler(
		lambdacontext.WithFields(lambdacontext.FieldFunctionARN()),
	)
	slog.SetDefault(slog.New(handler))

	lambda.Start(func(ctx context.Context) (string, error) {
		// Log output will include "functionArn" field
		slog.InfoContext(ctx, "function invoked")
		return "success", nil
	})
}
