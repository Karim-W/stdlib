package stdlib

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	tracer "github.com/BetaLixT/appInsightsTrace"
	"go.uber.org/zap"
)

type TracedNativeTx struct {
	*sql.Tx
	t      *tracer.AppInsightsCore
	logger *zap.Logger
	driver string
	name   string
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
func (t *TracedNativeTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	now := time.Now()
	res, err := t.Tx.QueryContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	if err != nil {
		t.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
		t.t.TraceException(ctx, err, 0, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
			"error": err.Error(),
		})
		t.t.TraceDependency(ctx, "", t.driver, t.name, "Query", false, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
			"error": err.Error(),
		})
	} else {
		t.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
		t.t.TraceDependency(ctx, "", t.driver, t.name, "Query", true, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
		})
	}
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
func (t *TracedNativeTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	now := time.Now()
	res, err := t.Tx.Query(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	if err != nil {
		t.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
	} else {
		t.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
	}
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
func (t *TracedNativeTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	now := time.Now()
	res, err := t.Tx.ExecContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	if err != nil {
		t.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
		t.t.TraceException(ctx, err, 0, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
			"error": err.Error(),
		})
		t.t.TraceDependency(ctx, "", t.driver, t.name, "Exec", false, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
			"error": err.Error(),
		})
	} else {
		t.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
		t.t.TraceDependency(ctx, "", t.driver, t.name, "Exec", true, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
		})
	}
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
func (t *TracedNativeTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	now := time.Now()
	res, err := t.Tx.Exec(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	if err != nil {
		t.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
	} else {
		t.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
	}
	return res, err
}

// Commit commits the transaction.
// params:
//   - none
//
// returns:
//   - error: error if any
func (t *TracedNativeTx) Commit() error {
	now := time.Now()
	err := t.Tx.Commit()
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	if err != nil {
		t.logger.Error("[DATABASE]  Error committing transaction",
			zap.Error(err))
	} else {
		t.logger.Info("[DATABASE]  Committing transaction",
			zap.Float64("elapsed(ms)", elapsed))
	}
	return err
}

// Rollback aborts the transaction.
// params:
//   - none
//
// returns:
//   - error: error if any
func (t *TracedNativeTx) Rollback() error {
	now := time.Now()
	err := t.Tx.Rollback()
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	if err != nil {
		t.logger.Error("[DATABASE]  Error rolling back transaction",
			zap.Error(err))
	} else {
		t.logger.Info("[DATABASE]  Rolling back transaction",
			zap.Float64("elapsed(ms)", elapsed))
	}
	return err
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// params:
//   - query: query to execute
//   - args: arguments to pass to query
//
// returns:
//   - *sql.Row: row returned by query
func (t *TracedNativeTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	now := time.Now()
	res := t.Tx.QueryRowContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	t.logger.Info("[DATABASE]  Executing query",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Float64("elapsed(ms)", elapsed))
	t.t.TraceDependency(ctx, "", t.driver, t.name, "QueryRow", true, now, after, map[string]string{
		"query": query,
		"args":  fmt.Sprintf("%v", args),
	})
	return res
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// params:
//   - query: query to execute
//   - args: arguments to pass to query
//
// returns:
//   - *sql.Row: row returned by query
func (t *TracedNativeTx) QueryRow(query string, args ...interface{}) *sql.Row {
	now := time.Now()
	res := t.Tx.QueryRow(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	t.logger.Info("[DATABASE]  Executing query",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Float64("elapsed(ms)", elapsed))
	return res
}
