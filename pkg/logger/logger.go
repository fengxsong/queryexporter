package logger

import (
	"context"

	"github.com/go-kit/log"
)

var loggerKey = struct{}{}

func InjectContext(ctx context.Context, logger log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) log.Logger {
	v := ctx.Value(loggerKey)
	if v == nil {
		return log.NewNopLogger()
	}
	return v.(log.Logger)
}
