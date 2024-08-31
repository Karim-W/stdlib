package sqldb

import (
	"context"
	"database/sql"
	"time"
)

type Tx struct {
	*sql.Tx
	hook  Hook
	ctx   context.Context
	begin time.Time
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
	start := time.Now()

	res, err := t.Tx.Query(query, args...)

	end := time.Now()

	if t.hook == nil {
		return res, err
	}

	sid, _ := generateParentId()

	t.hook.AfterQuery(
		ctx,
		sid,
		"QueryContext",
		query,
		args,
		start,
		end,
		err,
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
	start := time.Now()

	res, err := t.Tx.ExecContext(ctx, query, args...)

	end := time.Now()

	if t.hook == nil {
		return res, err
	}

	sid, _ := generateParentId()

	t.hook.AfterQuery(
		ctx,
		sid,
		"ExecContext",
		query,
		args,
		start,
		end,
		err,
	)

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
	start := time.Now()

	res := t.Tx.QueryRowContext(ctx, query, args...)

	end := time.Now()

	if t.hook == nil {
		return res
	}

	sid, _ := generateParentId()

	t.hook.AfterQuery(
		ctx,
		sid,
		"QueryRowContext",
		query,
		args,
		start,
		end,
		nil,
	)

	return res
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	start := time.Now()

	err := t.Tx.Commit()

	end := time.Now()

	if t.hook == nil {
		return err
	}

	sid, _ := generateParentId()

	t.hook.AfterBegin(
		t.ctx,
		sid,
		t.begin,
		end,
		err,
	)

	t.hook.AfterCommit(
		t.ctx,
		sid,
		start,
		time.Now(),
		err,
	)

	return err
}

// Rollback aborts the transaction.
func (t *Tx) Rollback() error {
	start := time.Now()

	err := t.Tx.Rollback()

	end := time.Now()

	if t.hook == nil {
		return err
	}

	sid, _ := generateParentId()

	t.hook.AfterBegin(
		t.ctx,
		sid,
		t.begin,
		end,
		err,
	)

	t.hook.AfterRollback(
		t.ctx,
		start,
		time.Now(),
		err,
	)

	return err
}
