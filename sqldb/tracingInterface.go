package sqldb

import (
	"context"
	"time"
)

type Trx interface {
	TraceDependency(
		ctx context.Context,
		spanId string,
		dependencyType string,
		serviceName string,
		commandName string,
		success bool,
		startTimestamp time.Time,
		eventTimestamp time.Time,
		fields map[string]string,
	)
	TraceException(
		ctx context.Context,
		err interface{},
		skip int,
		fields map[string]string,
	)
}
