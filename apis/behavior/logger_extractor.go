package behavior

import (
	"context"

	"github.com/mikros-dev/mikros/apis/features/logger"
)

// LoggerExtractor defines the contract for plugins that enrich log messages
// by extracting contextual information from a service execution context.
//
// This interface is used across all service types supported by mikros.
// Implementations are responsible for retrieving relevant attributes from the
// provided context and returning them as structured log fields.
//
// The extracted attributes will be automatically included in log messages
// generated during request or task processing, enabling enhanced observability
// and traceability across service types.
type LoggerExtractor interface {
	// Extract retrieves logging attributes from the given context.
	// These attributes will be included in structured log messages.
	Extract(ctx context.Context) []logger.Attribute
}
