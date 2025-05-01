package behavior

import (
	"context"

	"github.com/mikros-dev/mikros/apis/features/logger"
)

// LoggerExtractor is an interface that a plugin can implement to provide an API
// allowing the service extract content from its context to add them into
// log messages.
type LoggerExtractor interface {
	Extract(ctx context.Context) []logger.Attribute
}
