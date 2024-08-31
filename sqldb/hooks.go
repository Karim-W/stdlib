package sqldb

import (
	"context"
	"time"
)

type Hook interface {
	AfterQuery(
		context context.Context,
		sid string,
		method string,
		query string,
		args []interface{},
		start_time time.Time,
		end_time time.Time,
		err error,
	)
	AfterBegin(
		context context.Context,
		sid string,
		start_time,
		end_time time.Time,
		err error,
	)
	AfterCommit(
		context context.Context,
		sid string,
		start_time,
		end_time time.Time,
		err error,
	)
	AfterRollback(
		context context.Context,
		start_time, end_time time.Time,
		err error,
	)
}
