package stdlib

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type NativeDatabase interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Conn(ctx Context) (*sql.Conn, error)
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx Context, query string, args ...any) (sql.Result, error)
	Ping() error
	PingContext(ctx Context) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx Context, query string, args ...any) *sql.Row
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	Stats() sql.DBStats
}

type dbImpl struct {
	logger   *logger
	db       *sql.DB
	pingLock sync.Mutex
}

func NativeDatabaseProvider(Driver string, DSN string) NativeDatabase {
	l := getLoggerInstance()
	switch Driver {
	case "postgres":
		if db, err := sql.Open("postgres", DSN); err != nil {
			panic(err)
		} else {
			l.Info("[DATABASE]\tSucessfuly Connected to postgres database")
			ndb := &dbImpl{
				logger:   getLoggerInstance(),
				db:       db,
				pingLock: sync.Mutex{},
			}
			ndb.Ping()
			return ndb
		}
	default:
		panic("Unsupported driver")
	}
}

func (d *dbImpl) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *dbImpl) BeginTx(ctx Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx.Context, opts)
}

func (d *dbImpl) Close() error {
	return d.db.Close()
}

func (d *dbImpl) Conn(ctx Context) (*sql.Conn, error) {
	return d.db.Conn(ctx.Context)
}

func (d *dbImpl) Exec(query string, args ...any) (sql.Result, error) {
	now := time.Now()
	if res, err := d.db.Exec(query, args...); err != nil {
		d.logger.Errorf("[DATABASE]\tError executing query: %s with args: %v with error : %w ", query, args, err)
		return res, err
	} else {
		d.logger.Infof("[DATABASE]\t[%0.2f ms]\tExecuting query: %s with args: %v", float64(time.Since(now).Microseconds())/1000.0, query, args)
		return res, err
	}
}

func (d *dbImpl) ExecContext(ctx Context, query string, args ...any) (sql.Result, error) {
	now := time.Now()
	if res, err := d.db.ExecContext(ctx.Context, query, args...); err != nil {
		d.logger.Errorf("[DATABASE]\tError executing query: %s with args: %v with error : %w ", query, args, err)
		return res, err
	} else {
		d.logger.Infof("[DATABASE]\t[%0.2f ms]\tExecuting query: %s with args: %v", float64(time.Since(now).Microseconds())/1000.0, query, args)
		return res, err
	}
}

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

func (d *dbImpl) PingContext(ctx Context) error {
	return d.db.PingContext(ctx.Context)
}

func (d *dbImpl) Prepare(query string) (*sql.Stmt, error) {
	return d.db.Prepare(query)
}

func (d *dbImpl) PrepareContext(ctx Context, query string) (*sql.Stmt, error) {
	return d.db.PrepareContext(ctx.Context, query)
}

func (d *dbImpl) Query(query string, args ...any) (*sql.Rows, error) {
	now := time.Now()
	if res, err := d.db.Query(query, args...); err != nil {
		d.logger.Errorf("[DATABASE]\tError executing query: %s with args: %v with error : %w ", query, args, err)
		return res, err
	} else {
		d.logger.Infof("[DATABASE]\t[%0.2f ms]\tExecuting query: %s with args: %v", float64(time.Since(now).Microseconds())/1000.0, query, args)
		return res, err
	}
}

func (d *dbImpl) QueryContext(ctx Context, query string, args ...any) (*sql.Rows, error) {
	now := time.Now()
	if res, err := d.db.QueryContext(ctx.Context, query, args...); err != nil {
		d.logger.Errorf("[DATABASE]\tError executing query: %s with args: %v with error : %w ", query, args, err)
		return nil, err
	} else {
		d.logger.Infof("[DATABASE]\t[%0.2f ms]\tExecuting query: %s with args: %v", float64(time.Since(now).Microseconds())/1000.0, query, args)
		return res, err
	}
}

func (d *dbImpl) QueryRow(query string, args ...any) *sql.Row {
	now := time.Now()
	r := d.db.QueryRow(query, args...)
	d.logger.Infof("[DATABASE]\t[%0.2f ms]\tExecuting query: %s with args: %v", float64(time.Since(now).Microseconds())/1000.0, query, args)
	return r
}

func (d *dbImpl) QueryRowContext(ctx Context, query string, args ...any) *sql.Row {
	now := time.Now()
	r := d.db.QueryRowContext(ctx.Context, query, args...)
	d.logger.Infof("[DATABASE]\t[%0.2f ms]\tExecuting query: %s with args: %v", float64(time.Since(now).Microseconds())/1000.0, query, args)
	return r
}

func (i *dbImpl) SetConnMaxIdleTime(d time.Duration) {
	i.db.SetConnMaxIdleTime(d)
}

func (i *dbImpl) SetConnMaxLifetime(d time.Duration) {
	i.db.SetConnMaxLifetime(d)
}

func (i *dbImpl) Stats() sql.DBStats {
	return i.db.Stats()
}
