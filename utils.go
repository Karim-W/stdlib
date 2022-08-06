package stdlib

import (
	"context"

	"go.uber.org/zap"
)

func extractContextInfo(ctx context.Context) (bool, string, string) {
	return false, "", ""
}

func extractLoggerFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	l := *logger
	if ok, tid, pid := extractContextInfo(ctx); ok {
		l.With(zap.String("transactionId", tid), zap.String("parentId", pid))
		return &l
	} else {
		return logger
	}
}
