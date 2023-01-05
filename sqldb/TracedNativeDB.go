package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	tracer "github.com/BetaLixT/appInsightsTrace"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DB interface {
	Begin() (*Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error)
	Close() error
	Conn(ctx context.Context) (*sql.Conn, error)
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	Stats() sql.DBStats
	WithLogger(l *zap.Logger) DB
}

type dbImpl struct {
	logger   *zap.Logger
	db       *sql.DB
	pingLock sync.Mutex
	t        *tracer.AppInsightsCore
	driver   string
	name     string
}

// DBProvider returns a NativeDatabase interface for the given driver and DSN
// WARNING: It will panic if the driver is not supported
func DBProvider(Driver string, DSN string) DB {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	switch Driver {
	case "postgres":
		if db, err := sql.Open("postgres", DSN); err != nil {
			panic(err)
		} else {
			l.Info("[DATABASE]\tSucessfuly Connected to postgres database")
			ndb := &dbImpl{
				logger:   l,
				db:       db,
				pingLock: sync.Mutex{},
				driver:   Driver,
			}
			ndb.Ping()
			return ndb
		}
	default:
		panic("Unsupported driver")
	}
}

// TracedNativeDBWrapper returns a DB interface for the given driver and DSN
// WARNING: It will panic if the driver is not supported
func TracedNativeDBWrapper(
	Driver string,
	DSN string,
	t *tracer.AppInsightsCore,
	name string,
) DB {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	switch Driver {
	case "postgres":
		if db, err := sql.Open("postgres", DSN); err != nil {
			panic(err)
		} else {
			l.Info("[DATABASE]\tSucessfuly Connected to postgres database")
			ndb := &dbImpl{
				logger:   l,
				db:       db,
				pingLock: sync.Mutex{},
				t:        t,
				driver:   Driver,
				name:     name,
			}
			ndb.Ping()
			return ndb
		}
	default:
		panic("Unsupported driver")
	}
}

// WithLogger returns a Copy of the DB with the given logger
// shallow copy
func (d *dbImpl) WithLogger(l *zap.Logger) DB {
	newD := &dbImpl{}
	newD.logger = l
	newD.driver = d.driver
	if d.db != nil {
		newD.db = d.db
	}
	if d.t != nil {
		newD.t = d.t
	}
	if d.name != "" {
		newD.name = d.name
	}
	return newD
}

// Begin starts and returns a new transaction.
// params:
//   - none
//
// returns:
//   - *Tx: the transaction
//   - error: any error that occurred
func (d *dbImpl) Begin() (*Tx, error) {
	t, err := d.db.Begin()
	if err != nil {
		d.logger.Error("[DATABASE]  Error starting transaction",
			zap.Error(err))
		return nil, err
	}
	return &Tx{t, d.t, d.logger, d.driver, d.name}, nil
}

// BeginTx starts and returns a new transaction.
// params:
//   - ctx: the context for the transaction
//   - opts: the options for the transaction
//
// returns:
//   - *Tx: the transaction
//   - error: any error that occurred
func (d *dbImpl) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	t, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		d.logger.Error("[DATABASE]  Error starting transaction",
			zap.Error(err))
		return nil, err
	}
	tx := &Tx{t, d.t, d.logger, d.driver, d.name}
	return tx, nil
}

// Close closes the database, releasing any open resources.
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
func (d *dbImpl) Close() error {
	return d.db.Close()
}

