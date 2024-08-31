package sqldb

import (
	"context"
	"database/sql"
	"sync"
	"time"

	tracer "github.com/BetaLixT/appInsightsTrace"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
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
	WithHook(h Hook)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
}

type Options struct {
	MaxIdleConns   int
	MaxOpenConns   int
	PanicablePings bool
	Name           string
	Hook           Hook
}

type dbImpl struct {
	db       *sql.DB
	pingLock *sync.Mutex
	driver   string
	hook     Hook
}

// New returns a NativeDatabase interface for the given driver and DSN
// WARNING: It will panic if the driver is not supported
func New(Driver string, DSN string) DB {
	return NewWithOptions(Driver, DSN, nil)
}

// New returns a NativeDatabase interface for the given driver and DSN
// WARNING: It will panic if the driver is not supported
func NewWithOptions(Driver string, DSN string, opts *Options) DB {
	var db *sql.DB
	var err error
	switch Driver {
	case "postgres":
		db, err = sql.Open("postgres", DSN)
	case "mysql":
		db, err = sql.Open("mysql", DSN)
	case "sqlite3":
		db, err = sql.Open("sqlite3", DSN)
	case "sqlserver":
		panic("be fucking for real, make better choices")
	default:
		panic("Unsupported driver")
	}
	if err != nil {
		panic(err)
	}
	ndb := &dbImpl{
		db:     db,
		driver: Driver,
	}

	if opts == nil {
		return ndb
	}
	if opts.PanicablePings {
		ndb.pingLock = &sync.Mutex{}
		ndb.Ping()
	}
	if opts.MaxIdleConns > 0 {
		ndb.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.MaxOpenConns > 0 {
		ndb.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.Hook != nil {
		ndb.hook = opts.Hook
	}
	return ndb
}

// TracedNativeDBWrapper returns a DB interface for the given driver and DSN
// WARNING: It will panic if the driver is not supported
// DEPRECATED: Use New Or NewWithOptions instead
func TracedNativeDBWrapper(
	Driver string,
	DSN string,
	t Trx,
	name string,
) DB {
	switch Driver {
	case "postgres":
		if db, err := sql.Open("postgres", DSN); err != nil {
			panic(err)
		} else {
			ndb := &dbImpl{
				db:     db,
				driver: Driver,
			}
			return ndb
		}
	default:
		panic("Unsupported driver")
	}
}

// TracedNativeDBWrapperWithOptions returns a DB interface for the given driver and DSN
// WARNING: It will panic if the driver is not supported
func TracedNativeDBWrapperWithOptions(
	Driver string,
	DSN string,
	t *tracer.AppInsightsCore,
	name string,
	opts *Options,
) DB {
	var ndb *dbImpl
	switch Driver {
	case "postgres":
		if db, err := sql.Open("postgres", DSN); err != nil {
			panic(err)
		} else {
			ndb = &dbImpl{
				db:     db,
				driver: Driver,
			}
			if opts != nil {
				if opts.PanicablePings {
					ndb.pingLock = &sync.Mutex{}
					ndb.Ping()
				}
				ndb.hook = opts.Hook
			}
		}
	default:
		panic("Unsupported driver")
	}
	if opts != nil {
		if opts.MaxIdleConns == 0 {
			ndb.SetMaxIdleConns(2)
		}
		ndb.SetMaxOpenConns(opts.MaxOpenConns)
	}
	return ndb
}

// DBWarpper() returns the underlying DB wrapping the driver
func DBWarpper(
	db *sql.DB,
	t Trx,
	name string,
	logger *zap.Logger,
) DB {
	return &dbImpl{
		db: db,
	}
}

// Begin starts and returns a new transaction.
// params:
//   - none
//
// returns:
//   - *Tx: the transaction
//   - error: any error that occurred
func (d *dbImpl) Begin() (*Tx, error) {
	now := time.Now()

	t, err := d.db.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{t, d.hook, context.Background(), now}, nil
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
	now := time.Now()

	t, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{t, d.hook, ctx, now}, nil
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
	start := time.Now()

	res, err := d.db.Exec(query, args...)

	end := time.Now()

	if d.hook == nil {
		return res, err
	}

	sid, _ := generateParentId()

	d.hook.AfterQuery(
		context.Background(),
		sid,
		"Exec",
		query,
		args,
		start,
		end,
		err,
	)

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
	start := time.Now()

	res, err := d.db.ExecContext(ctx, query, args...)

	end := time.Now()

	if d.hook == nil {
		return res, err
	}

	sid, _ := generateParentId()

	d.hook.AfterQuery(
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

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
// params:
//   - N/A
//
// returns:
//   - error: any error that occurred
func (d *dbImpl) Ping() error {
	if d.pingLock == nil {
		return nil
	}

	d.pingLock.Lock()
	var err error
	go func() {
		for {
			time.Sleep(time.Second * 5)
			err = d.db.Ping()
			if err != nil {
				panic(err)
			}
		}
	}()
	return nil
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
	start := time.Now()

	res, err := d.db.Query(query, args...)

	end := time.Now()

	if d.hook == nil {
		return res, err
	}

	sid, _ := generateParentId()

	d.hook.AfterQuery(
		context.Background(),
		sid,
		"Query",
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
//   - ctx: the context for the query
//   - query: the query to execute
//   - args: the arguments for the query
//
// returns:
//   - *sql.Rows: the rows returned by the query
//   - error: any error that occurred
func (d *dbImpl) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()

	res, err := d.db.Query(query, args...)

	end := time.Now()

	if d.hook == nil {
		return res, err
	}

	sid, _ := generateParentId()

	d.hook.AfterQuery(
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
	start := time.Now()

	res := d.db.QueryRow(query, args...)

	end := time.Now()

	if d.hook == nil {
		return res
	}

	sid, _ := generateParentId()

	d.hook.AfterQuery(
		context.Background(),
		sid,
		"QueryRow",
		query,
		args,
		start,
		end,
		res.Err(),
	)

	return res
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
	start := time.Now()

	res := d.db.QueryRow(query, args...)

	end := time.Now()

	if d.hook == nil {
		return res
	}

	sid, _ := generateParentId()

	d.hook.AfterQuery(
		ctx,
		sid,
		"QueryRow",
		query,
		args,
		start,
		end,
		res.Err(),
	)

	return res
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

// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
// If n <= 0, no idle connections are retained.
// params:
//   - n: the number of connections to set
func (i *dbImpl) SetMaxIdleConns(n int) {
	i.db.SetMaxIdleConns(n)
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
// If n <= 0, then there is no limit on the number of open connections.
// The default is 0 (unlimited).
// params:
//   - n: the number of connections to set
func (i *dbImpl) SetMaxOpenConns(n int) {
	i.db.SetMaxOpenConns(n)
}

// WithHook sets the hook for the database
// params:
//   - h: the hook to set
func (i *dbImpl) WithHook(h Hook) {
	i.hook = h
}
