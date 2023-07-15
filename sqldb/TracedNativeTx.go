package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Tx struct {
	*sql.Tx
	t    Trx
	name string
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// params:
//   - ctx: context
//   - query: query to execute
//   - args: arguments to pass to query
//
// returns:
//   - *sql.Rows: rows returned by query
//   - error: error if any
func (t *Tx) QueryContext(
	ctx context.Context,
	query string,
	args ...interface{},
) (*sql.Rows, error) {
	now := time.Now()
	res, err := t.Tx.QueryContext(ctx, query, args...)
	after := time.Now()
	if t.t == nil {
		return res, err
	}
	fields := map[string]string{
		"query": query,
	}
	if err != nil {
		fields["error"] = err.Error()
		fields["args"] = fmt.Sprintf("%v", args)
		t.t.TraceException(ctx, err, 0, fields)
	}
	sid, err := generateParentId()
	if err != nil {
		sid = "0000"
	}
	t.t.TraceDependency(
		ctx,
		sid,
		"sql",
		t.name,
		"Query "+query,
		err == nil,
		now,
		after,
		fields,
	)
	return res, err
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// params:
//   - query: query to execute
//   - args: arguments to pass to query
//
// returns:
//   - sql.Result: result of query
//   - error: error if any
func (t *Tx) ExecContext(
	ctx context.Context,
	query string,
	args ...interface{},
) (sql.Result, error) {
	now := time.Now()
	res, err := t.Tx.ExecContext(ctx, query, args...)
	after := time.Now()
	if t.t == nil {
		return res, err
	}
	fields := map[string]string{
		"query": query,
	}
	if err != nil {
		fields["error"] = err.Error()
		fields["args"] = fmt.Sprintf("%v", args)
		t.t.TraceException(ctx, err, 0, fields)
	}
	sid, err := generateParentId()
	if err != nil {
		sid = "0000"
	}
	t.t.TraceDependency(ctx, sid, "sql", t.name, "Exec "+query, err == nil, now, after, fields)
	return res, err
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// params:
//   - query: query to execute
//   - args: arguments to pass to query
//
// returns:
//   - *sql.Row: row returned by query
func (t *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	now := time.Now()
	res := t.Tx.QueryRowContext(ctx, query, args...)
	after := time.Now()
	if t.t == nil {
		return res
	}
	fields := map[string]string{
		"query": query,
	}
	sid, err := generateParentId()
	if err != nil {
		sid = "0000"
	}
	t.t.TraceDependency(
		ctx,
		sid,
		"sql",
		t.name,
		"QueryRow "+query,
		true,
		now,
		after,
		fields,
	)
	return res
}