// Conn returns a single-use connection to the database.
// The connection is automatically returned to the idle connection pool
func (d *dbImpl) Conn(ctx context.Context) (*sql.Conn, error) {
	return d.db.Conn(ctx)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// params:
//   - query: the query to execute
//   - args: the arguments for the query
//
// returns:
//   - sql.Result: the result of the query
//   - error: any error that occurred
func (d *dbImpl) Exec(query string, args ...any) (sql.Result, error) {
	now := time.Now()
	res, err := d.db.Exec(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	if err != nil {
		d.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
	} else {
		d.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
	}
	return res, err
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// params:
//   - ctx: the context for the query
//   - query: the query to execute
//   - args: the arguments for the query
//
// returns:
//   - sql.Result: the result of the query
//   - error: any error that occurred
func (d *dbImpl) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	now := time.Now()
	res, err := d.db.ExecContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	if err != nil {
		d.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
		if d.t != nil {
			d.t.TraceException(ctx, err, 0, map[string]string{
				"query": query,
				"args":  fmt.Sprintf("%v", args),
				"error": err.Error(),
			})
			d.t.TraceDependency(ctx, "", d.driver, d.name, "EXEC", false, now, after, map[string]string{
				"query": query,
				"args":  fmt.Sprintf("%v", args),
				"error": err.Error(),
			})
		}
	} else {
		d.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
		if d.t != nil {
			d.t.TraceDependency(ctx, "", d.driver, d.name, "EXEC", true, now, after, map[string]string{
				"query": query,
				"args":  fmt.Sprintf("%v", args),
			})
		}
	}
	return res, err
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
// params:
//   - N/A
//
// returns:
//   - error: any error that occurred
func (d *dbImpl) Ping() error {
	d.pingLock.Lock()
	go func() {
		for {
			time.Sleep(time.Second * 5)
			d.db.Ping()
		}
	}()
	return d.db.Ping()
}

// PingContext verifies a connection to the database is still alive,
// establishing a connection if necessary.
// params:
//   - ctx: the context for the ping
//
// returns:
//   - error: any error that occurred
func (d *dbImpl) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the
// returned statement.
// params:
//   - query: the query to prepare
//
// returns:
//   - *sql.Stmt: the prepared statement
//   - error: any error that occurred
func (d *dbImpl) Prepare(query string) (*sql.Stmt, error) {
	return d.db.Prepare(query)
}

// PrepareContext creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the
// returned statement.
// params:
//   - ctx: the context for the prepare
//   - query: the query to prepare
//
// returns:
//   - *sql.Stmt: the prepared statement
//   - error: any error that occurred
func (d *dbImpl) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return d.db.PrepareContext(ctx, query)
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// params:
//   - query: the query to execute
//   - args: the arguments for the query
//
// returns:
//   - *sql.Rows: the rows returned by the query
//   - error: any error that occurred
func (d *dbImpl) Query(query string, args ...any) (*sql.Rows, error) {
	now := time.Now()
	res, err := d.db.Query(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	if err != nil {
		d.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
	} else {
		d.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
	}
	return res, err
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// params:
//   - ctx: the context for the query
//   - query: the query to execute
//   - args: the arguments for the query
//
// returns:
//   - *sql.Rows: the rows returned by the query
//   - error: any error that occurred
func (d *dbImpl) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	now := time.Now()
	res, err := d.db.QueryContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	if err != nil {
		d.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
		if d.t != nil {
			d.t.TraceException(ctx, err, 0, map[string]string{
				"query": query,
				"args":  fmt.Sprintf("%v", args),
				"error": err.Error(),
			})
			d.t.TraceDependency(ctx, "", d.driver, d.name, "Query", false, now, after, map[string]string{
				"query": query,
				"args":  fmt.Sprintf("%v", args),
				"error": err.Error(),
			})
		}
	} else {
		d.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
		if d.t != nil {
			d.t.TraceDependency(ctx, "", d.driver, d.name, "Query", true, now, after, map[string]string{
				"query": query,
				"args":  fmt.Sprintf("%v", args),
			})
		}
	}
	return res, err
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// params:
//   - query: the query to execute
//   - args: the arguments for the query
//
// returns:
//   - *sql.Row: the row returned by the query
func (d *dbImpl) QueryRow(query string, args ...any) *sql.Row {
	now := time.Now()
	r := d.db.QueryRow(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	d.logger.Info("[DATABASE]  Executing query",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Float64("elapsed(ms)", elapsed))
	return r
}

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// params:
//   - ctx: the context for the query
//   - query: the query to execute
//   - args: the arguments for the query
//
// returns:
//   - *sql.Row: the row returned by the query
func (d *dbImpl) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	now := time.Now()
	r := d.db.QueryRowContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	d.logger.Info("[DATABASE]  Executing query",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Float64("elapsed(ms)", elapsed))
	if d.t != nil {
		d.t.TraceDependency(ctx, "", d.driver, d.name, "QueryRow"+query, true, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
		})
	}
	return r
}

// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
// Expired connections may be closed lazily before reuse.
// If d <= 0, connections are not closed due to a connection's idle time.
// params:
//   - d: the duration to set
func (i *dbImpl) SetConnMaxIdleTime(d time.Duration) {
	i.db.SetConnMaxIdleTime(d)
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
// Expired connections may be closed lazily before reuse.
// If d <= 0, connections are not closed due to a connection's lifetime.
// params:
//   - d: the duration to set
func (i *dbImpl) SetConnMaxLifetime(d time.Duration) {
	i.db.SetConnMaxLifetime(d)
}

// Stats returns database statistics.
// params:
//   - none
//
// returns:
//   - sql.DBStats: the database statistics
func (i *dbImpl) Stats() sql.DBStats {
	return i.db.Stats()
}
