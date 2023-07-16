package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	tracer "github.com/BetaLixT/appInsightsTrace"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/libsql/libsql-client-go/libsql"
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
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	WithTrx(t Trx) DB
}

type Options struct {
	MaxIdleConns   int
	MaxOpenConns   int
	PanicablePings bool
	Name           string
}

type dbImpl struct {
	db       *sql.DB
	pingLock sync.Mutex
	t        Trx
	driver   string
	name     string
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
	case "libsql":
		db, err = sql.Open("libsql", DSN)
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
		ndb.pingLock = sync.Mutex{}
		ndb.Ping()
	}
	if opts.MaxIdleConns > 0 {
		ndb.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.MaxOpenConns > 0 {
		ndb.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.Name != "" {
		ndb.name = opts.Name
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
				db:     db,
				t:      t,
				driver: Driver,
				name:   name,
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
				t:      t,
				driver: Driver,
				name:   name,
			}
			if opts != nil {
				if opts.PanicablePings {
					ndb.pingLock = sync.Mutex{}
					ndb.Ping()
				}
			}
		}
	default:
		panic("Unsupported driver")
	}
	if ndb == nil {
		panic("db is nil")
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
		db:   db,
		t:    t,
		name: name,
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
	t, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{t, d.t, d.name}, nil
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
		return nil, err
	}
	tx := &Tx{t, d.t, d.name}
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
	res, err := d.db.Exec(query, args...)
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
	if d.t == nil {
		return res, err
	}
	fields := map[string]string{
		"query": query,
	}
	if err != nil {
		fields["error"] = err.Error()
		fields["args"] = fmt.Sprintf("%v", args)
		d.t.TraceException(ctx, err, 0, fields)
	}
	sid, err := generateParentId()
	if err != nil {
		sid = "0000"
	}
	d.t.TraceDependency(ctx, sid, "sql", d.name, "EXEC "+query, err == nil, now, after, fields)
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
	res, err := d.db.Query(query, args...)
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
	if d.t == nil {
		return res, err
	}
	fields := map[string]string{
		"query": query,
	}
	if err != nil {
		fields["error"] = err.Error()
		fields["args"] = fmt.Sprintf("%v", args)
		d.t.TraceException(ctx, err, 0, fields)
	}
	sid, err := generateParentId()
	if err != nil {
		sid = "0000"
	}
	d.t.TraceDependency(ctx, sid, "sql", d.name, "Query "+query, err == nil, now, after, fields)
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
	r := d.db.QueryRow(query, args...)
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
	if d.t == nil {
		return r
	}
	fields := map[string]string{
		"query": query,
	}
	sid, err := generateParentId()
	if err != nil {
		sid = "0000"
	}
	d.t.TraceDependency(ctx, sid, "sql", d.name, "QueryRow "+query, true, now, after, fields)
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

// WithTrx Adds a tracer to the Database Object
func (i *dbImpl) WithTrx(t Trx) DB {
	i.t = t
	return i
}
